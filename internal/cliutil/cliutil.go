package cliutil

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

var isTerminal = term.IsTerminal

// Notice はユーザーへの案内を表示して正常終了することを表します。
type Notice struct{ Message string }

// Error は案内メッセージを返します。
func (n *Notice) Error() string { return n.Message }

// Run はCLIコマンドを実行し、エラー時にメッセージを表示して終了します。
func Run(cmd *cli.Command) {
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		var notice *Notice
		if errors.As(err, &notice) {
			reportNotice(notice.Message, os.Stdin, os.Stdout, isTerminal(int(os.Stdin.Fd())))
			return
		}
		reportError(err, os.Stdin, os.Stderr, isTerminal(int(os.Stdin.Fd())))
		os.Exit(1)
	}
}

func reportNotice(message string, stdin io.Reader, stdout io.Writer, interactive bool) {
	_, _ = fmt.Fprintln(stdout, message)
	if !interactive {
		return
	}
	_, _ = fmt.Fprintln(stdout, "Enterキーを押すと終了します...")
	_, _ = fmt.Fscanln(stdin)
}

// PressEnterToContinue はメッセージを表示し、対話端末であればEnterキーの入力を待ってから処理を継続します。
// TUI描画等の後続処理に案内メッセージが上書きされて読めなくなるのを防ぐためのヘルパーです。
func PressEnterToContinue(message string) {
	pressEnterToContinue(message, os.Stdin, os.Stdout, isTerminal(int(os.Stdin.Fd())))
}

func pressEnterToContinue(message string, stdin io.Reader, stdout io.Writer, interactive bool) {
	_, _ = fmt.Fprintln(stdout, message)
	if !interactive {
		return
	}
	_, _ = fmt.Fprintln(stdout, "Enterキーを押すと続行します...")
	_, _ = fmt.Fscanln(stdin)
}

func reportError(err error, stdin io.Reader, stderr io.Writer, interactive bool) {
	_, _ = fmt.Fprintf(stderr, "エラー: %v\n", err)
	if !interactive {
		return
	}

	_, _ = fmt.Fprintln(stderr, "Enterキーを押すと終了します...")
	_, _ = fmt.Fscanln(stdin)
}
