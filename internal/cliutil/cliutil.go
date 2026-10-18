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

const (
	exitPrompt     = "Enterキーを押すと終了します..."
	continuePrompt = "Enterキーを押すと続行します..."
)

// Run はCLIコマンドを実行し、エラー時にメッセージを表示して終了します。
func Run(cmd *cli.Command) {
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		if notice, ok := errors.AsType[*Notice](err); ok {
			showAndWait(notice.Message, exitPrompt, os.Stdin, os.Stdout, stdinIsTerminal())
			return
		}
		reportError(err, os.Stdin, os.Stderr, stdinIsTerminal())
		os.Exit(1)
	}
}

// PressEnterToContinue はメッセージを表示し、対話端末であればEnterキーの入力を待ってから処理を継続します。
// TUI描画等の後続処理に案内メッセージが上書きされて読めなくなるのを防ぐためのヘルパーです。
func PressEnterToContinue(message string) {
	pressEnterToContinue(message, os.Stdin, os.Stdout, stdinIsTerminal())
}

// pressEnterToContinue はPressEnterToContinueの入出力を注入可能にした本体です。
func pressEnterToContinue(message string, stdin io.Reader, stdout io.Writer, interactive bool) {
	showAndWait(message, continuePrompt, stdin, stdout, interactive)
}

// reportError はエラーメッセージを表示し、対話端末であればEnterキーの入力を待ちます。
func reportError(err error, stdin io.Reader, stderr io.Writer, interactive bool) {
	showAndWait("エラー: "+err.Error(), exitPrompt, stdin, stderr, interactive)
}

// showAndWait はメッセージを表示し、対話端末であれば案内を表示してEnterキーの入力を待ちます。
// 非対話時(パイプやリダイレクト)に入力待ちでハングしないよう、待機は対話端末に限定します。
func showAndWait(message, prompt string, stdin io.Reader, w io.Writer, interactive bool) {
	_, _ = fmt.Fprintln(w, message)
	if !interactive {
		return
	}
	_, _ = fmt.Fprintln(w, prompt)
	_, _ = fmt.Fscanln(stdin)
}

// stdinIsTerminal は標準入力が対話端末かどうかを返します。
func stdinIsTerminal() bool {
	return isTerminal(int(os.Stdin.Fd()))
}
