package main

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/tana9/afxw-tools/cmd/afxw-open/config"
	"github.com/tana9/afxw-tools/internal/cliutil"
	"github.com/tana9/afxw-tools/internal/singleinstance"
	"github.com/urfave/cli/v3"
)

var version = "dev"

func main() {
	cmd := &cli.Command{
		Name:      "afxw-open",
		Usage:     "あふwで選択したファイルを任意のプログラムで開く",
		Version:   version,
		ArgsUsage: "<ファイルパス>...",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			files := cmd.Args().Slice()
			if len(files) == 0 {
				return fmt.Errorf("ファイルパスを指定してください")
			}
			return run(files)
		},
	}

	cliutil.Run(cmd)
}

// run はメインロジックを実行します。
func run(files []string) error {
	if err := singleinstance.Acquire("afxw-open"); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("設定の読み込みに失敗しました: %w", err)
	}

	return runWithConfig(files, cfg, selectProgram, openFiles)
}

func runWithConfig(files []string, cfg *config.Config, selectProgram func([]config.Program) (int, error), openFiles func(config.Program, []string) error) error {
	if len(cfg.Programs) == 0 {
		return fmt.Errorf("プログラムが設定されていません")
	}

	idx, err := selectProgram(cfg.Programs)
	if err != nil {
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return nil // ESC / Ctrl+C でキャンセル
		}
		return fmt.Errorf("プログラムの選択に失敗しました: %w", err)
	}

	return openFiles(cfg.Programs[idx], files)
}

// selectProgram はfuzzyfinderでプログラムを選択します。
func selectProgram(programs []config.Program) (int, error) {
	return fuzzyfinder.Find(programs, func(i int) string {
		p := programs[i]
		if p.Description != "" {
			return fmt.Sprintf("%s - %s", p.Name, p.Description)
		}
		return p.Name
	})
}

// openFiles は指定されたプログラムでファイルを開きます。
// プログラムは非同期で起動し、ツール自体はすぐに終了します。
func openFiles(p config.Program, files []string) error {
	return openFilesWith(p, files, config.FindCommand, startCommand)
}

func openFilesWith(p config.Program, files []string, findCommand func(string) (string, error), startCommand func(string, []string) error) error {
	cmdPath, err := findCommand(p.Command)
	if err != nil {
		return err
	}

	args := make([]string, 0, len(p.Args)+len(files))
	args = append(args, p.Args...)
	args = append(args, files...)

	if err := startCommand(cmdPath, args); err != nil {
		return fmt.Errorf("プログラムの起動に失敗しました (%s): %w", p.Command, err)
	}

	return nil
}

// startCommand はコマンドを非同期で起動します。終了は待ちません。
func startCommand(path string, args []string) error {
	return exec.Command(path, args...).Start()
}
