package selectnav

import (
	"errors"
	"testing"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/tana9/afxw-tools/internal/afxtest"
)

func TestSelectAndMove_Normal(t *testing.T) {
	afxMock := &afxtest.MockAFX{}
	finderMock := &afxtest.MockFinder{Idx: 1}

	if err := SelectAndMove(afxMock, finderMock, []string{`C:\Dir1`, `C:\Dir2`}); err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if afxMock.ExcdPath != `C:\Dir2` {
		t.Errorf("期待: C:\\Dir2, 取得: %s", afxMock.ExcdPath)
	}
}

func TestSelectAndMove_EmptyDirs(t *testing.T) {
	afxMock := &afxtest.MockAFX{}
	finderMock := &afxtest.MockFinder{}

	if err := SelectAndMove(afxMock, finderMock, nil); err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if afxMock.ExcdPath != "" {
		t.Errorf("EXCDが呼ばれるべきではありません: %s", afxMock.ExcdPath)
	}
}

func TestSelectAndMove_Cancelled(t *testing.T) {
	afxMock := &afxtest.MockAFX{}
	finderMock := &afxtest.MockFinder{Err: fuzzyfinder.ErrAbort}

	if err := SelectAndMove(afxMock, finderMock, []string{`C:\Dir1`}); err != nil {
		t.Fatalf("キャンセルはエラーになるべきではありません: %v", err)
	}
	if afxMock.ExcdPath != "" {
		t.Errorf("EXCDが呼ばれるべきではありません: %s", afxMock.ExcdPath)
	}
}

func TestSelectAndMove_FinderError(t *testing.T) {
	afxMock := &afxtest.MockAFX{}
	finderMock := &afxtest.MockFinder{Err: errors.New("finder error")}

	if err := SelectAndMove(afxMock, finderMock, []string{`C:\Dir1`}); err == nil {
		t.Error("エラーが期待されましたが、nilが返りました")
	}
}

func TestSelectAndMove_ExcdError(t *testing.T) {
	afxMock := &afxtest.MockAFX{ExcdErr: errors.New("excd error")}
	finderMock := &afxtest.MockFinder{Idx: 0}

	err := SelectAndMove(afxMock, finderMock, []string{`C:\Dir1`})
	if err == nil {
		t.Fatal("エラーが期待されましたが、nilが返りました")
	}
	if err.Error() != "ディレクトリ移動に失敗しました: excd error" {
		t.Errorf("予期しないエラーメッセージ: %v", err)
	}
}
