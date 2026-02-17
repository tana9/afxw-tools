package bookmark

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetDefaultPath はブックマークファイルのデフォルトパスを返します。
// 実行ファイルと同じディレクトリにある "bookmarks.txt" のパスを返します。
func GetDefaultPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exe), "bookmarks.txt"), nil
}

// Load は指定されたファイルからブックマークを読み込みます。
// 重復のない文字列のスライスを返します。
func Load(path string) ([]string, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ブックマークファイルのオープンに失敗しました: %w", err)
	}
	defer f.Close()

	var lines []string
	seen := make(map[string]struct{})
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if _, ok := seen[line]; !ok {
			lines = append(lines, line)
			seen[line] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ブックマークファイルのスキャンに失敗しました: %w", err)
	}
	return lines, nil
}

// Add は新しいブックマークをファイルに追記します。
// 重複するブックマークは追加しません。
func Add(path string, newItem string) error {
	// Windowsでの一貫性のため、パス区切り文字をバックスラッシュに正規化します
	newItem = filepath.Clean(newItem)

	lines, err := Load(path)
	if err != nil {
		return err
	}

	for _, line := range lines {
		if strings.EqualFold(line, newItem) { // Windowsパスの大文字小文字を区別しない比較
			return nil // 既に存在する場合は何もしない
		}
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("追記用ブックマークファイルのオープンに失敗しました: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(newItem + "\n"); err != nil {
		return fmt.Errorf("ブックマークファイルへの書き込みに失敗しました: %w", err)
	}

	return nil
}
