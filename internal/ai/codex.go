package ai

import (
	"context"

	"github.com/hayatosc/git-cx/internal/config"
	"github.com/hayatosc/git-cx/internal/execx"
)

// CodexProvider calls the Codex CLI to generate commit messages.
type CodexProvider struct{ cliProvider }

// NewCodexProvider creates a CodexProvider from config.
func NewCodexProvider(cfg *config.Config, runner execx.Runner) *CodexProvider {
	return &CodexProvider{cliProvider{
		cfg:        cliArgs{name: "codex", promptFlag: "exec", modelFlag: "--model"},
		model:      cfg.Model,
		candidates: cfg.Candidates,
		timeout:    cfg.Timeout,
		runner:     runner,
	}}
}

func (p *CodexProvider) Name() string { return "codex" }

func (p *CodexProvider) Generate(ctx context.Context, req GenerateRequest) ([]string, error) {
	return p.generate(ctx, req)
}

func (p *CodexProvider) GenerateDetail(ctx context.Context, req GenerateRequest) (string, string, error) {
	return p.generateDetail(ctx, req)
}
