package singleinstance

import (
	"errors"
	"testing"
)

func TestAcquire_FirstInstance(t *testing.T) {
	err := Acquire("test-afxw-singleinstance-" + t.Name())
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
}

func TestAcquire_SecondInstance(t *testing.T) {
	name := "test-afxw-singleinstance-" + t.Name()

	if err := Acquire(name); err != nil {
		t.Fatalf("1回目の取得に失敗しました: %v", err)
	}

	if err := Acquire(name); !errors.Is(err, ErrAlreadyRunning) {
		t.Errorf("ErrAlreadyRunning を期待しましたが、取得: %v", err)
	}
}
