package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tana9/afxw-tools/cmd/afxw-launcher/config"
)

// スタイル定義
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

// model はアプリケーションの状態を保持します。
type model struct {
	cfg      *config.Config
	cursor   int
	selected bool
	quitting bool
}

// Init は初期化時に実行されるコマンドを返します。
func (m model) Init() tea.Cmd {
	return nil
}

// Update はメッセージに応じて状態を更新します。
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

// View は画面に表示する内容を返します。
func (m model) View() string {
	if m.quitting && !m.selected {
		return ""
	}

	s := titleStyle.Render("=== あふw ツールランチャー ===")
	s += "\n\n"

	for i, item := range m.cfg.Menu {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
			s += selectedStyle.Render(fmt.Sprintf("%s %d. %s", cursor, i+1, item.Name))
		} else {
			s += normalStyle.Render(fmt.Sprintf("%s %d. %s", cursor, i+1, item.Name))
		}
		s += "\n"
		s += descStyle.Render(item.Description)
		s += "\n"
	}

	s += "\n"
	s += helpStyle.Render("↑/k: 上, ↓/j: 下, Enter: 実行, 1-9: 番号で選択, q/Esc: 終了")

	return s
}
