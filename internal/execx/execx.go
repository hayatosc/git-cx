package execx

import (
	"bytes"
	"context"
	"os/exec"
)

// Result holds stdout and stderr output.
type Result struct {
	Stdout string
	Stderr string
}

// Runner executes commands and returns captured output.
type Runner interface {
	Run(ctx context.Context, name string, args ...string) (Result, error)
	RunShell(ctx context.Context, command string) (Result, error)
}

// DefaultRunner executes commands via os/exec.
type DefaultRunner struct{}

// Run executes a command with arguments.
func (DefaultRunner) Run(ctx context.Context, name string, args ...string) (Result, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return Result{Stdout: stdout.String(), Stderr: stderr.String()}, err
}

// RunShell executes a shell command with sh -c.
func (DefaultRunner) RunShell(ctx context.Context, command string) (Result, error) {
	return DefaultRunner{}.Run(ctx, "sh", "-c", command)
}
