package cmdutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Find はコマンドのフルパスを返します。
// 相対パスの場合は dirs、実行ファイルと同じディレクトリ、PATH の順で検索します。
func Find(command string, dirs ...string) (string, error) {
	if filepath.IsAbs(command) {
		if _, err := os.Stat(command); err == nil {
			return command, nil
		}
		return "", fmt.Errorf("コマンドが見つかりません: %s", command)
	}

	for _, dir := range dirs {
		if p := filepath.Join(dir, command); fileExists(p) {
			return p, nil
		}
	}

	if p := filepath.Join(ExecDir(), command); fileExists(p) {
		return p, nil
	}

	if p, err := exec.LookPath(command); err == nil {
		return p, nil
	}

	return "", fmt.Errorf("コマンドが見つかりません: %s", command)
}

// ExecDir は実行ファイルのディレクトリを返します。
func ExecDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
