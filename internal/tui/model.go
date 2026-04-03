package tui

import (
	"fmt"

	"github.com/Aayush9029/nit/internal/render"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FetchDoneMsg struct {
	Entries    []render.TimelineEntry
	Summary    string
	NewCount   int
	APICalls   int
	CostStr    string
}

type FetchErrMsg struct {
	Err error
}

type Model struct {
	viewport viewport.Model
	spinner  spinner.Model
	entries  []render.TimelineEntry
	summary  string
	width    int
	height   int
	ready    bool
	loading  bool
	err      error
	fetchCmd tea.Cmd
}

var (
	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			PaddingLeft(1)
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1"))
)

func NewModel(fetchCmd tea.Cmd) *Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

	return &Model{
		spinner:  s,
		loading:  true,
		fetchCmd: fetchCmd,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fetchCmd)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		footerH := 2
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-footerH)
			m.ready = true
			if !m.loading && len(m.entries) > 0 {
				m.viewport.SetContent(render.FormatTimeline(m.entries, true))
			}
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - footerH
		}

	case FetchDoneMsg:
		m.loading = false
		m.entries = msg.Entries
		m.summary = fmt.Sprintf("%d new · %d API calls · ~%s",
			msg.NewCount, msg.APICalls, msg.CostStr)
		if m.ready {
			content := render.FormatTimeline(m.entries, true)
			if content == "" {
				content = "  No tweets in feed."
			}
			m.viewport.SetContent(content)
		}

	case FetchErrMsg:
		m.loading = false
		m.err = msg.Err

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	if !m.loading && m.ready {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.err != nil {
		return errorStyle.Render("✗ " + m.err.Error())
	}

	if m.loading {
		return fmt.Sprintf("\n  %s Fetching feed...\n", m.spinner.View())
	}

	if !m.ready {
		return ""
	}

	footer := footerStyle.Render(
		fmt.Sprintf("%s  ·  j/k scroll · q quit", m.summary),
	)

	return m.viewport.View() + "\n" + footer
}
