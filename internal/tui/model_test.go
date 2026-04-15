package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/hayatosc/git-cx/internal/ai"
	"github.com/hayatosc/git-cx/internal/app"
	"github.com/hayatosc/git-cx/internal/config"
	"github.com/hayatosc/git-cx/internal/execx"
	"github.com/hayatosc/git-cx/internal/git"
)

func newTestService(mock *execx.MockRunner) *app.CommitService {
	cfg := &config.Config{Provider: "mock", Candidates: 1, Commit: config.CommitConfig{}, Providers: []string{"mock"}}
	provider := &ai.MockProvider{NameValue: "mock", Candidates: []string{"feat: test"}}
	return app.NewCommitService(cfg, map[string]ai.Provider{"mock": provider}, "mock", git.NewRunnerWithExecutor(mock))
}

func newMultiProviderService() *app.CommitService {
	cfg := &config.Config{
		Provider:   "first",
		Providers:  []string{"first", "second"},
		Candidates: 1,
		Commit:     config.CommitConfig{},
	}
	providers := map[string]ai.Provider{
		"first":  &ai.MockProvider{NameValue: "first", Candidates: []string{"feat: first"}},
		"second": &ai.MockProvider{NameValue: "second", Candidates: []string{"feat: second"}},
	}
	return app.NewCommitService(cfg, providers, "first", git.NewRunnerWithExecutor(&execx.MockRunner{}))
}

func newModel(dryRun bool) Model {
	return New(newTestService(&execx.MockRunner{}), "diff", "stat", dryRun)
}

func pressEnter() tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyEnter} }
func pressKey(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

// --- Normal mode: View states ---

func TestView_selectType(t *testing.T) {
	m := newModel(false)
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = result.(Model)
	view := m.View()
	if !strings.Contains(view, "commit type") {
		t.Errorf("selectType view missing prompt, got: %q", view)
	}
}

func TestView_inputScope(t *testing.T) {
	m := newModel(false)
	m.state = stateInputScope
	view := m.View()
	if !strings.Contains(view, "scope") {
		t.Errorf("inputScope view missing 'scope', got: %q", view)
	}
}

func TestView_aiLoading(t *testing.T) {
	m := newModel(false)
	m.state = stateAILoading
	view := m.View()
	if !strings.Contains(view, "Generating") {
		t.Errorf("aiLoading view missing 'Generating', got: %q", view)
	}
}

func TestView_inputMsg(t *testing.T) {
	m := newModel(false)
	m.state = stateInputMsg
	view := m.View()
	if !strings.Contains(view, "commit message") {
		t.Errorf("inputMsg view missing prompt, got: %q", view)
	}
}

func TestView_inputBody(t *testing.T) {
	m := newModel(false)
	m.state = stateInputBody
	view := m.View()
	if !strings.Contains(view, "body") {
		t.Errorf("inputBody view missing 'body', got: %q", view)
	}
}

func TestView_inputFooter(t *testing.T) {
	m := newModel(false)
	m.state = stateInputFooter
	view := m.View()
	if !strings.Contains(view, "footer") {
		t.Errorf("inputFooter view missing 'footer', got: %q", view)
	}
}

func TestView_confirm(t *testing.T) {
	m := newModel(false)
	m.commitType = "feat"
	m.subject = "add thing"
	m.state = stateConfirm
	view := m.View()
	if !strings.Contains(view, "feat: add thing") {
		t.Errorf("confirm view missing commit message, got: %q", view)
	}
	if !strings.Contains(view, "y/Enter to commit") {
		t.Errorf("confirm view missing help text, got: %q", view)
	}
}

func TestView_aborted(t *testing.T) {
	m := newModel(false)
	m.quitting = true
	view := m.View()
	if !strings.Contains(view, "Aborted") {
		t.Errorf("aborted view missing 'Aborted', got: %q", view)
	}
}

func TestView_errorState(t *testing.T) {
	m := newModel(false)
	m.quitting = true
	m.err = errors.New("something went wrong")
	view := m.View()
	if !strings.Contains(view, "something went wrong") {
		t.Errorf("error view missing error message, got: %q", view)
	}
}

// --- Normal mode: state transitions via key ---

func TestHandleKey_selectType_enter_advancesToInputScope(t *testing.T) {
	m := newModel(false)
	m.state = stateSelectType
	result, _ := m.handleKey(pressEnter())
	next := result.(Model)
	if next.state != stateInputScope {
		t.Errorf("expected stateInputScope, got %v", next.state)
	}
}

func TestHandleKey_inputScope_enter_advancesToAILoading(t *testing.T) {
	m := newModel(false)
	m.state = stateInputScope
	result, _ := m.handleKey(pressEnter())
	next := result.(Model)
	if next.state != stateAILoading {
		t.Errorf("expected stateAILoading, got %v", next.state)
	}
}

func TestHandleKey_inputMsg_enter_advancesToSelectDetailMode(t *testing.T) {
	m := newModel(false)
	m.state = stateInputMsg
	m.input.SetValue("add feature")
	result, _ := m.handleKey(pressEnter())
	next := result.(Model)
	if next.state != stateSelectDetailMode {
		t.Errorf("expected stateSelectDetailMode, got %v", next.state)
	}
	if next.subject != "add feature" {
		t.Errorf("expected subject 'add feature', got %q", next.subject)
	}
}

func TestHandleKey_inputMsg_ctrlP_opensProviderSelection(t *testing.T) {
	m := New(newMultiProviderService(), "diff", "stat", false)
	m.state = stateInputMsg

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlP})
	next := result.(Model)
	if next.state != stateSelectProvider {
		t.Fatalf("expected stateSelectProvider, got %v", next.state)
	}
	next.providerList.Select(1)
	result, _ = next.handleSelectProviderKey(pressEnter())
	final := result.(Model)
	if final.state != stateAILoading {
		t.Fatalf("expected to start AI loading, got %v", final.state)
	}
	if final.providerName != "second" {
		t.Fatalf("expected provider 'second', got %q", final.providerName)
	}
}

func TestHandleKey_inputFooter_enter_advancesToConfirm(t *testing.T) {
	m := newModel(false)
	m.state = stateInputFooter
	result, _ := m.handleKey(pressEnter())
	next := result.(Model)
	if next.state != stateConfirm {
		t.Errorf("expected stateConfirm, got %v", next.state)
	}
}

func TestHandleKey_confirm_y_advancesToDone(t *testing.T) {
	m := newModel(false)
	m.commitType = "feat"
	m.subject = "add thing"
	m.state = stateConfirm
	result, _ := m.handleKey(pressKey('y'))
	next := result.(Model)
	if next.state != stateDone {
		t.Errorf("expected stateDone, got %v", next.state)
	}
}

func TestHandleKey_confirm_n_quits(t *testing.T) {
	m := newModel(false)
	m.state = stateConfirm
	result, _ := m.handleKey(pressKey('n'))
	next := result.(Model)
	if !next.quitting {
		t.Error("expected quitting=true after pressing n")
	}
}

func TestHandleKey_confirm_enter_advancesToDone(t *testing.T) {
	m := newModel(false)
	m.commitType = "fix"
	m.subject = "resolve"
	m.state = stateConfirm
	result, _ := m.handleKey(pressEnter())
	next := result.(Model)
	if next.state != stateDone {
		t.Errorf("expected stateDone, got %v", next.state)
	}
}

func TestHandleKey_ctrlC_quits(t *testing.T) {
	m := newModel(false)
	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyCtrlC})
	next := result.(Model)
	if !next.quitting {
		t.Error("expected quitting=true after Ctrl+C")
	}
}

// --- Normal mode: doCommit and commitDoneMsg ---

func TestDoCommit_propagatesError(t *testing.T) {
	mock := &execx.MockRunner{
		Errors: map[string]error{
			"git\x00commit\x00-m\x00feat: fail": errors.New("git error"),
		},
	}
	service := newTestService(mock)

	m := New(service, "diff", "stat", false)
	m.commitType = "feat"
	m.subject = "fail"

	msg := m.doCommit()()

	done, ok := msg.(commitDoneMsg)
	if !ok {
		t.Fatalf("expected commitDoneMsg, got %T", msg)
	}
	if done.err == nil {
		t.Fatal("expected error to be propagated")
	}
}

func TestUpdate_commitDoneMsg_setsError(t *testing.T) {
	m := newModel(false)
	result, _ := m.Update(commitDoneMsg{err: errors.New("commit failed")})
	next := result.(Model)
	if next.err == nil || !strings.Contains(next.err.Error(), "commit failed") {
		t.Errorf("expected err to be set, got %v", next.err)
	}
	if !next.quitting {
		t.Error("expected quitting=true")
	}
}

func TestUpdate_commitDoneMsg_success(t *testing.T) {
	m := newModel(false)
	result, _ := m.Update(commitDoneMsg{})
	next := result.(Model)
	if next.err != nil {
		t.Errorf("unexpected error: %v", next.err)
	}
	if !next.quitting {
		t.Error("expected quitting=true")
	}
}

func TestUpdate_commitDoneMsg_setsLogOutput(t *testing.T) {
	m := newModel(false)
	output := "stdout\nstderr"
	result, _ := m.Update(commitDoneMsg{output: output})
	next := result.(Model)
	if next.LogOutput() != output {
		t.Fatalf("expected log output %q, got %q", output, next.LogOutput())
	}
}

// --- AI result handling ---

func TestHandleAIResult_populatesMsgList(t *testing.T) {
	m := newModel(false)
	m.state = stateAILoading
	result, _ := m.handleAIResult(aiResultMsg{candidates: []string{"feat: a", "fix: b"}})
	next := result.(Model)
	if next.state != stateSelectMsg {
		t.Errorf("expected stateSelectMsg, got %v", next.state)
	}
	if len(next.msgList.Items()) != 4 { // 2 candidates + Manual + Regenerate
		t.Errorf("expected 4 items in msgList, got %d", len(next.msgList.Items()))
	}
}

func TestHandleAIResult_errorFallsBackToInputMsg(t *testing.T) {
	m := newModel(false)
	m.state = stateAILoading
	result, _ := m.handleAIResult(aiResultMsg{err: errors.New("ai failed")})
	next := result.(Model)
	if next.state != stateInputMsg {
		t.Errorf("expected stateInputMsg, got %v", next.state)
	}
	if next.err == nil {
		t.Error("expected err to be set")
	}
}

// --- dry-run mode ---

func TestDryRun_doCommit_skipsCommit(t *testing.T) {
	mock := &execx.MockRunner{Strict: true}
	service := newTestService(mock)

	m := New(service, "diff content", "stat content", true)
	m.commitType = "feat"
	m.subject = "add dry-run support"

	msg := m.doCommit()()

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

func TestDryRun_doCommit_includesMessage(t *testing.T) {
	m := New(newTestService(&execx.MockRunner{}), "diff", "stat", true)
	m.commitType = "chore"
	m.subject = "update deps"

	done := m.doCommit()().(commitDoneMsg)
	if done.message != "chore: update deps" {
		t.Fatalf("unexpected message: %q", done.message)
	}
}

func TestDryRun_View_doneState(t *testing.T) {
	m := newModel(true)
	m.state = stateDone
	m.quitting = true
	m.dryRunMsg = "feat: add feature"

	view := m.View()
	if !strings.Contains(view, "[DRY RUN]") {
		t.Errorf("expected '[DRY RUN]' in view, got: %q", view)
	}
	if !strings.Contains(view, "feat: add feature") {
		t.Errorf("expected commit message in view, got: %q", view)
	}
}

func TestDryRun_View_confirmState(t *testing.T) {
	m := newModel(true)
	m.commitType = "fix"
	m.subject = "resolve issue"
	m.state = stateConfirm

	view := m.View()
	if !strings.Contains(view, "[DRY RUN]") {
		t.Errorf("expected '[DRY RUN]' in confirm help text, got: %q", view)
	}
}
