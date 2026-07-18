package main

import (
	"context"

	"github.com/tana9/afxw-tools/cmd/afxw-zox/zoxide"
	"github.com/tana9/afxw-tools/internal/afx"
	"github.com/tana9/afxw-tools/internal/cliutil"
	"github.com/tana9/afxw-tools/internal/finder"
	"github.com/tana9/afxw-tools/internal/selectnav"
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
			importHistory := cmd.Bool("import-history")
			if !importHistory {
				if err := singleinstance.Acquire("afxw-zox"); err != nil {
					return err
				}
			}

			a, err := afx.Connect()
			if err != nil {
				return err
			}
			defer a.Close()

			if importHistory {
				return runImport(ctx, a)
			}
			return run(a, &finder.FuzzyFinder{}, func() ([]zoxide.Entry, error) {
				return zoxide.Query(ctx)
			})
		},
	}

	cliutil.Run(cmd)
}

// run はzoxideのデータベースから選択してあふwで移動します。データベースが空の場合はNoticeを返します。
func run(a afx.AFX, f finder.Finder, query func() ([]zoxide.Entry, error)) error {
	entries, err := query()
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return &cliutil.Notice{Message: "zoxideデータベースにディレクトリが見つかりません。\nターミナルでディレクトリを移動してzoxideのデータベースを構築してください。"}
	}

	return selectnav.SelectAndMove(a, f, zoxide.Paths(entries))
}
