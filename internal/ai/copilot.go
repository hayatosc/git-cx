package ai

import (
	"context"

	"git-cx/internal/config"
	"git-cx/internal/execx"
)

// CopilotProvider calls GitHub Copilot CLI to generate commit messages.
type CopilotProvider struct {
	model      string
	candidates int
	timeout    int
	runner     execx.Runner
}

// NewCopilotProvider creates a CopilotProvider from config.
func NewCopilotProvider(cfg *config.Config, runner execx.Runner) *CopilotProvider {
	return &CopilotProvider{
		model:      cfg.Model,
		candidates: cfg.Candidates,
		timeout:    cfg.Timeout,
		runner:     runner,
	}
}

func (p *CopilotProvider) Name() string { return "copilot" }

func (p *CopilotProvider) Generate(ctx context.Context, req GenerateRequest) ([]string, error) {
	prompt := buildPrompt(req)
	args := []string{"-p", prompt}
	if p.model != "" {
		args = append(args, "--model", p.model)
	}
	return runCLI(ctx, p.runner, "copilot", args, p.timeout, p.candidates)
}
