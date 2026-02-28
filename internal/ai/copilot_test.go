package ai

import (
	"context"
	"testing"

	"github.com/hayatosc/git-cx/internal/config"
	"github.com/hayatosc/git-cx/internal/execx"
)

func TestCopilotProviderUsesCLI(t *testing.T) {
	runner := &execx.MockRunner{Strict: true}
	prompt := buildPrompt(GenerateRequest{Diff: "diff", Candidates: 1})
	key := "copilot\x00-p\x00" + prompt + "\x00--model\x00gpt-4o"
	runner.Results = map[string]execx.Result{key: {Stdout: "feat: ok"}}

	cfg := &config.Config{Model: "gpt-4o", Candidates: 1, Timeout: 1}
	provider := NewCopilotProvider(cfg, runner)
	got, err := provider.Generate(context.Background(), GenerateRequest{Diff: "diff", Candidates: 1})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if len(got) != 1 || got[0] != "feat: ok" {
		t.Fatalf("unexpected candidates: %#v", got)
	}
}
