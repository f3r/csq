package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/f3r/csq/internal/config"
	"github.com/f3r/csq/internal/discovery"
	"github.com/f3r/csq/internal/launcher"
	"github.com/f3r/csq/internal/state"
	"github.com/sahilm/fuzzy"
)

var (
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	countStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	pathStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	footerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(1)
)

type model struct {
	input    textinput.Model
	projects []discovery.Project
	filtered []discovery.Project
	cursor   int
	selected *discovery.Project
	quit     bool
	cfg      config.Config
}

type projectSource []discovery.Project

func (p projectSource) String(i int) string { return p[i].Name }
func (p projectSource) Len() int            { return len(p) }

func Pick(projects []discovery.Project, cfg config.Config) (*discovery.Project, error) {
	ti := textinput.New()
	ti.Placeholder = "Search projects..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 40

	m := model{
		input:    ti,
		projects: projects,
		filtered: projects,
		cfg:      cfg,
	}

	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("TUI error: %w", err)
	}

	final := result.(model)
	if final.quit {
		return nil, nil
	}
	return final.selected, nil
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quit = true
			return m, tea.Quit
		case "enter":
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				picked := m.filtered[m.cursor]
				m.selected = &picked
				return m, tea.Quit
			}
			return m, nil
		case "up", "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "ctrl+n":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	query := m.input.Value()
	if query == "" {
		m.filtered = m.projects
	} else {
		matches := fuzzy.FindFrom(query, projectSource(m.projects))
		m.filtered = make([]discovery.Project, len(matches))
		for i, match := range matches {
			m.filtered[i] = m.projects[match.Index]
		}
	}

	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}

	return m, cmd
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString("  Search: ")
	b.WriteString(m.input.View())
	b.WriteString("\n\n")

	maxVisible := 20
	start := 0
	if m.cursor >= maxVisible {
		start = m.cursor - maxVisible + 1
	}

	end := start + maxVisible
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	for i := start; i < end; i++ {
		p := m.filtered[i]
		cursor := "  "
		nameStr := p.Name
		if i == m.cursor {
			cursor = "> "
			nameStr = selectedStyle.Render(p.Name)
		}

		sessions := state.CountSessions(launcher.HomeDir(p.Name, m.cfg))
		sessStr := ""
		if sessions > 0 {
			sessStr = countStyle.Render(fmt.Sprintf(" [%d sessions]", sessions))
		}

		path := pathStyle.Render(p.Path)
		b.WriteString(fmt.Sprintf("%s%-30s%s  %s\n", cursor, nameStr, sessStr, path))
	}

	footer := fmt.Sprintf("%d projects", len(m.projects))
	if len(m.filtered) != len(m.projects) {
		footer += fmt.Sprintf(" · %d matching", len(m.filtered))
	}
	b.WriteString(footerStyle.Render(footer))

	return b.String()
}
