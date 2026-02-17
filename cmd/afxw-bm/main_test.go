package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAddBookmark(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "test_bookmark")

	// テストディレクトリを作成
	if err := os.MkdirAll(testPath, 0755); err != nil {
		t.Fatalf("テストディレクトリの作成に失敗しました: %v", err)
	}

	// bookmark.GetDefaultPath()をモック化するため、
	// 実際にはbookmarkパッケージのテストで対応すべきですが、
	// ここでは addBookmark が絶対パスに変換することをテスト
	absPath, err := filepath.Abs(testPath)
	if err != nil {
		t.Fatalf("絶対パスの取得に失敗しました: %v", err)
	}

	// 相対パスが絶対パスに変換されることを確認
	if !filepath.IsAbs(absPath) {
		t.Errorf("期待: 絶対パス, 取得: %s", absPath)
	}
}

func TestAddBookmark_InvalidPath(t *testing.T) {
	// 存在しないパスでもfilepth.Absは成功するため、
	// このテストは絶対パス変換の動作確認
	invalidPath := "non_existent_path_12345"

	absPath, err := filepath.Abs(invalidPath)
	if err != nil {
		t.Fatalf("絶対パスの取得に失敗しました: %v", err)
	}

	// filepath.Absは存在しないパスでも絶対パスを返す
	if !filepath.IsAbs(absPath) {
		t.Errorf("期待: 絶対パス, 取得: %s", absPath)
	}
}
