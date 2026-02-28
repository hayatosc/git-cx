package config

import (
	"context"
	"testing"

	"git-cx/internal/execx"
	"git-cx/internal/git"
)

func TestLoadWithFile_GitConfigFormatOverrides(t *testing.T) {
	mock := &execx.MockRunner{
		Results: map[string]execx.Result{
			"git\x00config\x00--file\x00/tmp/cx.conf\x00--get\x00cx.provider":                {Stdout: "copilot\n"},
			"git\x00config\x00--file\x00/tmp/cx.conf\x00--get\x00cx.model":                   {Stdout: "gpt-4o\n"},
			"git\x00config\x00--file\x00/tmp/cx.conf\x00--get\x00cx.candidates":              {Stdout: "5\n"},
			"git\x00config\x00--file\x00/tmp/cx.conf\x00--get\x00cx.timeout":                 {Stdout: "45\n"},
			"git\x00config\x00--file\x00/tmp/cx.conf\x00--get\x00cx.commit.useEmoji":         {Stdout: "true\n"},
			"git\x00config\x00--file\x00/tmp/cx.conf\x00--get\x00cx.commit.maxSubjectLength": {Stdout: "80\n"},
			"git\x00config\x00--file\x00/tmp/cx.conf\x00--get-all\x00cx.commit.scopes":       {Stdout: "core\ncli\n"},
		},
	}

	cfg, err := LoadWithFile(context.Background(), git.NewRunnerWithExecutor(mock), "/tmp/cx.conf")
	if err != nil {
		t.Fatalf("LoadWithFile error: %v", err)
	}

	if cfg.Provider != "copilot" || cfg.Model != "gpt-4o" {
		t.Fatalf("unexpected provider/model: %s/%s", cfg.Provider, cfg.Model)
	}
	if cfg.Candidates != 5 || cfg.Timeout != 45 {
		t.Fatalf("unexpected candidates/timeout: %d/%d", cfg.Candidates, cfg.Timeout)
	}
	if !cfg.Commit.UseEmoji || cfg.Commit.MaxSubjectLength != 80 {
		t.Fatalf("unexpected commit config: %+v", cfg.Commit)
	}
	if len(cfg.Commit.Scopes) != 2 || cfg.Commit.Scopes[0] != "core" || cfg.Commit.Scopes[1] != "cli" {
		t.Fatalf("unexpected scopes: %#v", cfg.Commit.Scopes)
	}
}
