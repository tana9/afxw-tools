package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/tana9/afxw-tools/internal/afx"
	"github.com/tana9/afxw-tools/internal/cliutil"
	"github.com/tana9/afxw-tools/internal/finder"
	"github.com/tana9/afxw-tools/internal/singleinstance"
	"github.com/tana9/afxw-tools/internal/stringutil"
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
			if err := singleinstance.Acquire("afxw-his"); err != nil {
				return err
			}

			a, err := afx.NewOleAFX()
			if err != nil {
				return fmt.Errorf("afxw.objへの接続に失敗しました: %w", err)
			}
			defer a.Close()

			wins, err := parseWindowFlag(cmd.String("window"))
			if err != nil {
				return err
			}

			return run(a, &finder.FuzzyFinder{}, wins)
		},
	}

	cliutil.Run(cmd)
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
	dirs, err := a.Histories(wins)
	if err != nil {
		return fmt.Errorf("履歴の取得に失敗しました: %w", err)
	}

	dirs = stringutil.RemoveDuplicates(dirs)

	if len(dirs) == 0 {
		return nil
	}

	idx, err := f.Find(dirs)
	if err != nil {
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return nil // ESC / Ctrl+C でキャンセル
		}
		return err
	}

	if err := a.EXCD(dirs[idx]); err != nil {
		return fmt.Errorf("ディレクトリ移動に失敗しました: %w", err)
	}

	return nil
}
