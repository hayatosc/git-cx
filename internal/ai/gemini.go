package ai

import (
	"context"

	"github.com/hayatosc/git-cx/internal/config"
	"github.com/hayatosc/git-cx/internal/execx"
)

// GeminiProvider calls the Gemini CLI to generate commit messages.
type GeminiProvider struct{ cliProvider }

// NewGeminiProvider creates a GeminiProvider from config.
func NewGeminiProvider(pc config.ProviderConfig, runner execx.Runner) *GeminiProvider {
	return &GeminiProvider{cliProvider{
		cfg:        cliArgs{name: "gemini", promptFlag: "-p", modelFlag: "-m"},
		model:      pc.Model,
		candidates: pc.Candidates,
		timeout:    pc.Timeout,
		runner:     runner,
	}}
}

func (p *GeminiProvider) Name() string { return "gemini" }

func (p *GeminiProvider) Generate(ctx context.Context, req GenerateRequest) ([]string, error) {
	return p.generate(ctx, req)
}

func (p *GeminiProvider) GenerateDetail(ctx context.Context, req GenerateRequest) (string, string, error) {
	return p.generateDetail(ctx, req)
}
