package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/tana9/afxw-tools/cmd/afxw-zox/zoxide"
	"github.com/tana9/afxw-tools/internal/afx"
	"github.com/tana9/afxw-tools/internal/cliutil"
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
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "import-history",
				Aliases: []string{"i"},
				Usage:   "あふwの履歴をzoxideデータベースにインポート",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			a, err := afx.NewOleAFX()
			if err != nil {
				return fmt.Errorf("afxw.objへの接続に失敗しました: %w", err)
			}
			defer a.Close()

			if cmd.Bool("import-history") {
				return runImport(a)
			}

			if err := singleinstance.Acquire("afxw-zox"); err != nil {
				return err
			}

			return run(a, &finder.FuzzyFinder{}, zoxide.Query)
		},
	}

	cliutil.Run(cmd)
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
