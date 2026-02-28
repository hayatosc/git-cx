package ai

import (
	"context"
	"testing"

	"git-cx/internal/config"
	"git-cx/internal/execx"
)

func TestGeminiProviderUsesCLI(t *testing.T) {
	runner := &execx.MockRunner{Strict: true}
	prompt := buildPrompt(GenerateRequest{Diff: "diff", Candidates: 1})
	key := "gemini\x00-p\x00" + prompt + "\x00-m\x00gemini-model"
	runner.Results = map[string]execx.Result{key: {Stdout: "feat: ok"}}

	cfg := &config.Config{Model: "gemini-model", Candidates: 1, Timeout: 1}
	provider := NewGeminiProvider(cfg, runner)
	got, err := provider.Generate(context.Background(), GenerateRequest{Diff: "diff", Candidates: 1})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if len(got) != 1 || got[0] != "feat: ok" {
		t.Fatalf("unexpected candidates: %#v", got)
	}
}
