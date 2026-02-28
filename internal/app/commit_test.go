package app

import (
	"context"
	"errors"
	"testing"

	"git-cx/internal/ai"
	"git-cx/internal/commit"
	"git-cx/internal/config"
	"git-cx/internal/execx"
	"git-cx/internal/git"
)

func TestCommitService_GenerateCandidates(t *testing.T) {
	provider := &ai.MockProvider{Candidates: []string{"feat: ok"}}
	service := NewCommitService(
		&config.Config{Candidates: 1, Commit: config.CommitConfig{}},
		provider,
		git.NewRunnerWithExecutor(&execx.MockRunner{}),
	)

	got, err := service.GenerateCandidates(context.Background(), "diff", "stat", "feat", "core")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != "feat: ok" {
		t.Fatalf("unexpected candidates: %#v", got)
	}
	if provider.LastReq == nil || provider.LastReq.CommitType != "feat" || provider.LastReq.Scope != "core" {
		t.Fatalf("request not recorded: %#v", provider.LastReq)
	}
}

func TestCommitService_CommitEmptyMessage(t *testing.T) {
	service := NewCommitService(
		&config.Config{Candidates: 1, Commit: config.CommitConfig{}},
		&ai.MockProvider{},
		git.NewRunnerWithExecutor(&execx.MockRunner{}),
	)

	err := service.Commit(context.Background(), " ")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestCommitService_BuildMessage(t *testing.T) {
	service := NewCommitService(
		&config.Config{Candidates: 1, Commit: config.CommitConfig{UseEmoji: false, MaxSubjectLength: 72}},
		&ai.MockProvider{},
		git.NewRunnerWithExecutor(&execx.MockRunner{}),
	)

	msg := service.BuildMessage(&commit.ConventionalCommit{Type: "chore", Subject: "update"})
	if msg != "chore: update" {
		t.Fatalf("unexpected message: %s", msg)
	}
}

func TestCommitService_CommitPropagatesError(t *testing.T) {
	mock := &execx.MockRunner{
		Errors: map[string]error{
			"git\x00commit\x00-m\x00msg": errors.New("fail"),
		},
	}
	service := NewCommitService(
		&config.Config{Candidates: 1, Commit: config.CommitConfig{}},
		&ai.MockProvider{},
		git.NewRunnerWithExecutor(mock),
	)
	err := service.Commit(context.Background(), "msg")
	if err == nil {
		t.Fatalf("expected error")
	}
}
