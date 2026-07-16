package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/tana9/afxw-tools/cmd/afxw-bm/bookmark"
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
		Name:    "afxw-bm",
		Usage:   "あふw用ブックマーク管理ツール",
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "指定されたパス（省略時はカレントディレクトリまたはあふwのアクティブパス）をブックマークに追加",
				Value:   "",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.IsSet("add") {
				target := cmd.String("add")

				if target == "" || target == "." {
					if a, err := afx.Connect(); err == nil {
						defer a.Close()
						if path, err := a.GetActivePath(); err == nil && path != "" {
							target = path
						}
					}
					// あふwが起動していない場合はカレントディレクトリを使用
					if target == "" {
						target = "."
					}
				}

				return addBookmark(target)
			}

			if err := singleinstance.Acquire("afxw-bm"); err != nil {
				return err
			}

			bmPath, err := bookmark.GetDefaultPath()
			if err != nil {
				return fmt.Errorf("ブックマークファイルのパス取得に失敗しました: %w", err)
			}

			a, err := afx.Connect()
			if err != nil {
				return err
			}
			defer a.Close()

			return runSelect(a, &finder.FuzzyFinder{}, bmPath)
		},
	}

	cliutil.Run(cmd)
}

func addBookmark(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("絶対パスの解決に失敗しました: %w", err)
	}

	bmPath, err := bookmark.GetDefaultPath()
	if err != nil {
		return fmt.Errorf("ブックマークファイルのパス取得に失敗しました: %w", err)
	}

	if err := bookmark.Add(bmPath, absPath); err != nil {
		return err
	}

	fmt.Printf("ブックマークに追加しました: %s\n", absPath)
	return nil
}

// runSelect はブックマークから選択してあふwで移動します。ブックマークが空の場合はNoticeを返します。
func runSelect(a afx.AFX, f finder.Finder, bmPath string) error {
	dirs, err := bookmark.Load(bmPath)
	if err != nil {
		return err
	}

	if len(dirs) == 0 {
		return &cliutil.Notice{Message: "ブックマークが見つかりません。'afxw-bm -a' でブックマークを追加してください。"}
	}

	return selectnav.SelectAndMove(a, f, dirs)
}
