package config

import (
	"context"
	"errors"
	"testing"

	"github.com/hayatosc/git-cx/internal/execx"
	"github.com/hayatosc/git-cx/internal/git"
)

func TestLoadWithFile_GitConfigFormatOverrides(t *testing.T) {
	mock := &execx.MockRunner{
		Results: map[string]execx.Result{
			"git\x00config\x00--file\x00/tmp/cx.conf\x00--list": {Stdout: "cx.provider=copilot\ncx.model=gpt-4o\ncx.candidates=5\ncx.timeout=45\ncx.commit.useEmoji=yes\ncx.commit.maxSubjectLength=80\ncx.commit.scopes=core\ncx.commit.scopes=cli\n"},
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

func TestLoadWithFile_InvalidIntValue(t *testing.T) {
	mock := &execx.MockRunner{
		Results: map[string]execx.Result{
			"git\x00config\x00--file\x00/tmp/cx.conf\x00--list": {Stdout: "cx.candidates=abc\n"},
		},
	}

	_, err := LoadWithFile(context.Background(), git.NewRunnerWithExecutor(mock), "/tmp/cx.conf")
	if err == nil {
		t.Fatalf("expected error for invalid candidates")
	}
}

func TestLoadWithFile_InvalidBoolValue(t *testing.T) {
	mock := &execx.MockRunner{
		Results: map[string]execx.Result{
			"git\x00config\x00--file\x00/tmp/cx.conf\x00--list": {Stdout: "cx.commit.useEmoji=maybe\n"},
		},
	}

	_, err := LoadWithFile(context.Background(), git.NewRunnerWithExecutor(mock), "/tmp/cx.conf")
	if err == nil {
		t.Fatalf("expected error for invalid useEmoji")
	}
}

func TestLoadWithFile_GitConfigFormatMissingFile(t *testing.T) {
	mock := &execx.MockRunner{
		Errors: map[string]error{
			"git\x00config\x00--file\x00/tmp/missing.conf\x00--list": errors.New("exit status 128"),
		},
	}

	_, err := LoadWithFile(context.Background(), git.NewRunnerWithExecutor(mock), "/tmp/missing.conf")
	if err == nil {
		t.Fatalf("expected error for missing config file")
	}
}
