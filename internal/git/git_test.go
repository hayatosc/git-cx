package git

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"git-cx/internal/execx"
)

type exitError struct {
	code int
}

func (e exitError) Error() string {
	return fmt.Sprintf("exit status %d", e.code)
}

func (e exitError) ExitCode() int {
	return e.code
}

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

func TestConfigGetFromFile_NotSet(t *testing.T) {
	mock := &execx.MockRunner{
		Errors: map[string]error{
			"git\x00config\x00--file\x00/tmp/cx.conf\x00--get\x00cx.provider": exitError{code: 1},
		},
	}
	runner := NewRunnerWithExecutor(mock)
	got, err := runner.ConfigGetFromFile(context.Background(), "/tmp/cx.conf", "cx.provider")
	if err != nil {
		t.Fatalf("ConfigGetFromFile error: %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty value, got %q", got)
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

func TestConfigGetAllFromFile_NotSet(t *testing.T) {
	mock := &execx.MockRunner{
		Errors: map[string]error{
			"git\x00config\x00--file\x00/tmp/cx.conf\x00--get-all\x00cx.commit.scopes": exitError{code: 1},
		},
	}
	runner := NewRunnerWithExecutor(mock)
	got, err := runner.ConfigGetAllFromFile(context.Background(), "/tmp/cx.conf", "cx.commit.scopes")
	if err != nil {
		t.Fatalf("ConfigGetAllFromFile error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil scopes, got %#v", got)
	}
}

func TestConfigListFromFile(t *testing.T) {
	mock := &execx.MockRunner{
		Results: map[string]execx.Result{
			"git\x00config\x00--file\x00/tmp/cx.conf\x00--list": {Stdout: "cx.provider=gemini\ncx.commit.scopes=core\ncx.commit.scopes=cli\n"},
		},
	}
	runner := NewRunnerWithExecutor(mock)
	got, err := runner.ConfigListFromFile(context.Background(), "/tmp/cx.conf")
	if err != nil {
		t.Fatalf("ConfigListFromFile error: %v", err)
	}
	if got["cx.provider"][0] != "gemini" {
		t.Fatalf("unexpected provider: %#v", got["cx.provider"])
	}
	if len(got["cx.commit.scopes"]) != 2 || got["cx.commit.scopes"][0] != "core" || got["cx.commit.scopes"][1] != "cli" {
		t.Fatalf("unexpected scopes: %#v", got["cx.commit.scopes"])
	}
}
