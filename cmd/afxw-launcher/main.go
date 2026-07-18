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
			return run(ctx)
		},
	}

	cliutil.Run(cmd)
}

func run(ctx context.Context) error {
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

	return executeCommand(ctx, cfg, cfg.Menu[final.cursor])
}

func executeCommand(ctx context.Context, cfg *config.Config, item config.MenuItem) error {
	return executeCommandWith(item, cfg.FindCommand, resolveArgs, func(path string, args []string) error {
		return runCommand(ctx, path, args)
	})
}

func executeCommandWith(item config.MenuItem, findCommand func(string) (string, error), resolveArgs func([]string) ([]string, error), runCommand func(string, []string) error) error {
	cmdPath, err := findCommand(item.Command)
	if err != nil {
		return err
	}

	args, err := resolveArgs(item.Args)
	if err != nil {
		return err
	}

	if err := runCommand(cmdPath, args); err != nil {
		return fmt.Errorf("コマンドの実行に失敗しました: %w", err)
	}

	return nil
}

func runCommand(ctx context.Context, path string, args []string) error {
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
