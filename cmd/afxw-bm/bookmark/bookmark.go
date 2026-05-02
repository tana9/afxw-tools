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

func Load(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ブックマークファイルの読み込みに失敗しました: %w", err)
	}

	var lines []string
	for line := range strings.SplitSeq(string(data), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	return stringutil.RemoveDuplicates(lines), nil
}

func Add(path string, newItem string) error {
	// Windowsでの一貫性のため、パス区切り文字をバックスラッシュに正規化します
	newItem = filepath.Clean(newItem)

	lines, err := Load(path)
	if err != nil {
		return err
	}

	for _, line := range lines {
		if strings.EqualFold(line, newItem) { // Windowsパスは大文字小文字を区別しない
			return nil
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
