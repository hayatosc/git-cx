package ai

import (
	"context"
	"strings"

	"git-cx/internal/config"
	"git-cx/internal/execx"
)

// CustomProvider runs an arbitrary shell command with a {prompt} placeholder.
type CustomProvider struct {
	command    string
	candidates int
	timeout    int
	runner     execx.Runner
}

// NewCustomProvider creates a CustomProvider from config.
func NewCustomProvider(cfg *config.Config, runner execx.Runner) *CustomProvider {
	return &CustomProvider{
		command:    cfg.Command,
		candidates: cfg.Candidates,
		timeout:    cfg.Timeout,
		runner:     runner,
	}
}

func (p *CustomProvider) Name() string { return "custom" }

func (p *CustomProvider) Generate(ctx context.Context, req GenerateRequest) ([]string, error) {
	prompt := buildPrompt(req)
	cmdStr := strings.ReplaceAll(p.command, "{prompt}", prompt)
	return runShell(ctx, p.runner, cmdStr, p.timeout, p.candidates)
}

func (p *CustomProvider) GenerateDetail(ctx context.Context, req GenerateRequest) (string, string, error) {
	prompt := buildDetailPrompt(req)
	cmdStr := strings.ReplaceAll(p.command, "{prompt}", prompt)
	output, err := runShellOutput(ctx, p.runner, cmdStr, p.timeout)
	if err != nil {
		return "", "", err
	}
	body, footer := parseDetailOutput(output)
	return body, footer, nil
}
