package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultAddr = ":8080"

type Config struct {
	Addr         string
	Token        string
	Root         string
	Shell        string
	HTTPPath     string
	Stateless    bool
	JSONResponse bool
}

func LoadConfig() (*Config, error) {
	root := getenvDefault("SANDBOX_ROOT", ".")
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve SANDBOX_ROOT: %w", err)
	}

	httpPath := getenvDefault("MCP_HTTP_PATH", "/mcp")
	if httpPath == "" || !strings.HasPrefix(httpPath, "/") {
		return nil, errors.New("MCP_HTTP_PATH must start with '/'")
	}

	return &Config{
		Addr:         getenvDefault("MCP_ADDR", defaultAddr),
		Token:        os.Getenv("MCP_AUTH_TOKEN"),
		Root:         filepath.Clean(absRoot),
		Shell:        getenvDefault("SANDBOX_SHELL", "bash"),
		HTTPPath:     httpPath,
		Stateless:    envBool("MCP_STATELESS", false),
		JSONResponse: envBool("MCP_JSON_RESPONSE", false),
	}, nil
}

func getenvDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	case "":
		return fallback
	default:
		return fallback
	}
}
