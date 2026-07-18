package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/tana9/afxw-tools/cmd/afxw-rg/search"
	"github.com/tana9/afxw-tools/internal/afx"
	"github.com/tana9/afxw-tools/internal/cliutil"
	"github.com/tana9/afxw-tools/internal/cmdutil"
	"github.com/tana9/afxw-tools/internal/finder"
	"github.com/tana9/afxw-tools/internal/singleinstance"
	"github.com/urfave/cli/v3"
)

var version = "dev"

func main() {
	cmd := &cli.Command{
		Name:      "afxw-rg",
		Usage:     "あふwの現在のフォルダ以下をripgrepで検索",
		Version:   version,
		ArgsUsage: "[検索キーワード]",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "encoding", Aliases: []string{"E"}, Value: search.EncodingAuto, Usage: "文字コード (auto, utf-8, utf-8bom, sjis)"},
			&cli.BoolFlag{Name: "fixed-strings", Aliases: []string{"F"}, Usage: "キーワードを正規表現ではなく固定文字列として検索"},
			&cli.BoolFlag{Name: "hidden", Usage: "隠しファイルと隠しフォルダも検索"},
			&cli.StringFlag{Name: "glob", Aliases: []string{"g"}, Usage: "検索対象のglobパターン (例: *.go)"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if err := singleinstance.Acquire("afxw-rg"); err != nil {
				return err
			}

			keyword := strings.TrimSpace(strings.Join(cmd.Args().Slice(), " "))

			a, err := afx.Connect()
			if err != nil {
				return err
			}
			defer a.Close()

			root, err := a.GetActivePath()
			if err != nil {
				return fmt.Errorf("アクティブパスの取得に失敗しました: %w", err)
			}
			rgPath, err := cmdutil.Find("rg.exe")
			if err != nil {
				return fmt.Errorf("ripgrepが見つかりません: %w", err)
			}

			opts := search.Options{
				Root:         root,
				Encoding:     strings.ToLower(cmd.String("encoding")),
				Glob:         cmd.String("glob"),
				FixedStrings: cmd.Bool("fixed-strings"),
				Hidden:       cmd.Bool("hidden"),
			}
			query := func(ctx context.Context, opts search.Options) ([]search.Match, error) {
				return search.Run(ctx, rgPath, opts)
			}
			if keyword == "" {
				return runPrompted(ctx, a, &finder.FuzzyFinder{}, opts, os.Stdin, os.Stdout, query)
			}

			opts.Pattern = keyword
			return run(ctx, a, &finder.FuzzyFinder{}, opts, query)
		},
	}

	cliutil.Run(cmd)
}

// promptKeyword は標準入力から検索キーワードを1行読み取ります。
func promptKeyword(r io.Reader, w io.Writer) (string, error) {
	if _, err := fmt.Fprint(w, "検索キーワード: "); err != nil {
		return "", fmt.Errorf("プロンプトの表示に失敗しました: %w", err)
	}
	reader, ok := r.(*bufio.Reader)
	if !ok {
		reader = bufio.NewReader(r)
	}
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("検索キーワードの入力に失敗しました: %w", err)
	}
	return strings.TrimSpace(line), nil
}

// runPrompted は検索キーワードを対話的に受け取り、結果が見つかるまで再入力を受け付けます。
func runPrompted(ctx context.Context, a afx.AFX, f finder.Finder, opts search.Options, r io.Reader, w io.Writer, query func(context.Context, search.Options) ([]search.Match, error)) error {
	reader := bufio.NewReader(r)
	for {
		keyword, err := promptKeyword(reader, w)
		if err != nil {
			return err
		}
		if keyword == "" {
			return nil
		}

		opts.Pattern = keyword
		err = run(ctx, a, f, opts, query)
		var notice *cliutil.Notice
		if errors.As(err, &notice) {
			if _, writeErr := fmt.Fprintln(w, notice.Message); writeErr != nil {
				return fmt.Errorf("検索結果の表示に失敗しました: %w", writeErr)
			}
			continue
		}
		return err
	}
}

// run は検索結果を選択し、あふwを対象ファイルのフォルダへ移動します。
func run(ctx context.Context, a afx.AFX, f finder.Finder, opts search.Options, query func(context.Context, search.Options) ([]search.Match, error)) error {
	matches, err := query(ctx, opts)
	if err != nil {
		return fmt.Errorf("キーワード検索に失敗しました: %w", err)
	}
	if len(matches) == 0 {
		return &cliutil.Notice{Message: "検索結果が見つかりませんでした。"}
	}

	idx, err := f.Find(search.Labels(opts.Root, matches))
	if err != nil {
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return nil
		}
		return fmt.Errorf("検索結果の選択に失敗しました: %w", err)
	}
	if idx < 0 || idx >= len(matches) {
		return fmt.Errorf("検索結果の選択位置が範囲外です: %d", idx)
	}

	if err := a.EXCDFile(matches[idx].Path); err != nil {
		return fmt.Errorf("検索結果のファイルへの移動に失敗しました: %w", err)
	}
	return nil
}
