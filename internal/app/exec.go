package app

import (
	"errors"
	"os/exec"
)

func runCommand(cmd *exec.Cmd) (stdout string, stderr string, exitCode int, err error) {
	var outBuf, errBuf cappedBuffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()

	if err == nil {
		return stdout, stderr, 0, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return stdout, stderr, exitErr.ExitCode(), err
	}

	return stdout, stderr, 0, err
}

type cappedBuffer struct {
	data []byte
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	if len(b.data) < maxCommandOutput {
		remaining := maxCommandOutput - len(b.data)
		if remaining > len(p) {
			remaining = len(p)
		}
		b.data = append(b.data, p[:remaining]...)
	}
	return len(p), nil
}

func (b *cappedBuffer) String() string {
	if len(b.data) < maxCommandOutput {
		return string(b.data)
	}
	return string(b.data) + "\n... output truncated ..."
}
