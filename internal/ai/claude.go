package ai

import (
	"context"

	"github.com/hayatosc/git-cx/internal/config"
	"github.com/hayatosc/git-cx/internal/execx"
)

// ClaudeProvider calls the Claude CLI to generate commit messages.
type ClaudeProvider struct{ cliProvider }

// NewClaudeProvider creates a ClaudeProvider from config.
func NewClaudeProvider(cfg *config.Config, runner execx.Runner) *ClaudeProvider {
	return &ClaudeProvider{cliProvider{
		cfg:        cliArgs{name: "claude", promptFlag: "-p", modelFlag: "--model"},
		model:      cfg.Model,
		candidates: cfg.Candidates,
		timeout:    cfg.Timeout,
		runner:     runner,
	}}
}

func (p *ClaudeProvider) Name() string { return "claude" }

func (p *ClaudeProvider) Generate(ctx context.Context, req GenerateRequest) ([]string, error) {
	return p.generate(ctx, req)
}

func (p *ClaudeProvider) GenerateDetail(ctx context.Context, req GenerateRequest) (string, string, error) {
	return p.generateDetail(ctx, req)
}
