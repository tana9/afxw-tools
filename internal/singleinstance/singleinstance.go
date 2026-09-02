package singleinstance

import (
	"errors"
	"fmt"
	"runtime"

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
	h, err := acquireHandle(name, timeoutMs)
	if err != nil {
		return err
	}
	// h はプロセス終了まで保持する（意図的なリーク）
	_ = h
	return nil
}

// AcquireHandle は名前付きミューテックスを取得し、呼び出し側がReleaseで明示的に解放できるハンドルを返します。
// Acquireと異なり、短時間だけ排他したいユースケース（ファイルロック等）向けです。
// Windowsのミューテックスはスレッド単位で所有権を持つため、取得と解放を同一OSスレッドで行う必要があります。
// そのため取得成功時は呼び出し元のゴルーチンを現在のOSスレッドにロックします。必ずReleaseとペアで呼び出してください。
func AcquireHandle(name string, timeoutMs uint32) (windows.Handle, error) {
	runtime.LockOSThread()
	h, err := acquireHandle(name, timeoutMs)
	if err != nil {
		runtime.UnlockOSThread()
		return 0, err
	}
	return h, nil
}

// Release はAcquireHandleで取得したミューテックスを解放し、対応するOSスレッドロックも解除します。
func Release(h windows.Handle) {
	_ = windows.ReleaseMutex(h)
	_ = windows.CloseHandle(h)
	runtime.UnlockOSThread()
}

// acquireHandle は名前付きミューテックスを取得し、そのハンドルを返す共通処理です。
func acquireHandle(name string, timeoutMs uint32) (windows.Handle, error) {
	h, err := windows.CreateMutex(nil, true, windows.StringToUTF16Ptr("Local\\"+name))
	if err == nil {
		// 新規作成成功 - 呼び出し元が所有権を持つ
		return h, nil
	}
	if err != windows.ERROR_ALREADY_EXISTS {
		return 0, fmt.Errorf("ミューテックスの作成に失敗しました: %w", err)
	}

	// 別インスタンスが起動中 - 終了を待つ
	event, err := windows.WaitForSingleObject(h, timeoutMs)
	if err != nil {
		_ = windows.CloseHandle(h)
		return 0, fmt.Errorf("ミューテックスの待機に失敗しました: %w", err)
	}
	switch event {
	case waitObject0, waitAbandoned:
		// 前のインスタンスが終了した - 所有権を得た
		return h, nil
	case waitTimeout:
		_ = windows.CloseHandle(h)
		return 0, ErrTimeout
	default:
		_ = windows.CloseHandle(h)
		return 0, fmt.Errorf("ミューテックスの待機に失敗しました: 予期しない戻り値 %d", event)
	}
}
