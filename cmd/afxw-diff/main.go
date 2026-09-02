package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/tana9/afxw-tools/internal/afx"
	"github.com/tana9/afxw-tools/internal/cliutil"
	"github.com/tana9/afxw-tools/internal/cmdutil"
	"github.com/tana9/afxw-tools/internal/singleinstance"
	"github.com/urfave/cli/v3"
)

var version = "dev"

func main() {
	cmd := &cli.Command{
		Name:      "afxw-diff",
		Usage:     "あふwでマークした2つのファイルまたはフォルダをWinMergeで比較",
		Version:   version,
		ArgsUsage: "[比較元] [比較先]",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if err := singleinstance.Acquire("afxw-diff"); err != nil {
				return err
			}

			paths := cmd.Args().Slice()
			if len(paths) == 0 {
				a, err := afx.Connect()
				if err != nil {
					return err
				}
				defer a.Close()

				paths, err = a.GetMarkedFiles()
				if err != nil {
					return fmt.Errorf("マーク項目の取得に失敗しました: %w", err)
				}
			}

			return run(paths, resolveWinMerge(), startCommand)
		},
	}

	cliutil.Run(cmd)
}

func run(paths []string, winMergePath string, start func(string, []string) error) error {
	if len(paths) != 2 {
		return &cliutil.Notice{Message: fmt.Sprintf("比較するファイルまたはフォルダを2つ選択してください（現在%d個）。", len(paths))}
	}
	if err := start(winMergePath, paths); err != nil {
		return fmt.Errorf("WinMergeの起動に失敗しました (%s): %w", winMergePath, err)
	}
	return nil
}

// resolveWinMerge はPATH、ユーザー、64bit、32bitの標準インストール先の順に探索します。
func resolveWinMerge() string {
	return cmdutil.ResolveExecutable("WinMergeU.exe",
		executableIn("LOCALAPPDATA", "Programs", "WinMerge", "WinMergeU.exe"),
		executableIn("ProgramFiles", "WinMerge", "WinMergeU.exe"),
		executableIn("ProgramFiles(x86)", "WinMerge", "WinMergeU.exe"),
	)
}

func executableIn(env string, elems ...string) string {
	root := os.Getenv(env)
	if root == "" {
		return ""
	}
	return filepath.Join(append([]string{root}, elems...)...)
}

// startCommand はコマンドを非同期で起動します。終了は待ちません。
func startCommand(path string, args []string) error {
	return exec.Command(path, args...).Start()
}
