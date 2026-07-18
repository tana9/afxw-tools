package cliutil

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

var isTerminal = term.IsTerminal

// Run はCLIコマンドを実行し、エラー時にメッセージを表示して終了します。
func Run(cmd *cli.Command) {
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		reportError(err, os.Stdin, os.Stderr, isTerminal(int(os.Stdin.Fd())))
		os.Exit(1)
	}
}

func reportError(err error, stdin io.Reader, stderr io.Writer, interactive bool) {
	fmt.Fprintf(stderr, "エラー: %v\n", err)
	if !interactive {
		return
	}

	fmt.Fprintln(stderr, "何かキーを押すと終了します...")
	_, _ = fmt.Fscanln(stdin)
}
