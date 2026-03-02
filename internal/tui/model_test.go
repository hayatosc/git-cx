package tui

import (
	"strings"
	"testing"

	"github.com/hayatosc/git-cx/internal/ai"
	"github.com/hayatosc/git-cx/internal/app"
	"github.com/hayatosc/git-cx/internal/config"
	"github.com/hayatosc/git-cx/internal/execx"
	"github.com/hayatosc/git-cx/internal/git"
)

func newTestService(mock *execx.MockRunner) *app.CommitService {
	return app.NewCommitService(
		&config.Config{Candidates: 1, Commit: config.CommitConfig{}},
		&ai.MockProvider{Candidates: []string{"feat: test"}},
		git.NewRunnerWithExecutor(mock),
	)
}

// TestDryRun_doCommit_skipsCommit checks that doCommit does not call git commit when dryRun=true.
func TestDryRun_doCommit_skipsCommit(t *testing.T) {
	mock := &execx.MockRunner{Strict: true}
	service := newTestService(mock)

	m := New(service, "diff content", "stat content", true)
	m.commitType = "feat"
	m.subject = "add dry-run support"

	cmd := m.doCommit()
	msg := cmd()

	done, ok := msg.(commitDoneMsg)
	if !ok {
		t.Fatalf("expected commitDoneMsg, got %T", msg)
	}
	if done.err != nil {
		t.Fatalf("unexpected error: %v", done.err)
	}
	if done.message == "" {
		t.Fatal("expected non-empty message in dry-run mode")
	}

	for _, call := range mock.Calls {
		if call.Name == "git" && len(call.Args) > 0 && call.Args[0] == "commit" {
			t.Fatal("git commit was called in dry-run mode")
		}
	}
}

// TestDryRun_doCommit_includesMessage checks that the message in commitDoneMsg matches BuildMessage.
func TestDryRun_doCommit_includesMessage(t *testing.T) {
	mock := &execx.MockRunner{}
	service := newTestService(mock)

	m := New(service, "diff", "stat", true)
	m.commitType = "chore"
	m.subject = "update deps"

	cmd := m.doCommit()
	msg := cmd()

	done := msg.(commitDoneMsg)
	if done.message != "chore: update deps" {
		t.Fatalf("unexpected message: %q", done.message)
	}
}

// TestNoDryRun_doCommit_callsCommit checks that doCommit calls git commit when dryRun=false.
func TestNoDryRun_doCommit_callsCommit(t *testing.T) {
	mock := &execx.MockRunner{
		Results: map[string]execx.Result{
			"git\x00commit\x00-m\x00feat: normal commit": {},
		},
	}
	service := newTestService(mock)

	m := New(service, "diff", "stat", false)
	m.commitType = "feat"
	m.subject = "normal commit"

	cmd := m.doCommit()
	msg := cmd()

	done, ok := msg.(commitDoneMsg)
	if !ok {
		t.Fatalf("expected commitDoneMsg, got %T", msg)
	}
	if done.err != nil {
		t.Fatalf("unexpected error: %v", done.err)
	}

	committed := false
	for _, call := range mock.Calls {
		if call.Name == "git" && len(call.Args) > 0 && call.Args[0] == "commit" {
			committed = true
			break
		}
	}
	if !committed {
		t.Fatal("expected git commit to be called")
	}
}

// TestDryRun_View_doneState checks that View shows [DRY RUN] message after confirmation.
func TestDryRun_View_doneState(t *testing.T) {
	mock := &execx.MockRunner{}
	service := newTestService(mock)

	m := New(service, "diff", "stat", true)
	m.commitType = "feat"
	m.subject = "add feature"
	m.state = stateDone
	m.quitting = true
	m.dryRunMsg = "feat: add feature"

	view := m.View()
	if !strings.Contains(view, "[DRY RUN]") {
		t.Errorf("View() should contain '[DRY RUN]', got: %q", view)
	}
	if !strings.Contains(view, "feat: add feature") {
		t.Errorf("View() should contain commit message, got: %q", view)
	}
}

// TestNoDryRun_View_doneState checks that View shows success message when dryRun=false.
func TestNoDryRun_View_doneState(t *testing.T) {
	mock := &execx.MockRunner{}
	service := newTestService(mock)

	m := New(service, "diff", "stat", false)
	m.state = stateDone
	m.quitting = true

	view := m.View()
	if strings.Contains(view, "[DRY RUN]") {
		t.Errorf("View() should not contain '[DRY RUN]' in normal mode, got: %q", view)
	}
	if !strings.Contains(view, "Committed successfully") {
		t.Errorf("View() should contain 'Committed successfully', got: %q", view)
	}
}

// TestDryRun_View_confirmState checks that confirm screen shows [DRY RUN] in help text.
func TestDryRun_View_confirmState(t *testing.T) {
	mock := &execx.MockRunner{}
	service := newTestService(mock)

	m := New(service, "diff", "stat", true)
	m.commitType = "fix"
	m.subject = "resolve issue"
	m.state = stateConfirm

	view := m.View()
	if !strings.Contains(view, "[DRY RUN]") {
		t.Errorf("confirm view should contain '[DRY RUN]' in help text, got: %q", view)
	}
}

// TestNoDryRun_View_confirmState checks that confirm screen shows normal help text.
func TestNoDryRun_View_confirmState(t *testing.T) {
	mock := &execx.MockRunner{}
	service := newTestService(mock)

	m := New(service, "diff", "stat", false)
	m.commitType = "fix"
	m.subject = "resolve issue"
	m.state = stateConfirm

	view := m.View()
	if strings.Contains(view, "[DRY RUN]") {
		t.Errorf("confirm view should not contain '[DRY RUN]' in normal mode, got: %q", view)
	}
	if !strings.Contains(view, "y/Enter to commit") {
		t.Errorf("confirm view should contain 'y/Enter to commit', got: %q", view)
	}
}
