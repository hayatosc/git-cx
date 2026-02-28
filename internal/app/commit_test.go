package app

import (
	"context"
	"errors"
	"testing"

	"github.com/hayatosc/git-cx/internal/ai"
	"github.com/hayatosc/git-cx/internal/commit"
	"github.com/hayatosc/git-cx/internal/config"
	"github.com/hayatosc/git-cx/internal/execx"
	"github.com/hayatosc/git-cx/internal/git"
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

func TestCommitService_GenerateDetails(t *testing.T) {
	provider := &ai.MockProvider{Body: "body", Footer: "footer"}
	service := NewCommitService(
		&config.Config{Candidates: 1, Commit: config.CommitConfig{}},
		provider,
		git.NewRunnerWithExecutor(&execx.MockRunner{}),
	)

	body, footer, err := service.GenerateDetails(context.Background(), "diff", "stat", "feat", "core", "add")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body != "body" || footer != "footer" {
		t.Fatalf("unexpected result: %q %q", body, footer)
	}
	if provider.LastDetail == nil || provider.LastDetail.Subject != "add" {
		t.Fatalf("detail request not recorded: %#v", provider.LastDetail)
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
