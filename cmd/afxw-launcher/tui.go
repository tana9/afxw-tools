package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tana9/afxw-tools/cmd/afxw-launcher/config"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true).
			PaddingLeft(2)

	normalStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			PaddingLeft(4)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)
)

type model struct {
	cfg      *config.Config
	cursor   int
	selected bool
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.cfg.Menu)-1 {
				m.cursor++
			}

		case "enter":
			m.selected = true
			return m, tea.Quit

		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			idx := int(msg.String()[0] - '1')
			if idx < len(m.cfg.Menu) {
				m.cursor = idx
				m.selected = true
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting && !m.selected {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(titleStyle.Render("=== あふw ツールランチャー ==="))
	sb.WriteString("\n\n")

	for i, item := range m.cfg.Menu {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
			sb.WriteString(selectedStyle.Render(fmt.Sprintf("%s %d. %s", cursor, i+1, item.Name)))
		} else {
			sb.WriteString(normalStyle.Render(fmt.Sprintf("%s %d. %s", cursor, i+1, item.Name)))
		}
		sb.WriteString("\n")
		sb.WriteString(descStyle.Render(item.Description))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(helpStyle.Render("↑/k: 上, ↓/j: 下, Enter: 実行, 1-9: 番号で選択, q/Esc: 終了"))

	return sb.String()
}
