package zoxide

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Entry はzoxideのディレクトリエントリを表します。
type Entry struct {
	Path  string  // ディレクトリパス
	Score float64 // frecencyスコア
}

// Query はzoxideのクエリコマンドを実行してディレクトリリストを取得します。
// スコアの高い順（降順）でソートされたエントリを返します。
func Query() ([]Entry, error) {
	// zoxide query --list --score を実行
	cmd := exec.Command("zoxide", "query", "--list", "--score")
	output, err := cmd.Output()
	if err != nil {
		// zoxideがインストールされていない、またはデータベースが空の場合
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("zoxideコマンドの実行に失敗しました: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("zoxideコマンドの実行に失敗しました: %w", err)
	}

	return parseQueryOutput(string(output))
}

// parseQueryOutput はzoxide query --list --scoreの出力をパースします。
// 出力形式: "スコア パス" (例: "12.5 C:\Users\TanakaTakashi\Projects")
func parseQueryOutput(output string) ([]Entry, error) {
	var entries []Entry
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// スコアとパスを分割（最初のスペースで分割）
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue // 不正な行はスキップ
		}

		score, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			continue // スコアのパースに失敗した行はスキップ
		}

		path := parts[1]

		// パスが実際に存在するか確認
		if _, err := os.Stat(path); err == nil {
			entries = append(entries, Entry{
				Path:  path,
				Score: score,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("出力のパースに失敗しました: %w", err)
	}

	return entries, nil
}

// Paths はエントリからパスのみを抽出して返します。
func Paths(entries []Entry) []string {
	paths := make([]string, len(entries))
	for i, entry := range entries {
		paths[i] = entry.Path
	}
	return paths
}
