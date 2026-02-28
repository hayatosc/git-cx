package ai

import (
	"context"

	"github.com/hayatosc/git-cx/internal/config"
	"github.com/hayatosc/git-cx/internal/execx"
)

// ClaudeProvider calls the Claude CLI to generate commit messages.
type ClaudeProvider struct {
	model      string
	candidates int
	timeout    int
	runner     execx.Runner
}

// NewClaudeProvider creates a ClaudeProvider from config.
func NewClaudeProvider(cfg *config.Config, runner execx.Runner) *ClaudeProvider {
	return &ClaudeProvider{
		model:      cfg.Model,
		candidates: cfg.Candidates,
		timeout:    cfg.Timeout,
		runner:     runner,
	}
}

func (p *ClaudeProvider) Name() string { return "claude" }

func (p *ClaudeProvider) Generate(ctx context.Context, req GenerateRequest) ([]string, error) {
	prompt := buildPrompt(req)
	args := []string{"-p", prompt}
	if p.model != "" {
		args = append(args, "--model", p.model)
	}
	return runCLI(ctx, p.runner, "claude", args, p.timeout, p.candidates)
}

func (p *ClaudeProvider) GenerateDetail(ctx context.Context, req GenerateRequest) (string, string, error) {
	prompt := buildDetailPrompt(req)
	args := []string{"-p", prompt}
	if p.model != "" {
		args = append(args, "--model", p.model)
	}
	output, err := runCLIOutput(ctx, p.runner, "claude", args, p.timeout)
	if err != nil {
		return "", "", err
	}
	body, footer := parseDetailOutput(output)
	return body, footer, nil
}
