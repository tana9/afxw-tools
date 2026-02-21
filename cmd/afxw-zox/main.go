package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/tana9/afxw-tools/internal/afx"
	"github.com/tana9/afxw-tools/internal/finder"
	"github.com/tana9/afxw-tools/internal/zoxide"
	"github.com/urfave/cli/v3"
)

var version = "dev"

func main() {
	cmd := &cli.Command{
		Name:    "afxw-zox",
		Usage:   "zoxideのfrecencyデータベースから選択してあふwで移動",
		Version: version,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return run()
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		fmt.Fprintln(os.Stderr, "何かキーを押すと終了します...")
		fmt.Scanln()
		os.Exit(1)
	}
}

func run() error {
	// あふwに接続（ファジーファインダーの前に接続して、キャンセル時も確実にCloseされるようにする）
	a, err := afx.NewOleAFX()
	if err != nil {
		return fmt.Errorf("afxw.objへの接続に失敗しました: %w", err)
	}
	defer a.Close()

	// zoxideのディレクトリリストを取得
	entries, err := zoxide.Query()
	if err != nil {
		return fmt.Errorf("zoxideデータベースの取得に失敗しました: %w", err)
	}

	// 候補がなければ何もしない
	if len(entries) == 0 {
		fmt.Println("zoxideデータベースにディレクトリが見つかりません。")
		fmt.Println("ターミナルでディレクトリを移動してzoxideのデータベースを構築してください。")
		return nil
	}

	// パスのみを抽出
	paths := zoxide.Paths(entries)

	// ファジーファインダーで選択
	f := &finder.GoFuzzyFinder{}
	idx, err := f.Find(paths)
	if err != nil {
		// ESCやCtrl+Cでキャンセルされた場合は正常終了
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return nil
		}
		return err
	}

	// ディレクトリ移動
	if err := a.EXCD(paths[idx]); err != nil {
		return fmt.Errorf("ディレクトリ移動に失敗しました: %w", err)
	}

	return nil
}
