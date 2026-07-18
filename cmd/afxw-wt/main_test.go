package main

import (
	"errors"
	"testing"
)

func TestRun_Normal(t *testing.T) {
	var gotPath string
	var gotArgs []string
	start := func(path string, args []string) error {
		gotPath = path
		gotArgs = args
		return nil
	}

	if err := run(`C:\Users\Test\`, "wt.exe", start); err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	if gotPath != "wt.exe" {
		t.Errorf("期待: wt.exe, 取得: %s", gotPath)
	}
	if len(gotArgs) != 2 || gotArgs[0] != "-d" || gotArgs[1] != `C:\Users\Test\` {
		t.Errorf("予期しない引数: %v", gotArgs)
	}
}

func TestRun_EmptyDir(t *testing.T) {
	called := false
	start := func(path string, args []string) error {
		called = true
		return nil
	}

	if err := run("", "wt.exe", start); err == nil {
		t.Error("エラーが期待されましたが、nilが返りました")
	}
	if called {
		t.Error("空パスではWindows Terminalを起動すべきではありません")
	}
}

func TestRun_StartError(t *testing.T) {
	start := func(path string, args []string) error {
		return errors.New("start error")
	}

	err := run(`C:\Users\Test`, "wt.exe", start)
	if err == nil {
		t.Fatal("エラーが期待されましたが、nilが返りました")
	}
	if err.Error() != "wt.exeの起動に失敗しました: start error" {
		t.Errorf("予期しないエラーメッセージ: %v", err)
	}
}
