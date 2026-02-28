package ai

import (
	"context"
	"testing"

	"git-cx/internal/config"
	"git-cx/internal/execx"
)

func TestCodexProviderUsesCLI(t *testing.T) {
	runner := &execx.MockRunner{Strict: true}
	prompt := buildPrompt(GenerateRequest{Diff: "diff", Candidates: 1})
	key := "codex\x00exec\x00" + prompt + "\x00--model\x00gpt-5"
	runner.Results = map[string]execx.Result{key: {Stdout: "feat: ok"}}

	cfg := &config.Config{Model: "gpt-5", Candidates: 1, Timeout: 1}
	provider := NewCodexProvider(cfg, runner)
	got, err := provider.Generate(context.Background(), GenerateRequest{Diff: "diff", Candidates: 1})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if len(got) != 1 || got[0] != "feat: ok" {
		t.Fatalf("unexpected candidates: %#v", got)
	}
}
