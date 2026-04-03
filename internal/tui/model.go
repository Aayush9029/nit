package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Aayush9029/nit/internal/render"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FetchDoneMsg struct {
	Entries  []render.TimelineEntry
	Summary  string
	NewCount int
	APICalls int
	CostStr  string
}

type FetchErrMsg struct {
	Err error
}

type Model struct {
	viewport viewport.Model
	spinner  spinner.Model
	entries  []render.TimelineEntry
	content  string // full rendered timeline
	summary  string
	width    int
	height   int
	ready    bool
	loading  bool
	err      error
	fetchCmd tea.Cmd

	// Vim count prefix
	count int

	// Filter/search
	filtering   bool
	filterInput textinput.Model
	filterQuery string
	usernames   []string // unique usernames from entries
	completions []string // matching completions for current input
	compIdx     int      // tab-cycle index (-1 = none selected)
}

var (
	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			PaddingLeft(1)
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1"))
	compStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3"))
	compActiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("3"))
)

func NewModel(fetchCmd tea.Cmd) *Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

	ti := textinput.New()
	ti.Placeholder = "filter by @username or keyword..."
	ti.CharLimit = 100

	return &Model{
		spinner:     s,
		loading:     true,
		fetchCmd:    fetchCmd,
		filterInput: ti,
		compIdx:     -1,
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

func (m *Model) extractUsernames() {
	seen := make(map[string]bool)
	for _, e := range m.entries {
		if !seen[e.Username] {
			seen[e.Username] = true
			m.usernames = append(m.usernames, e.Username)
		}
	}
	sort.Strings(m.usernames)
}

func (m *Model) updateCompletions() {
	val := strings.ToLower(m.filterInput.Value())
	m.completions = nil
	m.compIdx = -1
	if val == "" {
		return
	}
	for _, u := range m.usernames {
		if strings.Contains(strings.ToLower(u), val) ||
			strings.Contains(strings.ToLower("@"+u), val) {
			m.completions = append(m.completions, u)
		}
	}
}

func (m *Model) applyFilter(query string) {
	m.filterQuery = query
	if query == "" {
		m.viewport.SetContent(m.content)
		return
	}

	q := strings.ToLower(query)
	lines := strings.Split(m.content, "\n")
	var filtered []string

	i := 0
	for i < len(lines) {
		line := lines[i]
		stripped := stripAnsi(line)
		if strings.HasPrefix(stripped, "◆") || strings.HasPrefix(stripped, "◇") {
			// Collect tweet block
			block := []string{line}
			blockText := stripped
			i++
			for i < len(lines) && strings.TrimSpace(lines[i]) != "" {
				block = append(block, lines[i])
				blockText += " " + stripAnsi(lines[i])
				i++
			}
			if i < len(lines) {
				block = append(block, lines[i])
				i++
			}

			if strings.Contains(strings.ToLower(blockText), q) {
				filtered = append(filtered, block...)
			}
		} else {
			filtered = append(filtered, line)
			i++
		}
	}

	m.viewport.SetContent(strings.Join(filtered, "\n"))
	m.viewport.GotoTop()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.filtering {
			switch msg.String() {
			case "enter":
				m.filtering = false
				// Apply whatever is in the input
				m.applyFilter(m.filterInput.Value())
				return m, nil
			case "esc":
				m.filtering = false
				m.filterInput.SetValue("")
				m.filterQuery = ""
				m.completions = nil
				m.compIdx = -1
				m.applyFilter("")
				return m, nil
			case "tab":
				// Cycle through completions
				if len(m.completions) > 0 {
					m.compIdx = (m.compIdx + 1) % len(m.completions)
					m.filterInput.SetValue("@" + m.completions[m.compIdx])
					m.filterInput.CursorEnd()
					// Live filter as we tab
					m.applyFilter(m.filterInput.Value())
				}
				return m, nil
			case "shift+tab":
				if len(m.completions) > 0 {
					m.compIdx--
					if m.compIdx < 0 {
						m.compIdx = len(m.completions) - 1
					}
					m.filterInput.SetValue("@" + m.completions[m.compIdx])
					m.filterInput.CursorEnd()
					m.applyFilter(m.filterInput.Value())
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.filterInput, cmd = m.filterInput.Update(msg)
				// Live filter as user types
				m.updateCompletions()
				m.applyFilter(m.filterInput.Value())
				return m, cmd
			}
		}

		key := msg.String()

		// Numeric prefix
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
			if m.filterQuery != "" {
				m.filterQuery = ""
				m.filterInput.SetValue("")
				m.completions = nil
				m.compIdx = -1
				m.applyFilter("")
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
			m.filtering = true
			m.filterInput.SetValue(m.filterQuery) // preserve previous query
			m.filterInput.Focus()
			m.filterInput.CursorEnd()
			m.count = 0
			m.updateCompletions()
			return m, textinput.Blink
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		footerH := 3
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
		m.extractUsernames()

		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
		dimHeader := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

		var header strings.Builder
		header.WriteString("\n")
		header.WriteString("  " + headerStyle.Render("nit") + dimHeader.Render(" · github.com/Aayush9029/nit") + "\n")
		header.WriteString("  " + dimHeader.Render(fmt.Sprintf("%d accounts · %s", len(m.usernames), m.summary)) + "\n")
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

	var footer strings.Builder
	if m.filtering {
		// Show input
		footer.WriteString("  / " + m.filterInput.View() + "\n")
		// Show completions
		if len(m.completions) > 0 {
			footer.WriteString("  ")
			for i, c := range m.completions {
				if i > 7 {
					footer.WriteString(compStyle.Render(fmt.Sprintf("+%d more", len(m.completions)-i)))
					break
				}
				if i == m.compIdx {
					footer.WriteString(compActiveStyle.Render(" @"+c+" ") + " ")
				} else {
					footer.WriteString(compStyle.Render("@"+c) + " ")
				}
			}
			footer.WriteString("\n")
		} else if m.filterInput.Value() != "" {
			footer.WriteString("  " + footerStyle.Render("no matches · esc to clear") + "\n")
		} else {
			footer.WriteString("  " + footerStyle.Render("type to filter · tab to complete · esc to cancel") + "\n")
		}
	} else if m.filterQuery != "" {
		footer.WriteString(footerStyle.Render(
			fmt.Sprintf("  %s  ·  filtered: %s · / edit · esc clear · q quit", m.summary, m.filterQuery),
		) + "\n")
	} else {
		footer.WriteString(footerStyle.Render(
			fmt.Sprintf("  %s  ·  j/k scroll · / filter · q quit", m.summary),
		) + "\n")
	}

	return m.viewport.View() + "\n" + footer.String()
}

func stripAnsi(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\033' {
			for i < len(s) && s[i] != 'm' {
				i++
			}
			if i < len(s) {
				i++
			}
		} else {
			b.WriteByte(s[i])
			i++
		}
	}
	return b.String()
}
