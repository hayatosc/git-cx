package git

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hayatosc/git-cx/internal/execx"
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

func TestUnstagedDiff(t *testing.T) {
	mock := &execx.MockRunner{
		Results: map[string]execx.Result{
			"git\x00diff\x00--no-color": {Stdout: "diff --git a/foo.go b/foo.go\n"},
		},
	}
	runner := NewRunnerWithExecutor(mock)
	got, err := runner.UnstagedDiff(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "diff --git a/foo.go b/foo.go" {
		t.Fatalf("unexpected diff: %q", got)
	}
}

func TestUnstagedDiff_NoChanges(t *testing.T) {
	mock := &execx.MockRunner{
		Results: map[string]execx.Result{
			"git\x00diff\x00--no-color": {Stdout: "\n"},
		},
	}
	runner := NewRunnerWithExecutor(mock)
	got, err := runner.UnstagedDiff(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestUnstagedStat(t *testing.T) {
	mock := &execx.MockRunner{
		Results: map[string]execx.Result{
			"git\x00diff\x00--stat\x00--no-color": {Stdout: " foo.go | 1 +\n 1 file changed\n"},
		},
	}
	runner := NewRunnerWithExecutor(mock)
	got, err := runner.UnstagedStat(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Fatal("expected non-empty stat")
	}
}

func TestLastCommitDiff(t *testing.T) {
	mock := &execx.MockRunner{
		Results: map[string]execx.Result{
			"git\x00show\x00HEAD\x00--no-color": {Stdout: "commit abc\ndiff --git a/foo.go b/foo.go\n"},
		},
	}
	runner := NewRunnerWithExecutor(mock)
	got, err := runner.LastCommitDiff(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Fatal("expected non-empty diff")
	}
}

func TestLastCommitStat(t *testing.T) {
	mock := &execx.MockRunner{
		Results: map[string]execx.Result{
			"git\x00show\x00HEAD\x00--stat\x00--no-color\x00--format=": {Stdout: " foo.go | 2 +-\n 1 file changed\n"},
		},
	}
	runner := NewRunnerWithExecutor(mock)
	got, err := runner.LastCommitStat(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Fatal("expected non-empty stat")
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

func TestCommit_success_returnsCombinedOutput(t *testing.T) {
	runner := NewRunnerWithExecutor(stubRunner{
		result: execx.Result{Stdout: "created", Stderr: "hook log"},
	})
	out, err := runner.Commit(context.Background(), "feat: ok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "created\nhook log" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestCommit_error_includesOutputAndMessage(t *testing.T) {
	runner := NewRunnerWithExecutor(stubRunner{
		result: execx.Result{Stderr: "lint failed\n"},
		err:    errors.New("exit status 1"),
	})
	out, err := runner.Commit(context.Background(), "feat: fail")
	if out != "lint failed" {
		t.Fatalf("unexpected output: %q", out)
	}
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "git commit:") {
		t.Fatalf("expected git commit prefix, got: %v", err)
	}
	if !strings.Contains(err.Error(), "lint failed") {
		t.Fatalf("expected stderr content, got: %v", err)
	}
}

type stubRunner struct {
	result execx.Result
	err    error
}

func (s stubRunner) Run(ctx context.Context, name string, args ...string) (execx.Result, error) {
	return s.result, s.err
}

func (s stubRunner) RunShell(ctx context.Context, command string) (execx.Result, error) {
	return s.Run(ctx, "sh", "-c", command)
}
