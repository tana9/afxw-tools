package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/tana9/afxw-tools/internal/afx"
)

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
