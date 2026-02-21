package singleinstance

import (
	"errors"

	"golang.org/x/sys/windows"
)

// ErrAlreadyRunning は別インスタンスがすでに起動していることを示します。
var ErrAlreadyRunning = errors.New("すでに起動しています")

// Acquire は名前付きミューテックスを取得します。
// すでに別インスタンスが起動中の場合は ErrAlreadyRunning を返します。
// 取得したミューテックスはプロセス終了時に自動的に解放されます。
func Acquire(name string) error {
	_, err := windows.CreateMutex(nil, false, windows.StringToUTF16Ptr("Local\\"+name))
	if err == windows.ERROR_ALREADY_EXISTS {
		return ErrAlreadyRunning
	}
	return err
}
