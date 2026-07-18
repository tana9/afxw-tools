package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/tana9/afxw-tools/cmd/afxw-bm/bookmark"
	"github.com/tana9/afxw-tools/internal/afx"
	"github.com/tana9/afxw-tools/internal/cliutil"
	"github.com/tana9/afxw-tools/internal/finder"
	"github.com/tana9/afxw-tools/internal/singleinstance"
	"github.com/urfave/cli/v3"
)

var version = "dev"

var (
	acquireBookmarkInstance = singleinstance.Acquire
	newAFXClient            = afx.NewOLEClient
	defaultBookmarkPath     = bookmark.GetDefaultPath
	addBookmarkEntry        = bookmark.Add
	newBookmarkFinder       = func() finder.Finder { return &finder.FuzzyFinder{} }
)

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
		Action: func(_ context.Context, cmd *cli.Command) error {
			return runBookmarkCommand(cmd.IsSet("add"), cmd.String("add"))
		},
	}

	cliutil.Run(cmd)
}

func runBookmarkCommand(add bool, target string) error {
	if err := acquireBookmarkInstance("afxw-bm"); err != nil {
		return err
	}
	if add {
		return addBookmark(resolveBookmarkTarget(target))
	}
	return selectBookmark()
}

func resolveBookmarkTarget(target string) string {
	if target != "" && target != "." {
		return target
	}

	client, err := newAFXClient()
	if err != nil {
		return "."
	}
	defer client.Close()

	if path, err := client.ActivePath(); err == nil && path != "" {
		return path
	}
	return "."
}

func selectBookmark() error {
	bmPath, err := defaultBookmarkPath()
	if err != nil {
		return fmt.Errorf("ブックマークファイルのパス取得に失敗しました: %w", err)
	}

	client, err := newAFXClient()
	if err != nil {
		return fmt.Errorf("afxw.obj への接続に失敗しました: %w", err)
	}
	defer client.Close()

	return runSelect(client, newBookmarkFinder(), bmPath)
}

func addBookmark(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("絶対パスの解決に失敗しました: %w", err)
	}

	bmPath, err := defaultBookmarkPath()
	if err != nil {
		return fmt.Errorf("ブックマークファイルのパス取得に失敗しました: %w", err)
	}

	if err := addBookmarkEntry(bmPath, absPath); err != nil {
		return fmt.Errorf("ブックマークの追加に失敗しました: %w", err)
	}

	fmt.Printf("ブックマークに追加しました: %s\n", absPath)
	return nil
}

func runSelect(client afx.Client, f finder.Finder, bmPath string) error {
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
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return nil // ESC / Ctrl+C でキャンセル
		}
		return err
	}

	if err := client.ChangeDirectory(dirs[idx]); err != nil {
		return fmt.Errorf("ディレクトリ移動に失敗しました: %w", err)
	}

	return nil
}
