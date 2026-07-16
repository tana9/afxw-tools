package bookmark

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tana9/afxw-tools/internal/stringutil"
)

func GetDefaultPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exe), "bookmarks.txt"), nil
}

// Load はブックマークファイルを読み込み、正規化と重複除去を行ったパス一覧を返します。
// ファイルが存在しない場合は空のスライスを返します。
func Load(path string) ([]string, error) {
	data, err := readAll(path)
	if err != nil {
		return nil, err
	}
	return parseLines(data), nil
}

// readAll はブックマークファイルの内容を文字列で返します。ファイルが無い場合は空文字列を返します。
func readAll(path string) (string, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("ブックマークファイルの読み込みに失敗しました: %w", err)
	}
	return string(data), nil
}

// normKey は重複判定用の正規化キーを返します。
// Windowsパスは大文字小文字を区別しないため小文字化して比較します。
// Loadの重複除去とAddの重複チェックが必ず同じ規則になるよう、比較はすべてこの関数を経由させます。
func normKey(path string) string {
	return strings.ToLower(path)
}

// parseLines は内容を行単位に分割し、filepath.Cleanで正規化したうえで
// 大文字小文字を無視して重複除去します（Windowsパスは大文字小文字を区別しないため）。
func parseLines(data string) []string {
	lines := make([]string, 0)
	for line := range strings.SplitSeq(data, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, filepath.Clean(line))
		}
	}
	return stringutil.RemoveDuplicatesBy(lines, normKey)
}

// Add はパスを正規化してブックマークファイルへ追記します。
// 既存エントリと大文字小文字を無視して重複する場合は何もしません。
func Add(path string, newItem string) error {
	// Windowsでの一貫性のため、パス区切り文字をバックスラッシュに正規化します
	newItem = filepath.Clean(newItem)

	data, err := readAll(path)
	if err != nil {
		return err
	}

	newKey := normKey(newItem)
	for _, line := range parseLines(data) {
		if normKey(line) == newKey { // Windowsパスは大文字小文字を区別しない
			return nil
		}
	}

	entry := newItem + "\n"
	// 手編集などで末尾改行が無い場合に既存の最終行と結合しないようにする
	if data != "" && !strings.HasSuffix(data, "\n") {
		entry = "\n" + entry
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("追記用ブックマークファイルのオープンに失敗しました: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("ブックマークファイルへの書き込みに失敗しました: %w", err)
	}

	return nil
}
