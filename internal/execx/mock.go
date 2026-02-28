package execx

import (
	"context"
	"fmt"
)

// MockRunner is a test double for executing commands.
type MockRunner struct {
	Results map[string]Result
	Errors  map[string]error
	Calls   []Call
	Strict  bool
}

// Call records a command invocation.
type Call struct {
	Name string
	Args []string
}

// Run executes a command returning canned results.
func (m *MockRunner) Run(ctx context.Context, name string, args ...string) (Result, error) {
	_ = ctx
	m.Calls = append(m.Calls, Call{Name: name, Args: append([]string{}, args...)})
	key := buildKey(name, args)
	if err, ok := m.Errors[key]; ok {
		return Result{}, err
	}
	if res, ok := m.Results[key]; ok {
		return res, nil
	}
	if m.Strict {
		return Result{}, fmt.Errorf("execx: unstubbed call: %s %q", name, args)
	}
	return Result{}, nil
}

// RunShell executes a shell command.
func (m *MockRunner) RunShell(ctx context.Context, command string) (Result, error) {
	return m.Run(ctx, "sh", "-c", command)
}

func buildKey(name string, args []string) string {
	if len(args) == 0 {
		return name
	}
	key := name
	for _, arg := range args {
		key += "\x00" + arg
	}
	return key
}
