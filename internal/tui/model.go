package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"git-cx/internal/ai"
	"git-cx/internal/commit"
	"git-cx/internal/config"
	"git-cx/internal/git"
)

// State represents TUI step.
type State int

const (
	stateSelectType State = iota
	stateInputScope
	stateAILoading
	stateSelectMsg
	stateInputMsg
	stateInputBody
	stateInputFooter
	stateConfirm
	stateDone
)

// item is a simple list.Item implementation.
type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

// aiResultMsg carries the AI generation result.
type aiResultMsg struct {
	candidates []string
	err        error
}

// commitDoneMsg signals that git commit completed.
type commitDoneMsg struct{ err error }

// Model is the bubbletea model.
type Model struct {
	state    State
	cfg      *config.Config
	provider ai.Provider
	diff     string

	typeList list.Model
	msgList  list.Model
	input    textinput.Model
	body     textarea.Model
	spin     spinner.Model

	commitType string
	scope      string
	candidates []string
	subject    string
	bodyText   string
	footer     string

	err      error
	quitting bool

	width  int
	height int
}

// New creates a new TUI Model.
func New(cfg *config.Config, provider ai.Provider, diff string) Model {
	// Type selector list
	typeItems := make([]list.Item, len(commit.CommitTypes))
	for i, t := range commit.CommitTypes {
		typeItems[i] = item{title: t, desc: commit.CommitTypeDescriptions[t]}
	}
	typeList := list.New(typeItems, list.NewDefaultDelegate(), 0, 0)
	typeList.Title = "Select commit type"
	typeList.SetShowStatusBar(false)
	typeList.SetFilteringEnabled(false)

	inp := textinput.New()
	inp.Placeholder = "(optional) press Enter to skip"
	inp.Focus()

	ta := textarea.New()
	ta.Placeholder = "(optional) press Ctrl+D to skip"
	ta.SetWidth(60)
	ta.SetHeight(5)

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = selectedStyle

	return Model{
		state:    stateSelectType,
		cfg:      cfg,
		provider: provider,
		diff:     diff,
		typeList: typeList,
		input:    inp,
		body:     ta,
		spin:     sp,
	}
}

func (m Model) Init() tea.Cmd {
	return m.spin.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.typeList.SetSize(msg.Width, msg.Height-4)
		if len(m.msgList.Items()) > 0 {
			m.msgList.SetSize(msg.Width, msg.Height-4)
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd

	case aiResultMsg:
		return m.handleAIResult(msg)

	case commitDoneMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		m.quitting = true
		return m, tea.Quit
	}

	return m.updateChildren(msg)
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlC {
		m.quitting = true
		return m, tea.Quit
	}

	switch m.state {
	case stateSelectType:
		if msg.Type == tea.KeyEnter {
			if i, ok := m.typeList.SelectedItem().(item); ok {
				m.commitType = i.title
				m.state = stateInputScope
				m.input.Placeholder = "(optional) scope, press Enter to skip"
				m.input.SetValue("")
				m.input.Focus()
			}
			return m, nil
		}
		var cmd tea.Cmd
		m.typeList, cmd = m.typeList.Update(msg)
		return m, cmd

	case stateInputScope:
		if msg.Type == tea.KeyEnter {
			m.scope = m.input.Value()
			m.input.SetValue("")
			m.state = stateAILoading
			return m, tea.Batch(m.spin.Tick, m.generateAI())
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd

	case stateSelectMsg:
		if msg.Type == tea.KeyEnter {
			if i, ok := m.msgList.SelectedItem().(item); ok {
				switch i.title {
				case "[手動入力]":
					m.state = stateInputMsg
					m.input.Placeholder = "commit subject"
					m.input.SetValue("")
					m.input.Focus()
				case "[再生成]":
					m.state = stateAILoading
					return m, tea.Batch(m.spin.Tick, m.generateAI())
				default:
					m.subject = i.title
					m.state = stateInputBody
					m.body.SetValue("")
					m.body.Focus()
				}
			}
			return m, nil
		}
		var cmd tea.Cmd
		m.msgList, cmd = m.msgList.Update(msg)
		return m, cmd

	case stateInputMsg:
		if msg.Type == tea.KeyEnter {
			m.subject = m.input.Value()
			m.input.SetValue("")
			m.state = stateInputBody
			m.body.SetValue("")
			m.body.Focus()
			return m, nil
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd

	case stateInputBody:
		if msg.Type == tea.KeyCtrlD {
			m.bodyText = m.body.Value()
			m.state = stateInputFooter
			m.input.Placeholder = "(optional) footer, press Enter to skip"
			m.input.SetValue("")
			m.input.Focus()
			return m, nil
		}
		var cmd tea.Cmd
		m.body, cmd = m.body.Update(msg)
		return m, cmd

	case stateInputFooter:
		if msg.Type == tea.KeyEnter {
			m.footer = m.input.Value()
			m.input.SetValue("")
			m.state = stateConfirm
			return m, nil
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd

	case stateConfirm:
		switch msg.String() {
		case "y", "Y":
			m.state = stateDone
			return m, m.doCommit()
		case "n", "N", "q":
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m Model) handleAIResult(msg aiResultMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil || len(msg.candidates) == 0 {
		m.err = msg.err
		m.state = stateInputMsg
		m.input.Placeholder = "commit subject"
		m.input.SetValue("")
		m.input.Focus()
		return m, nil
	}

	m.candidates = msg.candidates
	items := make([]list.Item, 0, len(msg.candidates)+2)
	for _, c := range msg.candidates {
		items = append(items, item{title: c})
	}
	items = append(items, item{title: "[手動入力]", desc: "Enter a commit message manually"})
	items = append(items, item{title: "[再生成]", desc: "Re-generate with AI"})

	m.msgList = list.New(items, list.NewDefaultDelegate(), m.width, m.height-4)
	m.msgList.Title = "Select commit message"
	m.msgList.SetShowStatusBar(false)
	m.msgList.SetFilteringEnabled(false)
	m.state = stateSelectMsg
	return m, nil
}

func (m Model) updateChildren(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.state {
	case stateSelectType:
		m.typeList, cmd = m.typeList.Update(msg)
	case stateInputScope, stateInputMsg, stateInputFooter:
		m.input, cmd = m.input.Update(msg)
	case stateInputBody:
		m.body, cmd = m.body.Update(msg)
	case stateSelectMsg:
		m.msgList, cmd = m.msgList.Update(msg)
	}
	return m, cmd
}

func (m Model) generateAI() tea.Cmd {
	return func() tea.Msg {
		req := ai.GenerateRequest{
			Diff:       m.diff,
			CommitType: m.commitType,
			Scope:      m.scope,
			Candidates: m.cfg.Candidates,
		}
		candidates, err := m.provider.Generate(context.Background(), req)
		return aiResultMsg{candidates: candidates, err: err}
	}
}

func (m Model) doCommit() tea.Cmd {
	return func() tea.Msg {
		c := &commit.ConventionalCommit{
			Type:    m.commitType,
			Scope:   m.scope,
			Subject: m.subject,
			Body:    m.bodyText,
			Footer:  m.footer,
		}
		msg := buildCommitMessage(m, c)
		err := git.Commit(msg)
		return commitDoneMsg{err: err}
	}
}

// buildCommitMessage decides whether to format or use raw subject.
func buildCommitMessage(m Model, c *commit.ConventionalCommit) string {
	if isConventionalHeader(c.Subject) {
		result := c.Subject
		if c.Body != "" {
			result += "\n\n" + c.Body
		}
		if c.Footer != "" {
			result += "\n\n" + c.Footer
		}
		return result
	}
	return commit.Format(c, m.cfg.Commit.UseEmoji, m.cfg.Commit.MaxSubjectLength)
}

func isConventionalHeader(s string) bool {
	for _, t := range commit.CommitTypes {
		if len(s) > len(t) && s[:len(t)] == t {
			rest := s[len(t):]
			if len(rest) > 0 && (rest[0] == '(' || rest[0] == ':' || rest[0] == '!') {
				return true
			}
		}
	}
	return false
}

// View renders the current state.
func (m Model) View() string {
	if m.quitting {
		if m.err != nil {
			return errorStyle.Render(fmt.Sprintf("Error: %v\n", m.err))
		}
		if m.state == stateDone {
			return selectedStyle.Render("Committed successfully!\n")
		}
		return dimStyle.Render("Aborted.\n")
	}

	switch m.state {
	case stateSelectType:
		return m.typeList.View()

	case stateInputScope:
		return fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			titleStyle.Render("Enter scope"),
			m.input.View(),
			helpStyle.Render("Enter to confirm • Ctrl+C to quit"),
		)

	case stateAILoading:
		return fmt.Sprintf(
			"\n  %s Generating commit messages...\n\n%s",
			m.spin.View(),
			helpStyle.Render("Ctrl+C to quit"),
		)

	case stateSelectMsg:
		view := m.msgList.View()
		if m.err != nil {
			view = errorStyle.Render(fmt.Sprintf("AI error: %v\n\n", m.err)) + view
		}
		return view

	case stateInputMsg:
		errMsg := ""
		if m.err != nil {
			errMsg = errorStyle.Render(fmt.Sprintf("AI error: %v\n\n", m.err))
		}
		return fmt.Sprintf(
			"%s%s\n\n%s\n\n%s",
			errMsg,
			titleStyle.Render("Enter commit message"),
			m.input.View(),
			helpStyle.Render("Enter to confirm • Ctrl+C to quit"),
		)

	case stateInputBody:
		return fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			titleStyle.Render("Enter commit body (optional)"),
			m.body.View(),
			helpStyle.Render("Ctrl+D to skip • Ctrl+C to quit"),
		)

	case stateInputFooter:
		return fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			titleStyle.Render("Enter commit footer (optional)"),
			m.input.View(),
			helpStyle.Render("Enter to confirm • Ctrl+C to quit"),
		)

	case stateConfirm:
		c := &commit.ConventionalCommit{
			Type:    m.commitType,
			Scope:   m.scope,
			Subject: m.subject,
			Body:    m.bodyText,
			Footer:  m.footer,
		}
		preview := buildCommitMessage(m, c)
		return fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			titleStyle.Render("Confirm commit message"),
			previewStyle.Render(preview),
			helpStyle.Render("y to commit • n to abort • Ctrl+C to quit"),
		)

	case stateDone:
		return selectedStyle.Render("Committing...\n")
	}

	return ""
}
