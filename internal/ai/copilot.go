package ai

import (
	"context"

	"git-cx/internal/config"
)

// CopilotProvider calls GitHub Copilot CLI to generate commit messages.
type CopilotProvider struct {
	model      string
	candidates int
	timeout    int
}

// NewCopilotProvider creates a CopilotProvider from config.
func NewCopilotProvider(cfg *config.Config) *CopilotProvider {
	return &CopilotProvider{
		model:      cfg.Model,
		candidates: cfg.Candidates,
		timeout:    cfg.Timeout,
	}
}

func (p *CopilotProvider) Name() string { return "copilot" }

func (p *CopilotProvider) Generate(ctx context.Context, req GenerateRequest) ([]string, error) {
	prompt := buildPrompt(req)
	args := []string{"-p", prompt}
	if p.model != "" {
		args = append(args, "--model", p.model)
	}
	return runCLI(ctx, "copilot", args, p.timeout, p.candidates)
}
