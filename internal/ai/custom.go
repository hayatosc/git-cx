package ai

import (
	"context"
	"strings"

	"git-cx/internal/config"
)

// CustomProvider runs an arbitrary shell command with a {prompt} placeholder.
type CustomProvider struct {
	command    string
	candidates int
	timeout    int
}

// NewCustomProvider creates a CustomProvider from config.
func NewCustomProvider(cfg *config.Config) *CustomProvider {
	return &CustomProvider{
		command:    cfg.Command,
		candidates: cfg.Candidates,
		timeout:    cfg.Timeout,
	}
}

func (p *CustomProvider) Name() string { return "custom" }

func (p *CustomProvider) Generate(ctx context.Context, req GenerateRequest) ([]string, error) {
	prompt := buildPrompt(req)
	cmdStr := strings.ReplaceAll(p.command, "{prompt}", prompt)
	return runShell(ctx, cmdStr, p.timeout, p.candidates)
}
