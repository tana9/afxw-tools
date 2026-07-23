package main

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tana9/afxw-tools/cmd/afxw-launcher/config"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "5", Dark: "170"}).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "5", Dark: "170"}).
			Bold(true).
			PaddingLeft(2)

	normalStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "243", Dark: "241"}).
			PaddingLeft(4)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "243", Dark: "241"}).
			MarginTop(1)
)

type model struct {
	cfg      *config.Config
	cursor   int
	selected bool
	quitting bool
	// numberInput は数字キーで入力中のメニュー番号(1始まり)を保持するバッファ。
	numberInput string
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
			m.numberInput = ""
			if len(m.cfg.Menu) > 0 {
				m.cursor = (m.cursor - 1 + len(m.cfg.Menu)) % len(m.cfg.Menu)
			}

		case "down", "j":
			m.numberInput = ""
			if len(m.cfg.Menu) > 0 {
				m.cursor = (m.cursor + 1) % len(m.cfg.Menu)
			}

		case "enter":
			if m.numberInput != "" {
				m.selectByNumber(m.numberInput)
				m.numberInput = ""
				if m.selected {
					return m, tea.Quit
				}
				return m, nil
			}
			if len(m.cfg.Menu) > 0 {
				m.selected = true
				return m, tea.Quit
			}

		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			if m.inputDigit(msg.String()) {
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

// inputDigit はメニュー番号入力バッファに数字を1桁追加します。
// 追加後の数値がメニュー総数を超える場合はそのキー入力を無視します。
// もう1桁追加すると必ずメニュー総数を超えてしまい候補が一意に定まる場合は、
// Enterを待たずに即座に選択します(メニューが9件以下なら常にこの条件を満たすため、
// 従来の1桁即時選択と同じ挙動になります)。選択が確定した場合はtrueを返します。
func (m *model) inputDigit(digit string) bool {
	if m.numberInput == "" && digit == "0" {
		return false
	}

	next := m.numberInput + digit
	n, err := strconv.Atoi(next)
	if err != nil || n > len(m.cfg.Menu) {
		return false
	}
	m.numberInput = next

	if n*10 > len(m.cfg.Menu) {
		m.selectByNumber(m.numberInput)
		m.numberInput = ""
		return m.selected
	}
	return false
}

// selectByNumber は1始まりの番号文字列をメニューのインデックスに変換して選択状態にします。
// 範囲外の場合は何もしません。
func (m *model) selectByNumber(input string) {
	n, err := strconv.Atoi(input)
	if err != nil {
		return
	}
	idx := n - 1
	if idx < 0 || idx >= len(m.cfg.Menu) {
		return
	}
	m.cursor = idx
	m.selected = true
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
		if strings.TrimSpace(item.Description) != "" {
			sb.WriteString(descStyle.Render(item.Description))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString(helpStyle.Render("↑/k: 上, ↓/j: 下, Enter: 実行, 数字: 番号で選択, q/Esc: 終了"))
	if m.numberInput != "" {
		sb.WriteString("\n")
		sb.WriteString(helpStyle.Render(fmt.Sprintf("番号入力: %s (Enterで決定)", m.numberInput)))
	}

	return sb.String()
}
