// Package search はripgrepを使った文字コード混在ファイルの検索を提供します。
package search

import (
	"bufio"
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"
)

const (
	EncodingAuto     = "auto"
	EncodingUTF8     = "utf-8"
	EncodingUTF8BOM  = "utf-8bom"
	EncodingSJIS     = "sjis"
	EncodingShiftJIS = "shift_jis"
	maxBatchSize     = 20_000
	// classifySampleSize は文字コード判定でファイル先頭から読む上限バイト数。
	// ファイル全体を読むと大きなツリーでripgrep本体より判定処理が支配的になるため、サンプリングに留める。
	classifySampleSize = 64 * 1024
)

// Options は検索条件を表します。
type Options struct {
	Root         string
	Pattern      string
	Encoding     string
	Glob         string
	FixedStrings bool
	Hidden       bool
}

// Match はripgrepの1件の検索結果を表します。
type Match struct {
	Path string
	Text string
	Line int
}

// Run は指定したフォルダ以下をripgrepで検索します。
func Run(ctx context.Context, executable string, opts Options) ([]Match, error) {
	encoding, err := normalizeEncoding(opts.Encoding)
	if err != nil {
		return nil, err
	}
	if err := validateOptions(opts); err != nil {
		return nil, err
	}
	if encoding != EncodingAuto {
		return runSearch(ctx, executable, opts, encoding, []string{"."})
	}

	files, err := listFiles(ctx, executable, opts)
	if err != nil {
		return nil, err
	}
	utf8Files, sjisFiles, err := classifyFiles(opts.Root, files)
	if err != nil {
		return nil, err
	}

	var matches []Match
	for _, group := range []struct {
		encoding string
		files    []string
	}{{EncodingUTF8, utf8Files}, {EncodingShiftJIS, sjisFiles}} {
		for _, batch := range batches(group.files, maxBatchSize) {
			found, err := runSearch(ctx, executable, opts, group.encoding, batch)
			if err != nil {
				return nil, err
			}
			matches = append(matches, found...)
		}
	}
	slices.SortFunc(matches, func(a, b Match) int {
		if c := cmp.Compare(a.Path, b.Path); c != 0 {
			return c
		}
		return cmp.Compare(a.Line, b.Line)
	})
	return matches, nil
}

// validateOptions は検索条件を検証します。
func validateOptions(opts Options) error {
	if strings.TrimSpace(opts.Root) == "" {
		return fmt.Errorf("検索対象フォルダが空です")
	}
	if strings.TrimSpace(opts.Pattern) == "" {
		return fmt.Errorf("検索キーワードが空です")
	}
	_, err := normalizeEncoding(opts.Encoding)
	return err
}

// normalizeEncoding は利用者向けの文字コード名をripgrepが受け付ける値へ変換します。
func normalizeEncoding(encoding string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case EncodingAuto:
		return EncodingAuto, nil
	case EncodingUTF8, EncodingUTF8BOM:
		return EncodingUTF8, nil
	case EncodingSJIS, EncodingShiftJIS:
		return EncodingShiftJIS, nil
	default:
		return "", fmt.Errorf("未対応の文字コードです: %s", encoding)
	}
}

// listFiles はripgrepのフィルタ規則に従って検索対象ファイルを列挙します。
func listFiles(ctx context.Context, executable string, opts Options) ([]string, error) {
	args := []string{"--files", "--null"}
	if opts.Hidden {
		args = append(args, "--hidden")
	}
	if opts.Glob != "" {
		args = append(args, "--glob", opts.Glob)
	}
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = opts.Root
	output, err := cmd.Output()
	if err != nil {
		return nil, commandError("検索対象ファイルの列挙", err)
	}

	var files []string
	for item := range bytes.SplitSeq(output, []byte{0}) {
		if len(item) > 0 {
			files = append(files, string(item))
		}
	}
	return files, nil
}

// classifyFiles はUTF-8として妥当なファイルとShift_JIS候補へ排他的に分類します。
func classifyFiles(root string, files []string) ([]string, []string, error) {
	utf8Files := make([]string, 0, len(files))
	sjisFiles := make([]string, 0)
	for _, path := range files {
		valid, err := isUTF8File(filepath.Join(root, path))
		if err != nil {
			return nil, nil, fmt.Errorf("文字コードの判定に失敗しました (%s): %w", path, err)
		}
		if valid {
			utf8Files = append(utf8Files, path)
		} else {
			sjisFiles = append(sjisFiles, path)
		}
	}
	return utf8Files, sjisFiles, nil
}

// isUTF8File はファイル先頭 classifySampleSize バイトが妥当なUTF-8かをストリーミングで判定します。
// サンプル内に不正なバイト列が無ければUTF-8として扱います(先頭で判定できない稀なケースはrg側の検索結果が乱れる程度に留まる)。
func isUTF8File(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() { _ = f.Close() }()

	r := bufio.NewReader(f)
	read := 0
	for read < classifySampleSize {
		rn, size, err := r.ReadRune()
		if errors.Is(err, io.EOF) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		if rn == utf8.RuneError && size == 1 {
			return false, nil
		}
		read += size
	}
	return true, nil
}

// batches はWindowsのコマンドライン長を超えないようファイル一覧を分割します。
func batches(files []string, maxSize int) [][]string {
	var result [][]string
	var batch []string
	size := 0
	for _, file := range files {
		cost := len(file) + 3
		if len(batch) > 0 && size+cost > maxSize {
			result = append(result, batch)
			batch = nil
			size = 0
		}
		batch = append(batch, file)
		size += cost
	}
	if len(batch) > 0 {
		result = append(result, batch)
	}
	return result
}

// runSearch は同じ文字コードのファイル群をripgrepで検索します。
func runSearch(ctx context.Context, executable string, opts Options, encoding string, paths []string) ([]Match, error) {
	args := []string{"--json", "--color", "never", "--encoding", encoding}
	if opts.FixedStrings {
		args = append(args, "--fixed-strings")
	}
	if opts.Hidden {
		args = append(args, "--hidden")
	}
	if opts.Glob != "" {
		args = append(args, "--glob", opts.Glob)
	}
	args = append(args, "-e", opts.Pattern, "--")
	args = append(args, paths...)

	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = opts.Root
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, commandError("ripgrep検索", err)
	}
	return parseJSON(opts.Root, output)
}

// commandError は外部コマンドのエラーへ標準エラー出力を付加します。
func commandError(operation string, err error) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
		return fmt.Errorf("%sに失敗しました: %s", operation, strings.TrimSpace(string(exitErr.Stderr)))
	}
	return fmt.Errorf("%sに失敗しました: %w", operation, err)
}

type jsonEvent struct {
	Type string `json:"type"`
	Data struct {
		Path struct {
			Text string `json:"text"`
		} `json:"path"`
		Lines struct {
			Text string `json:"text"`
		} `json:"lines"`
		LineNumber int `json:"line_number"`
	} `json:"data"`
}

// parseJSON はripgrepのJSON Lines出力からマッチ結果を取り出します。
func parseJSON(root string, output []byte) ([]Match, error) {
	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	var matches []Match
	for scanner.Scan() {
		var event jsonEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, fmt.Errorf("ripgrep出力の解析に失敗しました: %w", err)
		}
		if event.Type != "match" || event.Data.Path.Text == "" {
			continue
		}
		path := event.Data.Path.Text
		if !filepath.IsAbs(path) {
			path = filepath.Join(root, path)
		}
		matches = append(matches, Match{
			Path: filepath.Clean(path),
			Line: event.Data.LineNumber,
			Text: strings.TrimSpace(event.Data.Lines.Text),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ripgrep出力の読み込みに失敗しました: %w", err)
	}
	return matches, nil
}

// Labels はfuzzy finderへ表示する検索結果の一覧を返します。
func Labels(root string, matches []Match) []string {
	labels := make([]string, len(matches))
	for i, match := range matches {
		path, err := filepath.Rel(root, match.Path)
		if err != nil {
			path = match.Path
		}
		labels[i] = fmt.Sprintf("%s:%d: %s", path, match.Line, match.Text)
	}
	return labels
}
