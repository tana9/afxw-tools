package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tana9/afxw-tools/cmd/afxw-launcher/config"
	"github.com/tana9/afxw-tools/internal/cliutil"
	"github.com/tana9/afxw-tools/internal/singleinstance"
	"github.com/urfave/cli/v3"
)

var version = "dev"

func main() {
	cmd := &cli.Command{
		Name:    "afxw-launcher",
		Usage:   "あふw用ツールランチャー",
		Version: version,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return run()
		},
	}

	cliutil.Run(cmd)
}

func run() error {
	if err := singleinstance.Acquire("afxw-launcher"); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("設定の読み込みに失敗しました: %w", err)
	}

	if len(cfg.Menu) == 0 {
		return fmt.Errorf("メニュー項目が設定されていません")
	}

	p := tea.NewProgram(model{cfg: cfg})
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("メニューの表示に失敗しました: %w", err)
	}

	final := finalModel.(model)
	if !final.selected {
		return nil
	}

	return executeCommand(cfg, cfg.Menu[final.cursor])
}

func executeCommand(cfg *config.Config, item config.MenuItem) error {
	cmdPath, err := cfg.FindCommand(item.Command)
	if err != nil {
		return err
	}

	args, err := resolveArgs(item.Args)
	if err != nil {
		return err
	}

	cmd := exec.Command(cmdPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("コマンドの実行に失敗しました: %w", err)
	}

	return nil
}
