package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

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

// runImport はあふwの履歴をzoxideデータベースにインポートします。
func runImport(a afx.AFX) error {
	dirs, err := a.Histories([]int{afx.WindowLeft, afx.WindowRight})
	if err != nil {
		return fmt.Errorf("履歴の取得に失敗しました: %w", err)
	}

	dirs = removeDuplicates(dirs)

	if len(dirs) == 0 {
		fmt.Println("インポートする履歴がありません。")
		return nil
	}

	// z形式の一時ファイルに書き込む
	tmpFile, err := os.CreateTemp("", "afxw-his-*.txt")
	if err != nil {
		return fmt.Errorf("一時ファイルの作成に失敗しました: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(buildZFormat(dirs, time.Now().Unix())); err != nil {
		tmpFile.Close()
		return fmt.Errorf("一時ファイルへの書き込みに失敗しました: %w", err)
	}
	tmpFile.Close()

	// zoxide import --from z --merge <tmpfile>
	zoxCmd := exec.Command("zoxide", "import", "--from", "z", "--merge", tmpFile.Name())
	zoxCmd.Stdout = os.Stdout
	zoxCmd.Stderr = os.Stderr
	if err := zoxCmd.Run(); err != nil {
		return fmt.Errorf("zoxide importの実行に失敗しました: %w", err)
	}

	fmt.Printf("%d件の履歴をzoxideにインポートしました。\n", len(dirs))
	return nil
}

// buildZFormat はパス一覧をz.sh形式の文字列に変換します。
// 形式: パス|ランク|タイムスタンプ
func buildZFormat(paths []string, timestamp int64) string {
	var sb strings.Builder
	for _, p := range paths {
		fmt.Fprintf(&sb, "%s|1|%d\n", p, timestamp)
	}
	return sb.String()
}

// removeDuplicates はスライスから重複を除去します。出現順序を保持します。
func removeDuplicates(dirs []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		if !seen[dir] {
			seen[dir] = true
			result = append(result, dir)
		}
	}
	return result
}
