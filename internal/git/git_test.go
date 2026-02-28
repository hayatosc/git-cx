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
