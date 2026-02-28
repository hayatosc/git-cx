package ai

import (
	"context"
	"testing"

	"github.com/hayatosc/git-cx/internal/config"
	"github.com/hayatosc/git-cx/internal/execx"
)

func TestClaudeProviderUsesCLI(t *testing.T) {
	runner := &execx.MockRunner{Strict: true}
	prompt := buildPrompt(GenerateRequest{Diff: "diff", Candidates: 1})
	key := "claude\x00-p\x00" + prompt + "\x00--model\x00claude-model"
	runner.Results = map[string]execx.Result{key: {Stdout: "feat: ok"}}

	cfg := &config.Config{Model: "claude-model", Candidates: 1, Timeout: 1}
	provider := NewClaudeProvider(cfg, runner)
	got, err := provider.Generate(context.Background(), GenerateRequest{Diff: "diff", Candidates: 1})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if len(got) != 1 || got[0] != "feat: ok" {
		t.Fatalf("unexpected candidates: %#v", got)
	}
}

func TestClaudeProviderGenerateDetailUsesCLI(t *testing.T) {
	runner := &execx.MockRunner{Strict: true}
	prompt := buildDetailPrompt(GenerateRequest{Diff: "diff"})
	key := "claude\x00-p\x00" + prompt + "\x00--model\x00claude-model"
	runner.Results = map[string]execx.Result{key: {Stdout: "Body:\nbody\nFooter:\nfooter"}}

	cfg := &config.Config{Model: "claude-model", Candidates: 1, Timeout: 1}
	provider := NewClaudeProvider(cfg, runner)
	body, footer, err := provider.GenerateDetail(context.Background(), GenerateRequest{Diff: "diff"})
	if err != nil {
		t.Fatalf("GenerateDetail returned error: %v", err)
	}
	if body != "body" || footer != "footer" {
		t.Fatalf("unexpected details: %q %q", body, footer)
	}
}
