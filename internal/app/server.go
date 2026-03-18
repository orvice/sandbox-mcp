package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultToolTimeout = 30 * time.Second
	maxCommandOutput   = 1 << 20 // 1 MiB
)

type readFileParams struct {
	Path string `json:"path" jsonschema:"Path to a file inside the sandbox root"`
}

type readFileResult struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type writeFileParams struct {
	Path       string `json:"path" jsonschema:"Path to a file inside the sandbox root"`
	Content    string `json:"content" jsonschema:"Full file content to write"`
	CreateDirs bool   `json:"createDirs,omitempty" jsonschema:"Create parent directories when true"`
}

type writeFileResult struct {
	Path    string `json:"path"`
	Bytes   int    `json:"bytes"`
	Created bool   `json:"created"`
}

type bashParams struct {
	Command        string            `json:"command" jsonschema:"Shell command to execute inside the sandbox"`
	Cwd            string            `json:"cwd,omitempty" jsonschema:"Optional working directory relative to the sandbox root"`
	TimeoutSeconds int               `json:"timeoutSeconds,omitempty" jsonschema:"Optional timeout in seconds, defaults to 30"`
	Env            map[string]string `json:"env,omitempty" jsonschema:"Optional environment variables for the command"`
}

type bashResult struct {
	Cwd      string `json:"cwd"`
	ExitCode int    `json:"exitCode"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

func NewMCPServer(cfg *Config) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "sandbox-mcp",
		Version: "0.1.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ReadFile",
		Description: "Read the full content of a file inside the sandbox root",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in readFileParams) (*mcp.CallToolResult, readFileResult, error) {
		target, err := resolveSandboxPath(cfg.Root, in.Path)
		if err != nil {
			return nil, readFileResult{}, err
		}

		data, err := os.ReadFile(target)
		if err != nil {
			return nil, readFileResult{}, fmt.Errorf("read file: %w", err)
		}

		out := readFileResult{
			Path:    target,
			Content: string(data),
		}
		return textResult(out.Content), out, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "WriteFile",
		Description: "Write full content to a file inside the sandbox root",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in writeFileParams) (*mcp.CallToolResult, writeFileResult, error) {
		target, err := resolveSandboxPath(cfg.Root, in.Path)
		if err != nil {
			return nil, writeFileResult{}, err
		}

		if in.CreateDirs {
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return nil, writeFileResult{}, fmt.Errorf("create parent dirs: %w", err)
			}
		}

		_, statErr := os.Stat(target)
		created := errors.Is(statErr, os.ErrNotExist)

		if err := os.WriteFile(target, []byte(in.Content), 0o644); err != nil {
			return nil, writeFileResult{}, fmt.Errorf("write file: %w", err)
		}

		out := writeFileResult{
			Path:    target,
			Bytes:   len(in.Content),
			Created: created,
		}
		return textResult(fmt.Sprintf("wrote %d bytes to %s", out.Bytes, out.Path)), out, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "Bash",
		Description: "Execute a shell command inside the sandbox root",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in bashParams) (*mcp.CallToolResult, bashResult, error) {
		if strings.TrimSpace(in.Command) == "" {
			return nil, bashResult{}, errors.New("command is required")
		}

		cwd := cfg.Root
		var err error
		if strings.TrimSpace(in.Cwd) != "" {
			cwd, err = resolveSandboxPath(cfg.Root, in.Cwd)
			if err != nil {
				return nil, bashResult{}, err
			}
		}

		timeout := defaultToolTimeout
		if in.TimeoutSeconds > 0 {
			timeout = time.Duration(in.TimeoutSeconds) * time.Second
		}

		runCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		cmd := exec.CommandContext(runCtx, cfg.Shell, "-lc", in.Command)
		cmd.Dir = cwd
		cmd.Env = append(os.Environ(), envMapToList(in.Env)...)

		stdout, stderr, exitCode, err := runCommand(cmd)
		if err != nil && runCtx.Err() == context.DeadlineExceeded {
			err = fmt.Errorf("command timed out after %s", timeout)
		}
		if err != nil && exitCode == 0 {
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				return nil, bashResult{}, err
			}
		}

		out := bashResult{
			Cwd:      cwd,
			ExitCode: exitCode,
			Stdout:   stdout,
			Stderr:   stderr,
		}
		resultText := fmt.Sprintf("exitCode=%d\nstdout:\n%s\nstderr:\n%s", out.ExitCode, emptyFallback(out.Stdout), emptyFallback(out.Stderr))
		if err != nil {
			resultText += "\nerror: " + err.Error()
		}
		return textResult(resultText), out, nil
	})

	return server
}

func resolveSandboxPath(root, userPath string) (string, error) {
	if strings.TrimSpace(userPath) == "" {
		return "", errors.New("path is required")
	}

	var candidate string
	if filepath.IsAbs(userPath) {
		candidate = filepath.Clean(userPath)
	} else {
		candidate = filepath.Join(root, userPath)
	}

	candidate = filepath.Clean(candidate)
	rel, err := filepath.Rel(root, candidate)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q escapes sandbox root %q", userPath, root)
	}
	return candidate, nil
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}
}

func envMapToList(values map[string]string) []string {
	if len(values) == 0 {
		return nil
	}

	out := make([]string, 0, len(values))
	for key, value := range values {
		out = append(out, key+"="+value)
	}
	return out
}

func emptyFallback(s string) string {
	if s == "" {
		return "(empty)"
	}
	return s
}
