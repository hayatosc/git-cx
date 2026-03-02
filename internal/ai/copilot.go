package ai

import (
	"context"

	"github.com/hayatosc/git-cx/internal/config"
	"github.com/hayatosc/git-cx/internal/execx"
)

// CopilotProvider calls GitHub Copilot CLI to generate commit messages.
type CopilotProvider struct{ cliProvider }

// NewCopilotProvider creates a CopilotProvider from config.
func NewCopilotProvider(cfg *config.Config, runner execx.Runner) *CopilotProvider {
	return &CopilotProvider{cliProvider{
		cfg:        cliArgs{name: "copilot", promptFlag: "-p", modelFlag: "--model"},
		model:      cfg.Model,
		candidates: cfg.Candidates,
		timeout:    cfg.Timeout,
		runner:     runner,
	}}
}

func (p *CopilotProvider) Name() string { return "copilot" }

func (p *CopilotProvider) Generate(ctx context.Context, req GenerateRequest) ([]string, error) {
	return p.generate(ctx, req)
}

func (p *CopilotProvider) GenerateDetail(ctx context.Context, req GenerateRequest) (string, string, error) {
	return p.generateDetail(ctx, req)
}
