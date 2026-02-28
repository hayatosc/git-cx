package ai

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"git-cx/internal/config"
)

// Provider is the interface for AI commit message generators.
type Provider interface {
	Generate(ctx context.Context, req GenerateRequest) ([]string, error)
	Name() string
}

// GenerateRequest holds the input for a generate call.
type GenerateRequest struct {
	Diff       string
	CommitType string
	Scope      string
	Candidates int
}

// NewProvider returns the appropriate Provider based on config.
func NewProvider(cfg *config.Config) (Provider, error) {
	switch cfg.Provider {
	case "gemini":
		return NewGeminiProvider(cfg), nil
	case "copilot":
		return NewCopilotProvider(cfg), nil
	case "custom":
		if cfg.Command == "" {
			return nil, fmt.Errorf("cx.command is not set (required for custom provider)")
		}
		return NewCustomProvider(cfg), nil
	default:
		return nil, fmt.Errorf("unknown provider: %q (set cx.provider to gemini, copilot, or custom)", cfg.Provider)
	}
}

// runCLI executes name with args, returning parsed candidate lines.
func runCLI(ctx context.Context, name string, args []string, timeout, max int) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("%s failed: %s", name, msg)
	}
	return parseOutput(stdout.String(), max), nil
}

// runShell executes cmdStr via sh -c, returning parsed candidate lines.
func runShell(ctx context.Context, cmdStr string, timeout, max int) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("command failed: %s", msg)
	}
	return parseOutput(stdout.String(), max), nil
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
