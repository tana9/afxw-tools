package afx

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

const (
	// WindowLeft はあふの左窓を表します。
	WindowLeft = 0
	// WindowRight はあふの右窓を表します。
	WindowRight = 1
)

// AFX は afxw.obj と対話するためのインターフェースを提供します。
type AFX interface {
	Histories(wins []int) ([]string, error)
	EXCD(path string) error
	GetActivePath() (string, error)
	Close()
}

type oleAFX struct {
	afxw    *ole.IDispatch
	unknown *ole.IUnknown
}

// NewOleAFX は実際の afxw.obj と対話する新しい AFX インスタンスを作成します。
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

// Histories は指定されたウィンドウの履歴ディレクトリを取得します。
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

// getWindowHistories は指定されたウィンドウの履歴ディレクトリ一覧を取得します。
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

// EXCD は指定されたパスにディレクトリを変更します。
func (a *oleAFX) EXCD(path string) error {

	normalizedPath := ensureTrailingBackslash(path)

	_, err := oleutil.CallMethod(a.afxw, "Exec", fmt.Sprintf("&EXCD -P\"%s\"", normalizedPath))
	if err != nil {
		return fmt.Errorf("EXCD呼び出しに失敗しました: %w", err)
	}
	return nil
}

// GetActivePath はアクティブウィンドウのカレントディレクトリを取得します。
func (a *oleAFX) GetActivePath() (string, error) {
	// $P はアクティブウィンドウのカレントディレクトリに展開されます
	res, err := oleutil.CallMethod(a.afxw, "Extract", "$P")
	if err != nil {
		return "", fmt.Errorf("アクティブパスの取得に失敗しました: %w", err)
	}
	path := fmt.Sprint(res.Value())
	res.Clear()
	return path, nil
}

// Close はCOMリソースを解放し、OSスレッドのロックを解除します。
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

// ensureTrailingBackslash は指定されたパスの末尾にバックスラッシュを追加します（既にある場合は追加しません）。
func ensureTrailingBackslash(path string) string {
	if !strings.HasSuffix(path, "\\") {
		return path + "\\"
	}
	return path
}
