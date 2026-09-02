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
		if fileExists(command) {
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

// ResolveExecutable はPATHと候補パスから実行ファイルを探します。
// 見つからない場合はcommandをそのまま返し、実行時にエラーを報告させます。
func ResolveExecutable(command string, candidates ...string) string {
	if path, err := exec.LookPath(command); err == nil {
		return path
	}
	for _, candidate := range candidates {
		if candidate != "" && fileExists(candidate) {
			return candidate
		}
	}
	return command
}

// StartCommand はコマンドを非同期で起動します。終了は待ちません。
func StartCommand(path string, args []string) error {
	return exec.Command(path, args...).Start()
}

// ExecDir は実行ファイルのディレクトリを返します。
func ExecDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

// fileExists はパスがディレクトリ以外のファイルとして存在するかを返します。
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
