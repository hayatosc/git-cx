package ai

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"

	"git-cx/internal/config"
	"git-cx/internal/execx"
)

// Provider is the interface for AI commit message generators.
type Provider interface {
	Generate(ctx context.Context, req GenerateRequest) ([]string, error)
	GenerateDetail(ctx context.Context, req GenerateRequest) (string, string, error)
	Name() string
}

// GenerateRequest holds the input for a generate call.
type GenerateRequest struct {
	Diff       string
	Stat       string // git diff --cached --stat の出力
	CommitType string
	Scope      string
	Subject    string
	Candidates int
}

// NewProvider returns the appropriate Provider based on config.
func NewProvider(cfg *config.Config) (Provider, error) {
	switch cfg.Provider {
	case "gemini":
		return NewGeminiProvider(cfg, execx.DefaultRunner{}), nil
	case "copilot":
		return NewCopilotProvider(cfg, execx.DefaultRunner{}), nil
	case "custom":
		return NewCustomProvider(cfg, execx.DefaultRunner{}), nil
	default:
		return nil, fmt.Errorf("unknown provider: %q (set cx.provider to gemini, copilot, or custom)", cfg.Provider)
	}
}

// runCLI executes name with args, returning parsed candidate lines.
func runCLI(ctx context.Context, runner execx.Runner, name string, args []string, timeout, max int) ([]string, error) {
	output, err := runCLIOutput(ctx, runner, name, args, timeout)
	if err != nil {
		return nil, err
	}
	return parseOutput(output, max), nil
}

// runShell executes cmdStr via sh -c, returning parsed candidate lines.
func runShell(ctx context.Context, runner execx.Runner, cmdStr string, timeout, max int) ([]string, error) {
	output, err := runShellOutput(ctx, runner, cmdStr, timeout)
	if err != nil {
		return nil, err
	}
	return parseOutput(output, max), nil
}

// runCLIOutput executes name with args, returning raw stdout.
func runCLIOutput(ctx context.Context, runner execx.Runner, name string, args []string, timeout int) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	result, err := runner.Run(ctx, name, args...)
	if err != nil {
		msg := strings.TrimSpace(result.Stderr)
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%s failed: %s", name, msg)
	}
	return result.Stdout, nil
}

// runShellOutput executes cmdStr via sh -c, returning raw stdout.
func runShellOutput(ctx context.Context, runner execx.Runner, cmdStr string, timeout int) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	result, err := runner.RunShell(ctx, cmdStr)
	if err != nil {
		msg := strings.TrimSpace(result.Stderr)
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("command failed: %s", msg)
	}
	return result.Stdout, nil
}

// parseOutput extracts non-empty lines from output up to max count.
func parseOutput(output string, max int) []string {
	scanner := bufio.NewScanner(strings.NewReader(output))
	var results []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		results = append(results, line)
		if max > 0 && len(results) >= max {
			break
		}
	}
	return results
}
