package ai

import (
	"context"

	"git-cx/internal/config"
	"git-cx/internal/execx"
)

// GeminiProvider calls the Gemini CLI to generate commit messages.
type GeminiProvider struct {
	model      string
	candidates int
	timeout    int
	runner     execx.Runner
}

// NewGeminiProvider creates a GeminiProvider from config.
func NewGeminiProvider(cfg *config.Config, runner execx.Runner) *GeminiProvider {
	return &GeminiProvider{
		model:      cfg.Model,
		candidates: cfg.Candidates,
		timeout:    cfg.Timeout,
		runner:     runner,
	}
}

func (p *GeminiProvider) Name() string { return "gemini" }

func (p *GeminiProvider) Generate(ctx context.Context, req GenerateRequest) ([]string, error) {
	prompt := buildPrompt(req)
	args := []string{"-p", prompt}
	if p.model != "" {
		args = append(args, "-m", p.model)
	}
	return runCLI(ctx, p.runner, "gemini", args, p.timeout, p.candidates)
}
