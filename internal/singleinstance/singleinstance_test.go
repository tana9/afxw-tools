package singleinstance

import (
	"errors"
	"runtime"
	"testing"
)

func TestAcquire_FirstInstance(t *testing.T) {
	err := acquire("test-afxw-"+t.Name(), 0)
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
}

func TestAcquire_Timeout(t *testing.T) {
	name := "test-afxw-" + t.Name()

	held := make(chan struct{})
	release := make(chan struct{})

	// 別OSスレッドでミューテックスを保持し続けるゴルーチン
	go func() {
		runtime.LockOSThread()
		if err := acquire(name, 0); err != nil {
			t.Errorf("ミューテックスの取得に失敗しました: %v", err)
			close(held)
			return
		}
		close(held)
		<-release
		// ゴルーチン終了でOSスレッドも終了し、ミューテックスが放棄される
	}()

	<-held
	defer close(release)

	// 短いタイムアウトで取得を試みる → ErrTimeout になることを確認
	err := acquire(name, 200)
	if !errors.Is(err, ErrTimeout) {
		t.Errorf("ErrTimeout を期待しましたが、取得: %v", err)
	}
}

func TestAcquire_PreviousExited(t *testing.T) {
	name := "test-afxw-" + t.Name()

	exited := make(chan struct{})

	// ミューテックスを取得してすぐ終了するゴルーチン
	go func() {
		runtime.LockOSThread()
		acquire(name, 0)
		close(exited)
		// ゴルーチン終了でミューテックスが放棄される
	}()

	<-exited

	// 前のインスタンスが終了済みなので取得できることを確認
	err := acquire(name, 1000)
	if err != nil {
		t.Errorf("前のインスタンス終了後は取得できるべきですが、エラー: %v", err)
	}
}
