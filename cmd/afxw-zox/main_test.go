package main

import (
	"errors"
	"testing"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/tana9/afxw-tools/cmd/afxw-zox/zoxide"
	"github.com/tana9/afxw-tools/internal/afxtest"
)

func makeQuery(entries []zoxide.Entry, err error) func() ([]zoxide.Entry, error) {
	return func() ([]zoxide.Entry, error) {
		return entries, err
	}
}

func TestRun_Normal(t *testing.T) {
	afxMock := &afxtest.MockAFX{}
	finderMock := &afxtest.MockFinder{Idx: 1}
	query := makeQuery([]zoxide.Entry{
		{Path: `C:\Users\Test`, Score: 10.0},
		{Path: `C:\Projects`, Score: 20.0},
	}, nil)

	if err := run(afxMock, finderMock, query); err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	if afxMock.ExcdPath != `C:\Projects` {
		t.Errorf("期待: C:\\Projects, 取得: %s", afxMock.ExcdPath)
	}
}

func TestRun_EmptyEntries(t *testing.T) {
	afxMock := &afxtest.MockAFX{}
	finderMock := &afxtest.MockFinder{}
	query := makeQuery([]zoxide.Entry{}, nil)

	if err := run(afxMock, finderMock, query); err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	if afxMock.ExcdPath != "" {
		t.Errorf("EXCDが呼ばれるべきではありません: %s", afxMock.ExcdPath)
	}
}

func TestRun_QueryError(t *testing.T) {
	afxMock := &afxtest.MockAFX{}
	finderMock := &afxtest.MockFinder{}
	query := makeQuery(nil, errors.New("query error"))

	err := run(afxMock, finderMock, query)
	if err == nil {
		t.Fatal("エラーが期待されましたが、nilが返りました")
	}
	if err.Error() != "zoxideデータベースの取得に失敗しました: query error" {
		t.Errorf("予期しないエラーメッセージ: %v", err)
	}
}

func TestRun_FinderCancelled(t *testing.T) {
	afxMock := &afxtest.MockAFX{}
	finderMock := &afxtest.MockFinder{Err: fuzzyfinder.ErrAbort}
	query := makeQuery([]zoxide.Entry{{Path: `C:\Users\Test`, Score: 10.0}}, nil)

	if err := run(afxMock, finderMock, query); err != nil {
		t.Fatalf("キャンセルはエラーになるべきではありません: %v", err)
	}
}

func TestRun_FinderError(t *testing.T) {
	afxMock := &afxtest.MockAFX{}
	finderMock := &afxtest.MockFinder{Err: errors.New("finder error")}
	query := makeQuery([]zoxide.Entry{{Path: `C:\Users\Test`, Score: 10.0}}, nil)

	if err := run(afxMock, finderMock, query); err == nil {
		t.Fatal("エラーが期待されましたが、nilが返りました")
	}
}

func TestRun_ExcdError(t *testing.T) {
	afxMock := &afxtest.MockAFX{ExcdErr: errors.New("excd error")}
	finderMock := &afxtest.MockFinder{Idx: 0}
	query := makeQuery([]zoxide.Entry{{Path: `C:\Users\Test`, Score: 10.0}}, nil)

	err := run(afxMock, finderMock, query)
	if err == nil {
		t.Fatal("エラーが期待されましたが、nilが返りました")
	}
	if err.Error() != "ディレクトリ移動に失敗しました: excd error" {
		t.Errorf("予期しないエラーメッセージ: %v", err)
	}
}
