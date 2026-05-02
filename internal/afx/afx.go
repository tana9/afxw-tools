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

// AFX は afxw.obj と対話するためのインターフェースです。
type AFX interface {
	Histories(wins []int) ([]string, error)
	EXCD(path string) error
	GetActivePath() (string, error)
	GetCurrentFile() (string, error)
	// マークなし時はカーソルファイルを返す。パスにスペースを含む場合は正しく動作しない。
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
	count := res.Value().(int32)
	res.Clear()

	dirs := make([]string, 0, count)
	for i := 0; i < int(count); i++ {
		res, err := oleutil.CallMethod(a.afxw, "HisDir", win, i)
		if err != nil {
			return nil, fmt.Errorf("履歴の取得に失敗しました: %w", err)
		}
		dirs = append(dirs, fmt.Sprint(res.Value()))
		res.Clear()
	}
	return dirs, nil
}

func (a *oleAFX) EXCD(path string) error {
	normalizedPath := ensureTrailingBackslash(path)
	_, err := oleutil.CallMethod(a.afxw, "Exec", fmt.Sprintf("&EXCD -P\"%s\"", normalizedPath))
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
	// $MFP はスペース区切りで返されるため、パスにスペースが含まれる場合は正しく動作しない
	result, err := a.extract("$MFP")
	if err != nil {
		return nil, err
	}
	result = strings.TrimSpace(result)
	if result == "" {
		f, err := a.GetCurrentFile()
		if err != nil {
			return nil, err
		}
		return []string{f}, nil
	}
	return strings.Fields(result), nil
}

func (a *oleAFX) extract(variable string) (string, error) {
	res, err := oleutil.CallMethod(a.afxw, "Extract", variable)
	if err != nil {
		return "", fmt.Errorf("変数の展開に失敗しました (%s): %w", variable, err)
	}
	value := fmt.Sprint(res.Value())
	res.Clear()
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
