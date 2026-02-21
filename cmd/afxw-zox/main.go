package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/tana9/afxw-tools/cmd/afxw-zox/zoxide"
	"github.com/tana9/afxw-tools/internal/afx"
	"github.com/tana9/afxw-tools/internal/finder"
	"github.com/tana9/afxw-tools/internal/singleinstance"
	"github.com/urfave/cli/v3"
)

var version = "dev"

func main() {
	cmd := &cli.Command{
		Name:    "afxw-zox",
		Usage:   "zoxideのfrecencyデータベースから選択してあふwで移動",
		Version: version,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if err := singleinstance.Acquire("afxw-zox"); err != nil {
				if errors.Is(err, singleinstance.ErrAlreadyRunning) {
					return nil
				}
				return err
			}

			a, err := afx.NewOleAFX()
			if err != nil {
				return fmt.Errorf("afxw.objへの接続に失敗しました: %w", err)
			}
			defer a.Close()

			return run(a, &finder.GoFuzzyFinder{}, zoxide.Query)
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		fmt.Fprintln(os.Stderr, "何かキーを押すと終了します...")
		fmt.Scanln()
		os.Exit(1)
	}
}

func run(a afx.AFX, f finder.Finder, query func() ([]zoxide.Entry, error)) error {
	entries, err := query()
	if err != nil {
		return fmt.Errorf("zoxideデータベースの取得に失敗しました: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("zoxideデータベースにディレクトリが見つかりません。")
		fmt.Println("ターミナルでディレクトリを移動してzoxideのデータベースを構築してください。")
		return nil
	}

	paths := zoxide.Paths(entries)

	idx, err := f.Find(paths)
	if err != nil {
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return nil
		}
		return err
	}

	if err := a.EXCD(paths[idx]); err != nil {
		return fmt.Errorf("ディレクトリ移動に失敗しました: %w", err)
	}

	return nil
}
