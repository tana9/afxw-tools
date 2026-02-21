package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/tana9/afxw-tools/cmd/afxw-bm/bookmark"
	"github.com/tana9/afxw-tools/internal/afx"
	"github.com/tana9/afxw-tools/internal/finder"
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
			// -a フラグが指定されている場合
			if cmd.IsSet("add") {
				target := cmd.String("add")

				// パスが指定されていない場合、あふwから取得を試みる
				if target == "" || target == "." {
					if a, err := afx.NewOleAFX(); err == nil {
						defer a.Close()
						if path, err := a.GetActivePath(); err == nil && path != "" {
							target = path
						}
					}
					// まだ空の場合（例：あふwが起動していない）、カレントディレクトリを使用
					if target == "" {
						target = "."
					}
				}

				return addBookmark(target)
			}

			// デフォルト動作: ブックマーク選択
			bmPath, err := bookmark.GetDefaultPath()
			if err != nil {
				return fmt.Errorf("ブックマークファイルのパス取得に失敗しました: %w", err)
			}

			a, err := afx.NewOleAFX()
			if err != nil {
				return fmt.Errorf("afxw.obj への接続に失敗しました: %w", err)
			}
			defer a.Close()

			return runSelect(a, &finder.GoFuzzyFinder{}, bmPath)
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		fmt.Fprintln(os.Stderr, "何かキーを押すと終了します...")
		fmt.Scanln()
		os.Exit(1)
	}
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
		return fmt.Errorf("ブックマークの追加に失敗しました: %w", err)
	}

	fmt.Printf("ブックマークに追加しました: %s\n", absPath)
	return nil
}

func runSelect(a afx.AFX, f finder.Finder, bmPath string) error {
	dirs, err := bookmark.Load(bmPath)
	if err != nil {
		return fmt.Errorf("ブックマークの読み込みに失敗しました: %w", err)
	}

	if len(dirs) == 0 {
		fmt.Println("ブックマークが見つかりません。'afxw-bm -a' でブックマークを追加してください。")
		return nil
	}

	idx, err := f.Find(dirs)
	if err != nil {
		// ESCやCtrl+Cでキャンセルされた場合は正常終了
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return nil
		}
		return err
	}

	if err := a.EXCD(dirs[idx]); err != nil {
		return fmt.Errorf("ディレクトリ移動に失敗しました: %w", err)
	}

	return nil
}
