package cliutil

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

// Notice は正常終了扱いだが、終了前にユーザーへ表示してEnter入力を待つ案内メッセージを表します。
// あふwから起動されたコンソールウィンドウが案内を表示する間もなく閉じるのを防ぐため、
// コマンドのコアロジックは自分で待機せず Notice を返し、Run が表示と待機を一元的に行います。
type Notice struct {
	Message string
}

// Error は error インターフェースを満たすため案内メッセージを返します。
func (n *Notice) Error() string {
	return n.Message
}

// Run はCLIコマンドを実行し、Notice なら案内を表示して正常終了、
// それ以外のエラーならメッセージを表示して異常終了します。いずれもEnter入力を待ってから終了します。
func Run(cmd *cli.Command) {
	err := cmd.Run(context.Background(), os.Args)
	if err == nil {
		return
	}

	var notice *Notice
	if errors.As(err, &notice) {
		fmt.Println(notice.Message)
		WaitForEnter()
		return
	}

	fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
	WaitForEnter()
	os.Exit(1)
}

// WaitForEnter は案内を表示してEnterキーの入力を待ちます。
// あふwから起動されたコンソールウィンドウがメッセージ表示直後に閉じて読めなくなるのを防ぐために使います。
func WaitForEnter() {
	fmt.Fprintln(os.Stderr, "Enterキーを押すと終了します...")
	fmt.Scanln()
}
