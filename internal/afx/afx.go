package afx

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

const (
	WindowLeft       = 0
	WindowRight      = 1
	markedFilesMacro = "$JU$MF"
)

// AFX は afxw.obj と対話するためのインターフェースです。
type AFX interface {
	Histories(wins []int) ([]string, error)
	EXCD(path string) error
	EXCDFile(path string) error
	GetActivePath() (string, error)
	GetCurrentFile() (string, error)
	// マークなし時はカーソルファイルを返す。
	GetMarkedFiles() ([]string, error)
	Close()
}

// oleAFX は afxw.obj のCOMオブジェクトとその IUnknown を保持する AFX の実装です。
type oleAFX struct {
	afxw    *ole.IDispatch
	unknown *ole.IUnknown
}

// NewOleAFX は afxw.obj に接続して AFX を返します。COMの初期化とOSスレッドのロックを行い、失敗時は後始末してから返します。
func NewOleAFX() (AFX, error) {
	runtime.LockOSThread()
	success := false
	defer func() {
		if !success {
			runtime.UnlockOSThread()
		}
	}()

	if err := ole.CoInitialize(0); err != nil {
		return nil, fmt.Errorf("COMの初期化に失敗しました: %w", err)
	}
	defer func() {
		if !success {
			ole.CoUninitialize()
		}
	}()

	unknown, err := oleutil.CreateObject("afxw.obj")
	if err != nil {
		return nil, fmt.Errorf("afxw.objの作成に失敗しました: %w", err)
	}
	defer func() {
		if !success {
			unknown.Release()
		}
	}()

	afxw, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return nil, fmt.Errorf("IDispatchの取得に失敗しました: %w", err)
	}

	success = true
	return &oleAFX{afxw: afxw, unknown: unknown}, nil
}

// Connect はあふwへ接続し、失敗時は日本語コンテキスト付きのエラーを返します。
// 各コマンドで重複していた接続エラーのラップを一元化するためのヘルパーです。
func Connect() (AFX, error) {
	a, err := NewOleAFX()
	if err != nil {
		return nil, fmt.Errorf("afxw.objへの接続に失敗しました（あふwが起動しているか確認してください）: %w", err)
	}
	return a, nil
}

// Histories は指定したウィンドウ番号のフォルダ履歴をまとめて返します。
func (a *oleAFX) Histories(wins []int) ([]string, error) {
	var dirs []string
	for _, win := range wins {
		winDirs, err := a.getWindowHistories(win)
		if err != nil {
			return nil, err
		}
		dirs = append(dirs, winDirs...)
	}
	return dirs, nil
}

// getWindowHistories は1つのウィンドウのフォルダ履歴を件数分取得して返します。
func (a *oleAFX) getWindowHistories(win int) ([]string, error) {
	res, err := oleutil.CallMethod(a.afxw, "HisDirCount", win)
	if err != nil {
		return nil, fmt.Errorf("履歴件数の取得に失敗しました: %w", err)
	}
	count, err := toInt(res.Value())
	_ = res.Clear()
	if err != nil {
		return nil, fmt.Errorf("履歴件数の取得に失敗しました: %w", err)
	}

	dirs := make([]string, 0, count)
	for i := range count {
		res, err := oleutil.CallMethod(a.afxw, "HisDir", win, i)
		if err != nil {
			return nil, fmt.Errorf("履歴の取得に失敗しました: %w", err)
		}
		dirs = append(dirs, fmt.Sprint(res.Value()))
		_ = res.Clear()
	}
	return dirs, nil
}

// toInt はCOMのVARIANTから返る整数値を型を問わずintに変換します。
// afxw.obj側の実装差異でVARIANTのサブタイプ(int16/int32/int64や符号なし系等)が変わってもpanicしないようにするための変換です。
func toInt(v any) (int, error) {
	switch n := v.(type) {
	case int:
		return n, nil
	case int16:
		return int(n), nil
	case int32:
		return int(n), nil
	case int64:
		return int(n), nil
	case uint16:
		return int(n), nil
	case uint32:
		return int(n), nil
	case uint64:
		return int(n), nil
	default:
		return 0, fmt.Errorf("予期しない型です (%T)", v)
	}
}

// EXCD は指定したフォルダへ移動します。末尾に \ が無ければ補って実行します。
func (a *oleAFX) EXCD(path string) error {
	normalizedPath := ensureTrailingBackslash(path)
	return a.execEXCD(normalizedPath)
}

// EXCDFile は対象ファイルのフォルダへ移動し、そのファイルにカーソルを合わせます。
func (a *oleAFX) EXCDFile(path string) error {
	return a.execEXCD(path)
}

// execEXCD はあふwの &EXCD コマンドを実行して指定パスへ移動させます。
func (a *oleAFX) execEXCD(path string) error {
	res, err := oleutil.CallMethod(a.afxw, "Exec", fmt.Sprintf("&EXCD -P\"%s\"", path))
	if err != nil {
		return fmt.Errorf("EXCD呼び出しに失敗しました: %w", err)
	}
	_ = res.Clear()
	return nil
}

// GetActivePath はアクティブウィンドウのカレントディレクトリのパスを返します。
func (a *oleAFX) GetActivePath() (string, error) {
	// $P はアクティブウィンドウのカレントディレクトリに展開されます
	return a.extract("$P")
}

// GetCurrentFile はカーソル位置のファイルのフルパスを返します。
func (a *oleAFX) GetCurrentFile() (string, error) {
	// $P はカレントディレクトリ（末尾に \ なし）、$F はカーソル上のファイル名。
	// 区切りの \ を補って1回のExtractで取得し、ルートでの \ の重複はCleanで正規化する。
	path, err := a.extract(`$P\$F`)
	if err != nil {
		return "", err
	}
	return filepath.Clean(path), nil
}

// GetMarkedFiles はマークされたファイルのフルパス一覧を返します。マークが無い場合はカーソル位置のファイルを返します。
func (a *oleAFX) GetMarkedFiles() ([]string, error) {
	return getMarkedFiles(a.extract, a.GetCurrentFile)
}

// getMarkedFiles はマーク項目を改行区切りで展開してパス一覧に変換します。
func getMarkedFiles(extract func(string) (string, error), getCurrentFile func() (string, error)) ([]string, error) {
	// $JU はマーク項目の区切りを改行にするため、空白を含むパスも安全に分離できる。
	result, err := extract(markedFilesMacro)
	if err != nil {
		return nil, err
	}
	return parseMarkedFiles(result, getCurrentFile)
}

// parseMarkedFiles は $JU の展開結果を行ごとに分割してパス一覧にします。結果が空の場合はカーソル位置のファイルを返します。
func parseMarkedFiles(result string, getCurrentFile func() (string, error)) ([]string, error) {
	result = strings.TrimSpace(result)
	if result == "" {
		f, err := getCurrentFile()
		if err != nil {
			return nil, err
		}
		return []string{f}, nil
	}
	var files []string
	for line := range strings.SplitSeq(result, "\n") {
		// CRLF区切りの \r もTrimSpaceで取り除かれる
		path := strings.TrimSpace(line)
		path = strings.Trim(path, "\"")
		if path != "" {
			files = append(files, path)
		}
	}
	return files, nil
}

// extract はあふwの Extract メソッドで変数マクロを展開した文字列を返します。
func (a *oleAFX) extract(variable string) (string, error) {
	res, err := oleutil.CallMethod(a.afxw, "Extract", variable)
	if err != nil {
		return "", fmt.Errorf("変数の展開に失敗しました (%s): %w", variable, err)
	}
	value := fmt.Sprint(res.Value())
	_ = res.Clear()
	return value, nil
}

// Close は afxw.obj の参照を解放し、COMの終了とOSスレッドのロック解除を行います。
func (a *oleAFX) Close() {
	defer runtime.UnlockOSThread()
	defer ole.CoUninitialize()

	if a.afxw != nil {
		a.afxw.Release()
	}
	if a.unknown != nil {
		a.unknown.Release()
	}
}

// ensureTrailingBackslash はパス末尾に \ が無い場合に \ を付加して返します。
func ensureTrailingBackslash(path string) string {
	if !strings.HasSuffix(path, "\\") {
		return path + "\\"
	}
	return path
}
