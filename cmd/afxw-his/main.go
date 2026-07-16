package main

import (
	"context"
	"fmt"

	"github.com/tana9/afxw-tools/internal/afx"
	"github.com/tana9/afxw-tools/internal/cliutil"
	"github.com/tana9/afxw-tools/internal/finder"
	"github.com/tana9/afxw-tools/internal/selectnav"
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

			a, err := afx.Connect()
			if err != nil {
				return err
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

// run はフォルダ履歴から選択してあふwで移動します。履歴が空の場合はNoticeを返します。
func run(a afx.AFX, f finder.Finder, wins []int) error {
	dirs, err := a.Histories(wins)
	if err != nil {
		return err
	}

	dirs = stringutil.RemoveDuplicates(dirs)
	if len(dirs) == 0 {
		return &cliutil.Notice{Message: "フォルダ履歴が見つかりません。"}
	}

	return selectnav.SelectAndMove(a, f, dirs)
}
