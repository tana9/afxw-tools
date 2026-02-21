package singleinstance

import (
	"errors"
	"fmt"

	"golang.org/x/sys/windows"
)

// ErrTimeout は前のプロセスの終了待ちがタイムアウトしたことを示します。
var ErrTimeout = errors.New("起動中のプロセスが応答しません")

// defaultWaitMs は前のプロセスの終了を待つデフォルト時間（ミリ秒）です。
const defaultWaitMs uint32 = 3000

// WaitForSingleObject の戻り値定数（uint32）
const (
	waitObject0   uint32 = 0x00000000
	waitAbandoned uint32 = 0x00000080
	waitTimeout   uint32 = 0x00000102
)

// Acquire は名前付きミューテックスを取得します。
// すでに別インスタンスが起動中の場合は終了を最大 defaultWaitMs 待ちます。
// タイムアウトした場合は ErrTimeout を返します。
// 取得したミューテックスはプロセス終了時に自動的に解放されます。
func Acquire(name string) error {
	return acquire(name, defaultWaitMs)
}

func acquire(name string, timeoutMs uint32) error {
	h, err := windows.CreateMutex(nil, true, windows.StringToUTF16Ptr("Local\\"+name))
	if err == nil {
		// 新規作成成功 - プロセス終了まで保持（意図的なリーク）
		_ = h
		return nil
	}
	if err != windows.ERROR_ALREADY_EXISTS {
		return fmt.Errorf("ミューテックスの作成に失敗しました: %w", err)
	}

	// 別インスタンスが起動中 - 終了を待つ
	event, _ := windows.WaitForSingleObject(h, timeoutMs)
	switch event {
	case waitObject0, waitAbandoned:
		// 前のインスタンスが終了した - h を保持し続ける
		_ = h
		return nil
	case waitTimeout:
		windows.CloseHandle(h)
		return ErrTimeout
	default:
		windows.CloseHandle(h)
		return fmt.Errorf("ミューテックスの待機に失敗しました")
	}
}
