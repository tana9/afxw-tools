package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tana9/afxw-tools/internal/afx"
	"github.com/tana9/afxw-tools/internal/finder"
	"github.com/urfave/cli/v3"
)

var version = "dev"

func main() {
	cmd := &cli.Command{
		Name:    "afxw-his",
		Usage:   "あふwのフォルダ履歴から選択して移動",
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "window",
				Aliases: []string{"w"},
				Usage:   "対象ウィンドウ (left, right, both)",
				Value:   "both",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			a, err := afx.NewOleAFX()
			if err != nil {
				return fmt.Errorf("afxw.objへの接続に失敗しました: %w", err)
			}
			defer a.Close()

			wins, err := parseWindowFlag(cmd.String("window"))
			if err != nil {
				return err
			}

			f := &finder.GoFuzzyFinder{}
			return run(a, f, wins)
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		fmt.Fprintln(os.Stderr, "何かキーを押すと終了します...")
		fmt.Scanln()
		os.Exit(1)
	}
}

func parseWindowFlag(window string) ([]int, error) {
	switch window {
	case "left":
		return []int{afx.WindowLeft}, nil
	case "right":
		return []int{afx.WindowRight}, nil
	case "both":
		return []int{afx.WindowLeft, afx.WindowRight}, nil
	default:
		return nil, fmt.Errorf("無効な対象ウィンドウ: %s", window)
	}
}

func run(a afx.AFX, f finder.Finder, wins []int) error {
	// あふのフォルダ履歴取得
	dirs, err := a.Histories(wins)
	if err != nil {
		return fmt.Errorf("履歴の取得に失敗しました: %w", err)
	}

	// 候補がなければ何もしない
	if len(dirs) == 0 {
		return nil
	}

	// 検索
	idx, err := f.Find(dirs)
	if err != nil {
		return err
	}

	// フォルダ変更
	if err := a.EXCD(dirs[idx]); err != nil {
		return fmt.Errorf("ディレクトリ移動に失敗しました: %w", err)
	}

	return nil
}
