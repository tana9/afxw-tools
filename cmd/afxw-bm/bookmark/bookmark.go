package bookmark

import (
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/sys/windows"

	"github.com/tana9/afxw-tools/internal/singleinstance"
	"github.com/tana9/afxw-tools/internal/stringutil"
)

var addMu sync.Mutex

// lockTimeoutMs はブックマークファイルの排他ロック取得を待つ最大時間（ミリ秒）です。
const lockTimeoutMs uint32 = 3000

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
// 既存エントリと大文字小文字を無視して重複する場合はfalseを返します。
// 新規に追加した場合はtrueを返します。
// -a パスは複数プロセスから同時実行され得るため、プロセス間ロックで
// 読み込み～追記の間の競合（重複追記）を防ぎます。
func Add(path string, newItem string) (added bool, err error) {
	addMu.Lock()
	defer addMu.Unlock()

	lock, err := acquireFileLock(path)
	if err != nil {
		return false, err
	}
	defer releaseFileLock(lock)

	// Windowsでの一貫性のため、パス区切り文字をバックスラッシュに正規化します
	newItem = filepath.Clean(newItem)

	data, err := readAll(path)
	if err != nil {
		return false, err
	}

	newKey := normKey(newItem)
	for _, line := range parseLines(data) {
		if normKey(line) == newKey { // Windowsパスは大文字小文字を区別しない
			return false, nil
		}
	}

	entry := newItem + "\n"
	// 手編集などで末尾改行が無い場合に既存の最終行と結合しないようにする
	if data != "" && !strings.HasSuffix(data, "\n") {
		entry = "\n" + entry
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return false, fmt.Errorf("追記用ブックマークファイルのオープンに失敗しました: %w", err)
	}

	if _, err := f.WriteString(entry); err != nil {
		_ = f.Close()
		return false, fmt.Errorf("ブックマークファイルへの書き込みに失敗しました: %w", err)
	}
	if err := f.Close(); err != nil {
		return false, fmt.Errorf("ブックマークファイルのクローズに失敗しました: %w", err)
	}

	return true, nil
}

// acquireFileLock はブックマークファイルパスに対応する名前付きミューテックスを取得します。
// 複数プロセスがAddを同時実行した場合でも、読み込みから追記までを排他区間として扱うためのロックです。
func acquireFileLock(path string) (windows.Handle, error) {
	h, err := singleinstance.AcquireHandle(mutexNameForPath(path), lockTimeoutMs)
	if err != nil {
		if errors.Is(err, singleinstance.ErrTimeout) {
			return 0, fmt.Errorf("ブックマークファイルのロック取得がタイムアウトしました (%s): %w", path, err)
		}
		return 0, fmt.Errorf("ブックマークファイルのロック取得に失敗しました (%s): %w", path, err)
	}
	return h, nil
}

// releaseFileLock は acquireFileLock で取得したロックを解放します。
func releaseFileLock(h windows.Handle) {
	singleinstance.Release(h)
}

// mutexNameForPath はファイルパスから名前付きミューテックス名を生成します。
// パスをそのまま使うとミューテックス名として不正な記号や長さ制限超過があり得るため、
// 絶対パスを小文字化したうえでハッシュ化し、固定長の名前にします。
func mutexNameForPath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	abs = strings.ToLower(filepath.Clean(abs))

	h := fnv.New64a()
	_, _ = h.Write([]byte(abs))
	return fmt.Sprintf("afxw-bm-%016x", h.Sum64())
}
