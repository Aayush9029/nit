package tui

import (
	"fmt"
	"strings"

	"github.com/Aayush9029/nit/internal/render"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
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
	content  string // rendered timeline content
	summary  string
	width    int
	height   int
	ready    bool
	loading  bool
	err      error
	fetchCmd tea.Cmd

	// Vim-style count prefix (e.g. 10j)
	count int

	// Search
	searching bool
	searchInput textinput.Model
	searchQuery string
	matchLines  []int // line numbers with matches
	matchIdx    int   // current match index
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

	ti := textinput.New()
	ti.Placeholder = "search tweets or @username..."
	ti.CharLimit = 100

	return &Model{
		spinner:     s,
		loading:     true,
		fetchCmd:    fetchCmd,
		searchInput: ti,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fetchCmd)
}

func (m *Model) getCount() int {
	if m.count > 0 {
		n := m.count
		m.count = 0
		return n
	}
	return 1
}

func (m *Model) applyFilter() {
	if m.searchQuery == "" {
		m.viewport.SetContent(m.content)
		m.matchLines = nil
		return
	}

	query := strings.ToLower(m.searchQuery)
	lines := strings.Split(m.content, "\n")
	var filtered []string
	var matchLineNums []int

	// Filter: show tweets that match query (check tweet block: header + body)
	i := 0
	for i < len(lines) {
		// Detect tweet blocks: lines starting with ◆ or ◇ (indicator)
		line := lines[i]
		stripped := stripAnsi(line)
		if strings.HasPrefix(stripped, "◆") || strings.HasPrefix(stripped, "◇") {
			// Collect the whole tweet block until next empty line
			block := []string{line}
			blockText := stripped
			i++
			for i < len(lines) && strings.TrimSpace(lines[i]) != "" {
				block = append(block, lines[i])
				blockText += " " + stripAnsi(lines[i])
				i++
			}
			// Include trailing blank line
			if i < len(lines) {
				block = append(block, lines[i])
				i++
			}

			if strings.Contains(strings.ToLower(blockText), query) {
				matchLineNums = append(matchLineNums, len(filtered))
				filtered = append(filtered, block...)
			}
		} else {
			filtered = append(filtered, line)
			i++
		}
	}

	m.matchLines = matchLineNums
	m.matchIdx = 0
	m.viewport.SetContent(strings.Join(filtered, "\n"))
	m.viewport.GotoTop()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Search mode: forward keys to text input
		if m.searching {
			switch msg.String() {
			case "enter":
				m.searching = false
				m.searchQuery = m.searchInput.Value()
				m.applyFilter()
				return m, nil
			case "esc":
				m.searching = false
				m.searchInput.SetValue("")
				m.searchQuery = ""
				m.applyFilter()
				return m, nil
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				return m, cmd
			}
		}

		key := msg.String()

		// Accumulate numeric prefix
		if len(key) == 1 && key[0] >= '1' && key[0] <= '9' {
			m.count = m.count*10 + int(key[0]-'0')
			return m, nil
		}
		if len(key) == 1 && key[0] == '0' && m.count > 0 {
			m.count = m.count * 10
			return m, nil
		}

		switch key {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.searchQuery != "" {
				m.searchQuery = ""
				m.searchInput.SetValue("")
				m.applyFilter()
				return m, nil
			}
			return m, tea.Quit
		case "j", "down":
			n := m.getCount()
			for range n {
				m.viewport.ScrollDown(1)
			}
			return m, nil
		case "k", "up":
			n := m.getCount()
			for range n {
				m.viewport.ScrollUp(1)
			}
			return m, nil
		case "d", "ctrl+d":
			n := m.getCount()
			for range n {
				m.viewport.HalfPageDown()
			}
			return m, nil
		case "u", "ctrl+u":
			n := m.getCount()
			for range n {
				m.viewport.HalfPageUp()
			}
			return m, nil
		case "G":
			m.count = 0
			m.viewport.GotoBottom()
			return m, nil
		case "g":
			m.count = 0
			m.viewport.GotoTop()
			return m, nil
		case "/":
			m.searching = true
			m.searchInput.SetValue("")
			m.searchInput.Focus()
			m.count = 0
			return m, textinput.Blink
		case "n":
			// Next search match
			if len(m.matchLines) > 0 {
				m.matchIdx = (m.matchIdx + 1) % len(m.matchLines)
				m.viewport.SetYOffset(m.matchLines[m.matchIdx])
			}
			return m, nil
		case "N":
			// Previous search match
			if len(m.matchLines) > 0 {
				m.matchIdx--
				if m.matchIdx < 0 {
					m.matchIdx = len(m.matchLines) - 1
				}
				m.viewport.SetYOffset(m.matchLines[m.matchIdx])
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		footerH := 2
		if m.searching {
			footerH = 3
		}
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-footerH)
			m.ready = true
			if !m.loading && m.content != "" {
				m.viewport.SetContent(m.content)
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

		// Build content with header
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
		dimHeader := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

		var header strings.Builder
		header.WriteString("\n")
		header.WriteString("  " + headerStyle.Render("nit") + dimHeader.Render(" · github.com/Aayush9029/nit") + "\n")
		header.WriteString("  " + dimHeader.Render(fmt.Sprintf("%d accounts · %s", len(m.entries), m.summary)) + "\n")
		header.WriteString("\n")

		timeline := render.FormatTimeline(m.entries, true)
		if timeline == "" {
			timeline = "  No tweets in feed.\n"
		}
		m.content = header.String() + timeline

		if m.ready {
			m.viewport.SetContent(m.content)
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

	var footer string
	if m.searching {
		footer = "  /" + m.searchInput.View()
	} else if m.searchQuery != "" {
		matches := len(m.matchLines)
		footer = footerStyle.Render(
			fmt.Sprintf("%s  ·  /%s (%d matches) · n/N next/prev · esc clear · q quit", m.summary, m.searchQuery, matches),
		)
	} else {
		footer = footerStyle.Render(
			fmt.Sprintf("%s  ·  j/k scroll · / search · q quit", m.summary),
		)
	}

	return m.viewport.View() + "\n" + footer
}

// stripAnsi removes ANSI escape sequences for search matching.
func stripAnsi(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\033' {
			// Skip until 'm' (end of ANSI escape)
			for i < len(s) && s[i] != 'm' {
				i++
			}
			if i < len(s) {
				i++ // skip 'm'
			}
		} else {
			b.WriteByte(s[i])
			i++
		}
	}
	return b.String()
}
