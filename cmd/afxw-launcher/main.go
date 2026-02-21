package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tana9/afxw-tools/cmd/afxw-launcher/config"
	"github.com/urfave/cli/v3"
)

var version = "dev"

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
	s += helpStyle.Render("↑/k: 上, ↓/j: 下, Enter: 実行, q/Esc: 終了")

	return s
}

func main() {
	cmd := &cli.Command{
		Name:    "afxw-launcher",
		Usage:   "あふw用ツールランチャー",
		Version: version,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return run()
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		fmt.Fprintln(os.Stderr, "何かキーを押すと終了します...")
		fmt.Scanln()
		os.Exit(1)
	}
}

// run はメインロジックを実行します。
func run() error {
	// 設定ファイルを読み込み
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("設定の読み込みに失敗しました: %w", err)
	}

	if len(cfg.Menu) == 0 {
		return fmt.Errorf("メニュー項目が設定されていません")
	}

	// Bubbletea を起動
	m := model{
		cfg:    cfg,
		cursor: 0,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("メニューの表示に失敗しました: %w", err)
	}

	// 結果を取得
	final := finalModel.(model)
	if !final.selected {
		return nil // キャンセル
	}

	// 選択されたコマンドを実行
	selectedItem := cfg.Menu[final.cursor]
	return executeCommand(cfg, selectedItem)
}

// executeCommand は選択されたコマンドを実行します。
func executeCommand(cfg *config.Config, item config.MenuItem) error {
	// コマンドのフルパスを検索
	cmdPath, err := cfg.FindCommand(item.Command)
	if err != nil {
		return err
	}

	// コマンドを実行
	cmd := exec.Command(cmdPath, item.Args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("コマンドの実行に失敗しました: %w", err)
	}

	return nil
}
