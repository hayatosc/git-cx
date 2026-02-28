package git

import (
	"context"
	"errors"
	"testing"

	"git-cx/internal/execx"
)

func TestStagedDiff_NoChanges(t *testing.T) {
	mock := &execx.MockRunner{
		Results: map[string]execx.Result{
			"git\x00diff\x00--cached\x00--no-color": {Stdout: "\n"},
		},
	}
	runner := NewRunnerWithExecutor(mock)
	_, err := runner.StagedDiff(context.Background())
	if !errors.Is(err, ErrNoStagedChanges) {
		t.Fatalf("expected ErrNoStagedChanges, got %v", err)
	}
}

func TestConfigGetAll(t *testing.T) {
	mock := &execx.MockRunner{
		Results: map[string]execx.Result{
			"git\x00config\x00--get-all\x00cx.commit.scopes": {Stdout: "core\ncli\n"},
		},
	}
	runner := NewRunnerWithExecutor(mock)
	got := runner.ConfigGetAll(context.Background(), "cx.commit.scopes")
	if len(got) != 2 || got[0] != "core" || got[1] != "cli" {
		t.Fatalf("unexpected scopes: %#v", got)
	}
}

func TestConfigGetFromFile(t *testing.T) {
	mock := &execx.MockRunner{
		Results: map[string]execx.Result{
			"git\x00config\x00--file\x00/tmp/cx.conf\x00--get\x00cx.provider": {Stdout: "gemini\n"},
		},
	}
	runner := NewRunnerWithExecutor(mock)
	got, err := runner.ConfigGetFromFile(context.Background(), "/tmp/cx.conf", "cx.provider")
	if err != nil {
		t.Fatalf("ConfigGetFromFile error: %v", err)
	}
	if got != "gemini" {
		t.Fatalf("unexpected provider: %q", got)
	}
}

func TestConfigGetAllFromFile(t *testing.T) {
	mock := &execx.MockRunner{
		Results: map[string]execx.Result{
			"git\x00config\x00--file\x00/tmp/cx.conf\x00--get-all\x00cx.commit.scopes": {Stdout: "core\ncli\n"},
		},
	}
	runner := NewRunnerWithExecutor(mock)
	got, err := runner.ConfigGetAllFromFile(context.Background(), "/tmp/cx.conf", "cx.commit.scopes")
	if err != nil {
		t.Fatalf("ConfigGetAllFromFile error: %v", err)
	}
	if len(got) != 2 || got[0] != "core" || got[1] != "cli" {
		t.Fatalf("unexpected scopes: %#v", got)
	}
}

func TestConfigListFromFile(t *testing.T) {
	mock := &execx.MockRunner{
		Results: map[string]execx.Result{
			"git\x00config\x00--file\x00/tmp/cx.conf\x00--list": {Stdout: "cx.provider=gemini\n"},
		},
	}
	runner := NewRunnerWithExecutor(mock)
	got, err := runner.ConfigListFromFile(context.Background(), "/tmp/cx.conf")
	if err != nil {
		t.Fatalf("ConfigListFromFile error: %v", err)
	}
	if got != "cx.provider=gemini" {
		t.Fatalf("unexpected output: %q", got)
	}
}
