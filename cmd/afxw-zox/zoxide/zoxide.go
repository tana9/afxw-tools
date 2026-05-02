package zoxide

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type Entry struct {
	Path  string
	Score float64
}

// Query はzoxideのクエリを実行してスコア降順のエントリを返します。
func Query() ([]Entry, error) {
	cmd := exec.Command("zoxide", "query", "--list", "--score")
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("zoxideコマンドの実行に失敗しました: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("zoxideコマンドの実行に失敗しました: %w", err)
	}

	return parseQueryOutput(string(output))
}

// parseQueryOutput はzoxide query --list --scoreの出力（"スコア パス"形式）をパースします。
func parseQueryOutput(output string) ([]Entry, error) {
	var raw []Entry
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}
		score, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			continue
		}
		raw = append(raw, Entry{Path: parts[1], Score: score})
	}

	exists := make([]bool, len(raw))
	var wg sync.WaitGroup
	for i, e := range raw {
		wg.Add(1)
		go func(idx int, path string) {
			defer wg.Done()
			_, err := os.Stat(path)
			exists[idx] = err == nil
		}(i, e.Path)
	}
	wg.Wait()

	entries := make([]Entry, 0, len(raw))
	for i, e := range raw {
		if exists[i] {
			entries = append(entries, e)
		}
	}
	return entries, nil
}

func Paths(entries []Entry) []string {
	paths := make([]string, len(entries))
	for i, entry := range entries {
		paths[i] = entry.Path
	}
	return paths
}
