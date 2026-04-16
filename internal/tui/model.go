package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/hayatosc/git-cx/internal/app"
	"github.com/hayatosc/git-cx/internal/commit"
)

// State represents TUI step.
type State int

const (
	stateSelectType State = iota
	stateInputScope
	stateSelectProvider
	stateAILoading
	stateSelectMsg
	stateInputMsg
	stateSelectDetailMode
	stateDetailAILoading
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

type providerAction int

const (
	providerActionNone providerAction = iota
	providerActionMessage
	providerActionDetail
)

// aiResultMsg carries the AI generation result.
type aiResultMsg struct {
	candidates []string
	err        error
}

// aiDetailResultMsg carries AI detail generation result.
type aiDetailResultMsg struct {
	body   string
	footer string
	err    error
}

// commitDoneMsg signals that git commit completed.
type commitDoneMsg struct {
	err     error
	message string
	output  string
}

// Model is the bubbletea model.
type Model struct {
	state   State
	service *app.CommitService
	diff    string
	stat    string

	typeList     list.Model
	msgList      list.Model
	detailList   list.Model
	providerList list.Model
	input        textinput.Model
	body         textarea.Model
	spin         spinner.Model

	commitType     string
	scope          string
	candidates     []string
	subject        string
	bodyText       string
	footer         string
	providerName   string
	providerTarget providerAction
	returnState    State

	err      error
	quitting bool

	dryRun    bool
	dryRunMsg string
	logOutput string

	width  int
	height int
}

// New creates a new TUI Model.
func New(service *app.CommitService, diff, stat string, dryRun bool) Model {
	// Type selector list
	typeItems := make([]list.Item, len(commit.CommitTypes))
	for i, t := range commit.CommitTypes {
		typeItems[i] = item{title: t, desc: commit.CommitTypeDescriptions[t]}
	}
	typeList := list.New(typeItems, list.NewDefaultDelegate(), 0, 0)
	typeList.Title = "Select commit type"
	typeList.SetShowStatusBar(false)
	typeList.SetFilteringEnabled(false)

	detailItems := []list.Item{
		item{title: "[Skip]", desc: "Skip body and footer"},
		item{title: "[Generate with AI]", desc: "Generate body/footer with AI"},
		item{title: "[Manual entry]", desc: "Enter body/footer manually"},
	}
	detailList := list.New(detailItems, list.NewDefaultDelegate(), 0, 0)
	detailList.Title = "Select detail input"
	detailList.SetShowStatusBar(false)
	detailList.SetFilteringEnabled(false)

	providerItems := buildProviderItems(service.ProviderNames(), service.CurrentProvider())
	providerList := list.New(providerItems, list.NewDefaultDelegate(), 0, 0)
	providerList.Title = "Select AI provider"
	providerList.SetShowStatusBar(false)
	providerList.SetFilteringEnabled(false)
	selectProviderInList(&providerList, service.CurrentProvider())

	inp := textinput.New()
	inp.Placeholder = "(optional) press Enter to skip"
	inp.Focus()

	ta := textarea.New()
	ta.Placeholder = "(optional) press Esc to skip, Tab to confirm"
	ta.SetWidth(60)
	ta.SetHeight(5)

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = selectedStyle

	return Model{
		state:        stateSelectType,
		service:      service,
		diff:         diff,
		stat:         stat,
		typeList:     typeList,
		detailList:   detailList,
		providerList: providerList,
		input:        inp,
		body:         ta,
		spin:         sp,
		dryRun:       dryRun,
		providerName: service.CurrentProvider(),
	}
}

// LogOutput returns git commit output (stdout+stderr) if available.
func (m Model) LogOutput() string {
	return m.logOutput
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
		m.detailList.SetSize(msg.Width, msg.Height-4)
		m.providerList.SetSize(msg.Width, msg.Height-4)
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

	case aiDetailResultMsg:
		return m.handleAIDetailResult(msg)

	case commitDoneMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		m.dryRunMsg = msg.message
		m.logOutput = msg.output
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
		return m.handleSelectTypeKey(msg)
	case stateInputScope:
		return m.handleInputScopeKey(msg)
	case stateSelectMsg:
		return m.handleSelectMsgKey(msg)
	case stateInputMsg:
		return m.handleInputMsgKey(msg)
	case stateSelectDetailMode:
		return m.handleSelectDetailModeKey(msg)
	case stateInputBody:
		return m.handleInputBodyKey(msg)
	case stateInputFooter:
		return m.handleInputFooterKey(msg)
	case stateConfirm:
		return m.handleConfirmKey(msg)
	case stateSelectProvider:
		return m.handleSelectProviderKey(msg)
	}

	return m, nil
}

func (m Model) handleSelectTypeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
}

func (m Model) handleInputScopeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyEnter {
		m.scope = m.input.Value()
		m.input.SetValue("")
		m.state = stateAILoading
		return m, m.startAIGeneration()
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) handleSelectMsgKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlP {
		return m.startProviderSelection(providerActionMessage)
	}
	if msg.Type == tea.KeyEnter {
		if i, ok := m.msgList.SelectedItem().(item); ok {
			switch i.title {
			case "[Manual entry]":
				m.state = stateInputMsg
				m.input.Placeholder = m.subjectPlaceholder()
				m.input.SetValue("")
				m.input.Focus()
			case "[Regenerate]":
				m.state = stateAILoading
				return m, m.startAIGeneration()
			default:
				m.err = nil
				m.subject = i.title
				m.state = stateSelectDetailMode
				m.detailList.Select(1)
			}
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.msgList, cmd = m.msgList.Update(msg)
	return m, cmd
}

func (m Model) handleInputMsgKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlR {
		m.err = nil
		m.state = stateAILoading
		return m, m.startAIGeneration()
	}
	if msg.Type == tea.KeyCtrlP {
		return m.startProviderSelection(providerActionMessage)
	}
	if msg.Type == tea.KeyEnter {
		m.err = nil
		m.subject = m.input.Value()
		m.input.SetValue("")
		m.state = stateSelectDetailMode
		m.detailList.Select(1)
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) handleSelectDetailModeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlR {
		m.err = nil
		m.state = stateDetailAILoading
		return m, m.startAIDetailGeneration()
	}
	if msg.Type == tea.KeyCtrlP {
		return m.startProviderSelection(providerActionDetail)
	}
	if msg.Type == tea.KeyEnter {
		if i, ok := m.detailList.SelectedItem().(item); ok {
			switch i.title {
			case "[Skip]":
				m.err = nil
				m.bodyText = ""
				m.footer = ""
				m.state = stateConfirm
				return m, nil
			case "[Generate with AI]":
				m.err = nil
				m.state = stateDetailAILoading
				return m, m.startAIDetailGeneration()
			default:
				m.err = nil
				m.bodyText = ""
				m.footer = ""
				m.state = stateInputBody
				m.body.SetValue("")
				m.body.Focus()
			}
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.detailList, cmd = m.detailList.Update(msg)
	return m, cmd
}

func (m Model) handleInputBodyKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyEsc {
		m.bodyText = ""
		m.state = stateInputFooter
		m.input.Placeholder = "(optional) footer, press Enter to skip"
		m.input.SetValue(m.footer)
		m.input.Focus()
		return m, nil
	}
	if msg.Type == tea.KeyTab {
		m.bodyText = m.body.Value()
		m.state = stateInputFooter
		m.input.Placeholder = "(optional) footer, press Enter to skip"
		m.input.SetValue(m.footer)
		m.input.Focus()
		return m, nil
	}
	var cmd tea.Cmd
	m.body, cmd = m.body.Update(msg)
	return m, cmd
}

func (m Model) handleInputFooterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyEnter {
		m.footer = m.input.Value()
		m.input.SetValue("")
		m.state = stateConfirm
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) handleSelectProviderKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyEsc {
		m.providerTarget = providerActionNone
		m.state = m.returnState
		return m, nil
	}
	if msg.Type == tea.KeyEnter {
		if i, ok := m.providerList.SelectedItem().(item); ok {
			if err := m.service.UseProvider(i.title); err != nil {
				m.err = err
				m.state = m.returnState
				m.providerTarget = providerActionNone
				return m, nil
			}
			m.providerName = i.title
			m.err = nil
			target := m.providerTarget
			m.providerTarget = providerActionNone
			switch target {
			case providerActionMessage:
				m.state = stateAILoading
				return m, m.startAIGeneration()
			case providerActionDetail:
				m.state = stateDetailAILoading
				return m, m.startAIDetailGeneration()
			default:
				m.state = m.returnState
				return m, nil
			}
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.providerList, cmd = m.providerList.Update(msg)
	return m, cmd
}

func (m Model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyEsc {
		m.state = stateInputFooter
		m.input.Placeholder = "(optional) footer, press Enter to skip"
		m.input.SetValue(m.footer)
		m.input.Focus()
		return m, nil
	}
	switch msg.String() {
	case "y", "Y":
		m.state = stateDone
		return m, m.doCommit()
	case "n", "N":
		m.quitting = true
		return m, tea.Quit
	}
	if msg.Type == tea.KeyEnter {
		m.state = stateDone
		return m, m.doCommit()
	}
	return m, nil
}

func (m Model) handleAIResult(msg aiResultMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil || len(msg.candidates) == 0 {
		m.err = msg.err
		m.state = stateInputMsg
		m.input.Placeholder = m.subjectPlaceholder()
		m.input.SetValue("")
		m.input.Focus()
		return m, nil
	}

	m.candidates = msg.candidates
	items := make([]list.Item, 0, len(msg.candidates)+2)
	for _, c := range msg.candidates {
		items = append(items, item{title: c})
	}
	manualDesc := "Enter a commit message manually"
	if m.commitType == "auto" {
		manualDesc = "Enter a Conventional header manually"
	}
	items = append(items, item{title: "[Manual entry]", desc: manualDesc})
	items = append(items, item{title: "[Regenerate]", desc: "Regenerate with AI"})

	m.msgList = list.New(items, list.NewDefaultDelegate(), m.width, m.height-4)
	m.msgList.Title = "Select commit message"
	m.msgList.SetShowStatusBar(false)
	m.msgList.SetFilteringEnabled(false)
	m.state = stateSelectMsg
	return m, nil
}

func (m Model) handleAIDetailResult(msg aiDetailResultMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.state = stateSelectDetailMode
		m.detailList.Select(1)
		return m, nil
	}

	m.err = nil
	m.bodyText = msg.body
	m.footer = msg.footer
	m.body.SetValue(msg.body)
	m.state = stateInputBody
	m.body.Focus()
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
	case stateSelectDetailMode:
		m.detailList, cmd = m.detailList.Update(msg)
	case stateSelectProvider:
		m.providerList, cmd = m.providerList.Update(msg)
	}
	return m, cmd
}

func (m Model) startProviderSelection(target providerAction) (tea.Model, tea.Cmd) {
	if len(m.service.ProviderNames()) <= 1 {
		return m, nil
	}
	items := buildProviderItems(m.service.ProviderNames(), m.providerName)
	m.providerList.SetItems(items)
	selectProviderInList(&m.providerList, m.providerName)
	m.providerList.SetSize(m.width, m.height-4)
	m.providerTarget = target
	m.returnState = m.state
	m.state = stateSelectProvider
	return m, nil
}

func buildProviderItems(names []string, current string) []list.Item {
	items := make([]list.Item, 0, len(names))
	for _, name := range names {
		desc := ""
		if name == current {
			desc = "current"
		}
		items = append(items, item{title: name, desc: desc})
	}
	return items
}

func selectProviderInList(l *list.Model, current string) {
	for idx, it := range l.Items() {
		if i, ok := it.(item); ok && i.title == current {
			l.Select(idx)
			return
		}
	}
	if len(l.Items()) > 0 {
		l.Select(0)
	}
}

func (m Model) startAIGeneration() tea.Cmd {
	return tea.Batch(m.spin.Tick, m.generateAI())
}

func (m Model) startAIDetailGeneration() tea.Cmd {
	return tea.Batch(m.spin.Tick, m.generateAIDetail())
}

func (m Model) generateAI() tea.Cmd {
	return func() tea.Msg {
		commitType := m.commitType
		if commitType == "auto" {
			commitType = ""
		}
		candidates, err := m.service.GenerateCandidates(context.Background(), m.diff, m.stat, commitType, m.scope)
		return aiResultMsg{candidates: candidates, err: err}
	}
}

func (m Model) generateAIDetail() tea.Cmd {
	return func() tea.Msg {
		commitType := m.commitType
		if commitType == "auto" {
			commitType = ""
		}
		body, footer, err := m.service.GenerateDetails(context.Background(), m.diff, m.stat, commitType, m.scope, m.subject)
		return aiDetailResultMsg{body: body, footer: footer, err: err}
	}
}

func (m Model) subjectPlaceholder() string {
	if m.commitType == "auto" {
		return "commit subject (Conventional header)"
	}
	return "commit subject"
}

func (m Model) commitTypeForMessage() string {
	if m.commitType == "auto" {
		return ""
	}
	return m.commitType
}

func (m Model) doCommit() tea.Cmd {
	return func() tea.Msg {
		c := &commit.ConventionalCommit{
			Type:    m.commitTypeForMessage(),
			Scope:   m.scope,
			Subject: m.subject,
			Body:    m.bodyText,
			Footer:  m.footer,
		}
		msg := m.service.BuildMessage(c)
		if m.dryRun {
			return commitDoneMsg{message: msg}
		}
		out, err := m.service.Commit(context.Background(), msg)
		return commitDoneMsg{output: out, err: err}
	}
}

// View renders the current state.
func (m Model) View() string {
	if m.quitting {
		if m.err != nil {
			return errorStyle.Render(fmt.Sprintf("Error: %v\n", m.err))
		}
		if m.state == stateDone {
			if m.dryRun {
				return fmt.Sprintf("%s\n\n%s\n",
					titleStyle.Render("[DRY RUN] Commit message (not committed):"),
					previewStyle.Render(m.dryRunMsg),
				)
			}
			return selectedStyle.Render("Committed successfully!\n")
		}
		return dimStyle.Render("Aborted.\n")
	}

	switch m.state {
	case stateSelectType:
		return m.typeList.View()
	case stateInputScope:
		return m.viewInputScope()
	case stateAILoading:
		return m.viewAILoading()
	case stateSelectMsg:
		return m.viewSelectMsg()
	case stateInputMsg:
		return m.viewInputMsg()
	case stateSelectDetailMode:
		return m.viewSelectDetailMode()
	case stateSelectProvider:
		return m.viewSelectProvider()
	case stateDetailAILoading:
		return fmt.Sprintf(
			"\n  %s Generating commit details with %s...\n\n%s",
			m.spin.View(),
			m.providerName,
			helpStyle.Render("Ctrl+C to quit"),
		)
	case stateInputBody:
		return m.viewInputBody()
	case stateInputFooter:
		return m.viewInputFooter()
	case stateConfirm:
		return m.viewConfirm()
	case stateDone:
		return selectedStyle.Render("Committing...\n")
	}

	return ""
}

func (m Model) viewSelectProvider() string {
	view := m.providerList.View()
	return view + "\n" + helpStyle.Render("Enter to select • Esc to cancel")
}

func (m Model) viewInputScope() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		titleStyle.Render("Enter scope"),
		m.input.View(),
		helpStyle.Render("Enter to confirm • Ctrl+C to quit"),
	)
}

func (m Model) viewAILoading() string {
	return fmt.Sprintf(
		"\n  %s Generating commit messages with %s...\n\n%s",
		m.spin.View(),
		m.providerName,
		helpStyle.Render("Ctrl+C to quit"),
	)
}

func (m Model) viewSelectMsg() string {
	view := m.msgList.View()
	if m.err != nil {
		view = errorStyle.Render(fmt.Sprintf("AI error: %v\n\n", m.err)) + view
	}
	providerInfo := helpStyle.Render(fmt.Sprintf("Provider: %s", m.providerName))
	return view + "\n" + providerInfo + "\n" + helpStyle.Render("Enter to select • Ctrl+P to switch provider • Ctrl+C to quit")
}

func (m Model) viewInputMsg() string {
	errMsg := ""
	if m.err != nil {
		errMsg = errorStyle.Render(fmt.Sprintf("AI error: %v\n\n", m.err))
	}
	providerInfo := helpStyle.Render(fmt.Sprintf("Provider: %s", m.providerName))
	return fmt.Sprintf(
		"%s%s\n\n%s\n\n%s\n\n%s",
		errMsg,
		titleStyle.Render("Enter commit message"),
		m.input.View(),
		providerInfo,
		helpStyle.Render("Enter to confirm • Ctrl+R to retry AI • Ctrl+P to switch provider • Ctrl+C to quit"),
	)
}

func (m Model) viewSelectDetailMode() string {
	view := m.detailList.View()
	if m.err != nil {
		view = errorStyle.Render(fmt.Sprintf("AI error: %v\n\n", m.err)) + view
	}
	providerInfo := helpStyle.Render(fmt.Sprintf("Provider: %s", m.providerName))
	return view + "\n" + providerInfo + "\n" + helpStyle.Render("Ctrl+R to retry AI • Ctrl+P to switch provider • Ctrl+C to quit")
}

func (m Model) viewInputBody() string {
	errMsg := ""
	if m.err != nil {
		errMsg = errorStyle.Render(fmt.Sprintf("AI error: %v\n\n", m.err))
	}
	return fmt.Sprintf(
		"%s%s\n\n%s\n\n%s",
		errMsg,
		titleStyle.Render("Enter commit body (optional)"),
		m.body.View(),
		helpStyle.Render("Esc to skip • Tab to confirm • Ctrl+C to quit"),
	)
}

func (m Model) viewInputFooter() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		titleStyle.Render("Enter commit footer (optional)"),
		m.input.View(),
		helpStyle.Render("Enter to confirm (empty to skip) • Ctrl+C to quit"),
	)
}

func (m Model) viewConfirm() string {
	c := &commit.ConventionalCommit{
		Type:    m.commitTypeForMessage(),
		Scope:   m.scope,
		Subject: m.subject,
		Body:    m.bodyText,
		Footer:  m.footer,
	}
	preview := m.service.BuildMessage(c)
	helpText := "y/Enter to commit • n to abort • Esc to edit footer • Ctrl+C to quit"
	if m.dryRun {
		helpText = "[DRY RUN] y/Enter to preview • n to abort • Esc to edit footer • Ctrl+C to quit"
	}
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		titleStyle.Render("Confirm commit message"),
		previewStyle.Render(preview),
		helpStyle.Render(helpText),
	)
}
