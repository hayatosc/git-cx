package ai

import (
	"context"

	"git-cx/internal/config"
	"git-cx/internal/execx"
)

// CodexProvider calls the Codex CLI to generate commit messages.
type CodexProvider struct {
	model      string
	candidates int
	timeout    int
	runner     execx.Runner
}

// NewCodexProvider creates a CodexProvider from config.
func NewCodexProvider(cfg *config.Config, runner execx.Runner) *CodexProvider {
	return &CodexProvider{
		model:      cfg.Model,
		candidates: cfg.Candidates,
		timeout:    cfg.Timeout,
		runner:     runner,
	}
}

func (p *CodexProvider) Name() string { return "codex" }

func (p *CodexProvider) Generate(ctx context.Context, req GenerateRequest) ([]string, error) {
	prompt := buildPrompt(req)
	args := []string{"exec", prompt}
	if p.model != "" {
		args = append(args, "--model", p.model)
	}
	return runCLI(ctx, p.runner, "codex", args, p.timeout, p.candidates)
}
