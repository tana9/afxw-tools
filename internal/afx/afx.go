package afx

import (
	"fmt"
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

type oleAFX struct {
	afxw    *ole.IDispatch
	unknown *ole.IUnknown
}

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
// afxw.obj側の実装差異でVARIANTのサブタイプ(int16/int32/int64等)が変わってもpanicしないようにするための変換です。
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
	default:
		return 0, fmt.Errorf("予期しない型です (%T)", v)
	}
}

func (a *oleAFX) EXCD(path string) error {
	normalizedPath := ensureTrailingBackslash(path)
	return a.execEXCD(normalizedPath)
}

// EXCDFile は対象ファイルのフォルダへ移動し、そのファイルにカーソルを合わせます。
func (a *oleAFX) EXCDFile(path string) error {
	return a.execEXCD(path)
}

func (a *oleAFX) execEXCD(path string) error {
	_, err := oleutil.CallMethod(a.afxw, "Exec", fmt.Sprintf("&EXCD -P\"%s\"", path))
	if err != nil {
		return fmt.Errorf("EXCD呼び出しに失敗しました: %w", err)
	}
	return nil
}

func (a *oleAFX) GetActivePath() (string, error) {
	// $P はアクティブウィンドウのカレントディレクトリに展開されます
	return a.extract("$P")
}

func (a *oleAFX) GetCurrentFile() (string, error) {
	// $P はカレントディレクトリ（末尾に \ あり）、$F はカーソル上のファイル名
	dir, err := a.extract("$P")
	if err != nil {
		return "", err
	}
	name, err := a.extract("$F")
	if err != nil {
		return "", err
	}
	return dir + name, nil
}

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

func parseMarkedFiles(result string, getCurrentFile func() (string, error)) ([]string, error) {
	result = strings.TrimSpace(result)
	if result == "" {
		f, err := getCurrentFile()
		if err != nil {
			return nil, err
		}
		return []string{f}, nil
	}
	lines := strings.Split(strings.ReplaceAll(result, "\r\n", "\n"), "\n")
	files := make([]string, 0, len(lines))
	for _, line := range lines {
		path := strings.TrimSpace(line)
		path = strings.Trim(path, "\"")
		if path != "" {
			files = append(files, path)
		}
	}
	return files, nil
}

func (a *oleAFX) extract(variable string) (string, error) {
	res, err := oleutil.CallMethod(a.afxw, "Extract", variable)
	if err != nil {
		return "", fmt.Errorf("変数の展開に失敗しました (%s): %w", variable, err)
	}
	value := fmt.Sprint(res.Value())
	_ = res.Clear()
	return value, nil
}

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

func ensureTrailingBackslash(path string) string {
	if !strings.HasSuffix(path, "\\") {
		return path + "\\"
	}
	return path
}
