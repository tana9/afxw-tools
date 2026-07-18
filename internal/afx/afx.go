package afx

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

const (
	WindowLeft  = 0
	WindowRight = 1
)

// Client は afxw.obj と対話するためのインターフェースです。
type Client interface {
	DirectoryHistories(windows []int) ([]string, error)
	ChangeDirectory(path string) error
	ActivePath() (string, error)
	CurrentFile() (string, error)
	// MarkedFiles はマーク済みファイルを返し、マークがなければカーソル位置のファイルを返します。
	MarkedFiles() ([]string, error)
	Close()
}

type oleClient struct {
	afxw    *ole.IDispatch
	unknown *ole.IUnknown
}

// callCOMMethod はCOM境界を単体テストで差し替え可能にするための関数変数です。
var callCOMMethod = oleutil.CallMethod

func NewOLEClient() (Client, error) {
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
	return &oleClient{afxw: afxw, unknown: unknown}, nil
}

func (a *oleClient) DirectoryHistories(windows []int) ([]string, error) {
	var dirs []string
	for _, window := range windows {
		windowDirs, err := a.windowDirectoryHistories(window)
		if err != nil {
			return nil, err
		}
		dirs = append(dirs, windowDirs...)
	}
	return dirs, nil
}

func (a *oleClient) windowDirectoryHistories(window int) ([]string, error) {
	res, err := callCOMMethod(a.afxw, "HisDirCount", window)
	if err != nil {
		return nil, fmt.Errorf("履歴件数の取得に失敗しました: %w", err)
	}
	count, err := toInt(res.Value())
	res.Clear()
	if err != nil {
		return nil, fmt.Errorf("履歴件数の取得に失敗しました: %w", err)
	}

	dirs := make([]string, 0, count)
	for i := range count {
		res, err := callCOMMethod(a.afxw, "HisDir", window, i)
		if err != nil {
			return nil, fmt.Errorf("履歴の取得に失敗しました: %w", err)
		}
		dirs = append(dirs, fmt.Sprint(res.Value()))
		res.Clear()
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

func (a *oleClient) ChangeDirectory(path string) error {
	normalizedPath := ensureTrailingBackslash(path)
	res, err := callCOMMethod(a.afxw, "Exec", fmt.Sprintf("&EXCD -P\"%s\"", normalizedPath))
	if err != nil {
		return fmt.Errorf("EXCD呼び出しに失敗しました: %w", err)
	}
	res.Clear()
	return nil
}

func (a *oleClient) ActivePath() (string, error) {
	// $P はアクティブウィンドウのカレントディレクトリに展開されます
	return a.extract("$P")
}

func (a *oleClient) CurrentFile() (string, error) {
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

func (a *oleClient) MarkedFiles() ([]string, error) {
	// $JU はマークファイル間の区切りをLFにし、$QN は各パスの引用符を外す。
	// 空白区切りでは空白を含むパスを判別できないため、行単位で取得する。
	result, err := a.extract("$JU$QN$MF")
	if err != nil {
		return nil, err
	}
	return parseMarkedFiles(result, a.CurrentFile)
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
	files := make([]string, 0, strings.Count(result, "\n")+1)
	for line := range strings.SplitSeq(result, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func (a *oleClient) extract(variable string) (string, error) {
	res, err := callCOMMethod(a.afxw, "Extract", variable)
	if err != nil {
		return "", fmt.Errorf("変数の展開に失敗しました (%s): %w", variable, err)
	}
	value := fmt.Sprint(res.Value())
	res.Clear()
	return value, nil
}

func (a *oleClient) Close() {
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
