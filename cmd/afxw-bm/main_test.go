package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/tana9/afxw-tools/internal/afxtest"
)

func TestRunSelect_Normal(t *testing.T) {
	tmpDir := t.TempDir()
	bmPath := filepath.Join(tmpDir, "bookmarks.txt")

	content := "C:\\Users\\Test\\Dir1\nC:\\Users\\Test\\Dir2\n"
	if err := os.WriteFile(bmPath, []byte(content), 0644); err != nil {
		t.Fatalf("テストファイル作成に失敗しました: %v", err)
	}

	afxMock := &afxtest.MockAFX{}
	finderMock := &afxtest.MockFinder{Idx: 1}

	if err := runSelect(afxMock, finderMock, bmPath); err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	if afxMock.ExcdPath != `C:\Users\Test\Dir2` {
		t.Errorf("期待: C:\\Users\\Test\\Dir2, 取得: %s", afxMock.ExcdPath)
	}
}

func TestRunSelect_EmptyBookmarks(t *testing.T) {
	tmpDir := t.TempDir()
	bmPath := filepath.Join(tmpDir, "bookmarks.txt")

	afxMock := &afxtest.MockAFX{}
	finderMock := &afxtest.MockFinder{}

	// ファイルなし（空のブックマーク）
	if err := runSelect(afxMock, finderMock, bmPath); err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	// finderもEXCDも呼ばれていないこと
	if afxMock.ExcdPath != "" {
		t.Errorf("EXCDが呼ばれるべきではありません: %s", afxMock.ExcdPath)
	}
}

func TestRunSelect_FinderCancelled(t *testing.T) {
	tmpDir := t.TempDir()
	bmPath := filepath.Join(tmpDir, "bookmarks.txt")

	if err := os.WriteFile(bmPath, []byte("C:\\Users\\Test\\Dir1\n"), 0644); err != nil {
		t.Fatalf("テストファイル作成に失敗しました: %v", err)
	}

	afxMock := &afxtest.MockAFX{}
	finderMock := &afxtest.MockFinder{Err: fuzzyfinder.ErrAbort}

	// キャンセルは正常終了
	if err := runSelect(afxMock, finderMock, bmPath); err != nil {
		t.Fatalf("キャンセルはエラーになるべきではありません: %v", err)
	}
}

func TestRunSelect_FinderError(t *testing.T) {
	tmpDir := t.TempDir()
	bmPath := filepath.Join(tmpDir, "bookmarks.txt")

	if err := os.WriteFile(bmPath, []byte("C:\\Users\\Test\\Dir1\n"), 0644); err != nil {
		t.Fatalf("テストファイル作成に失敗しました: %v", err)
	}

	afxMock := &afxtest.MockAFX{}
	finderMock := &afxtest.MockFinder{Err: errors.New("finder error")}

	if err := runSelect(afxMock, finderMock, bmPath); err == nil {
		t.Error("エラーが期待されましたが、nilが返りました")
	}
}

func TestRunSelect_ExcdError(t *testing.T) {
	tmpDir := t.TempDir()
	bmPath := filepath.Join(tmpDir, "bookmarks.txt")

	if err := os.WriteFile(bmPath, []byte("C:\\Users\\Test\\Dir1\n"), 0644); err != nil {
		t.Fatalf("テストファイル作成に失敗しました: %v", err)
	}

	afxMock := &afxtest.MockAFX{ExcdErr: errors.New("excd error")}
	finderMock := &afxtest.MockFinder{Idx: 0}

	err := runSelect(afxMock, finderMock, bmPath)
	if err == nil {
		t.Error("エラーが期待されましたが、nilが返りました")
	}
	if err.Error() != "ディレクトリ移動に失敗しました: excd error" {
		t.Errorf("予期しないエラーメッセージ: %v", err)
	}
}
