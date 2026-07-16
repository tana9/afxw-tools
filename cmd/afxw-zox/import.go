package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/tana9/afxw-tools/cmd/afxw-zox/zoxide"
	"github.com/tana9/afxw-tools/internal/afx"
	"github.com/tana9/afxw-tools/internal/stringutil"
)

func runImport(a afx.AFX) error {
	dirs, err := a.Histories([]int{afx.WindowLeft, afx.WindowRight})
	if err != nil {
		return err
	}

	dirs = stringutil.RemoveDuplicates(dirs)

	if len(dirs) == 0 {
		fmt.Println("インポートする履歴がありません。")
		return nil
	}

	tmpFile, err := os.CreateTemp("", "afxw-his-*.txt")
	if err != nil {
		return fmt.Errorf("一時ファイルの作成に失敗しました: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.WriteString(buildZFormat(dirs, time.Now().Unix())); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("一時ファイルへの書き込みに失敗しました: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("一時ファイルのクローズに失敗しました: %w", err)
	}

	zoxCmd := exec.Command(zoxide.ResolveExecutable(), "import", "--from", "z", "--merge", tmpFile.Name())
	zoxCmd.Stdout = os.Stdout
	zoxCmd.Stderr = os.Stderr
	if err := zoxCmd.Run(); err != nil {
		return fmt.Errorf("zoxide importの実行に失敗しました: %w", err)
	}

	fmt.Printf("%d件の履歴をzoxideにインポートしました。\n", len(dirs))
	return nil
}

// buildZFormat はパス一覧をz.sh形式（パス|ランク|タイムスタンプ）に変換します。
func buildZFormat(paths []string, timestamp int64) string {
	var sb strings.Builder
	for _, p := range paths {
		fmt.Fprintf(&sb, "%s|1|%d\n", p, timestamp)
	}
	return sb.String()
}
