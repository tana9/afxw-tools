package cliutil

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

// Run はCLIコマンドを実行し、エラー時にメッセージを表示して終了します。
func Run(cmd *cli.Command) {
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		fmt.Fprintln(os.Stderr, "何かキーを押すと終了します...")
		fmt.Scanln()
		os.Exit(1)
	}
}
