package bookmark

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_NonExistentFile(t *testing.T) {
	// 存在しないファイルパスを指定
	tmpDir := t.TempDir()
	nonExistentPath := filepath.Join(tmpDir, "non_existent.txt")

	// 存在しないファイルの場合、空のスライスを返すことを確認
	dirs, err := Load(nonExistentPath)
	if err != nil {
		t.Fatalf("エラーが発生しました: %v", err)
	}

	if len(dirs) != 0 {
		t.Errorf("期待: 空のスライス, 取得: %v", dirs)
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	// 空のファイルを作成
	tmpDir := t.TempDir()
	emptyPath := filepath.Join(tmpDir, "empty.txt")

	if err := os.WriteFile(emptyPath, []byte(""), 0644); err != nil {
		t.Fatalf("空のファイル作成に失敗しました: %v", err)
	}

	dirs, err := Load(emptyPath)
	if err != nil {
		t.Fatalf("エラーが発生しました: %v", err)
	}

	if len(dirs) != 0 {
		t.Errorf("期待: 空のスライス, 取得: %v", dirs)
	}
}

func TestLoad_WithContent(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "bookmarks.txt")

	content := `C:\Users\Test\Dir1
C:\Users\Test\Dir2
C:\Users\Test\Dir3
`
	if err := os.WriteFile(testPath, []byte(content), 0644); err != nil {
		t.Fatalf("テストファイル作成に失敗しました: %v", err)
	}

	dirs, err := Load(testPath)
	if err != nil {
		t.Fatalf("エラーが発生しました: %v", err)
	}

	expected := 3
	if len(dirs) != expected {
		t.Errorf("期待: %d個, 取得: %d個", expected, len(dirs))
	}

	if dirs[0] != `C:\Users\Test\Dir1` {
		t.Errorf("期待: C:\\Users\\Test\\Dir1, 取得: %s", dirs[0])
	}
}

func TestLoad_WithDuplicates(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "bookmarks.txt")

	content := `C:\Users\Test\Dir1
C:\Users\Test\Dir2
C:\Users\Test\Dir1
C:\Users\Test\Dir3
C:\Users\Test\Dir2
`
	if err := os.WriteFile(testPath, []byte(content), 0644); err != nil {
		t.Fatalf("テストファイル作成に失敗しました: %v", err)
	}

	dirs, err := Load(testPath)
	if err != nil {
		t.Fatalf("エラーが発生しました: %v", err)
	}

	// 重複を除外して3個になることを確認
	expected := 3
	if len(dirs) != expected {
		t.Errorf("期待: %d個, 取得: %d個", expected, len(dirs))
	}
}

func TestLoad_WithEmptyLines(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "bookmarks.txt")

	content := `C:\Users\Test\Dir1

C:\Users\Test\Dir2

C:\Users\Test\Dir3
`
	if err := os.WriteFile(testPath, []byte(content), 0644); err != nil {
		t.Fatalf("テストファイル作成に失敗しました: %v", err)
	}

	dirs, err := Load(testPath)
	if err != nil {
		t.Fatalf("エラーが発生しました: %v", err)
	}

	// 空行を除外して3個になることを確認
	expected := 3
	if len(dirs) != expected {
		t.Errorf("期待: %d個, 取得: %d個", expected, len(dirs))
	}
}

func TestAdd_NewItem(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "bookmarks.txt")

	newItem := `C:\Users\Test\NewDir`
	if err := Add(testPath, newItem); err != nil {
		t.Fatalf("追加に失敗しました: %v", err)
	}

	// ファイルから読み込んで確認
	dirs, err := Load(testPath)
	if err != nil {
		t.Fatalf("読み込みに失敗しました: %v", err)
	}

	if len(dirs) != 1 {
		t.Errorf("期待: 1個, 取得: %d個", len(dirs))
	}

	// filepath.Cleanされた結果を期待
	expectedPath := filepath.Clean(newItem)
	if dirs[0] != expectedPath {
		t.Errorf("期待: %s, 取得: %s", expectedPath, dirs[0])
	}
}

func TestAdd_DuplicateItem(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "bookmarks.txt")

	item := `C:\Users\Test\Dir1`

	// 1回目の追加
	if err := Add(testPath, item); err != nil {
		t.Fatalf("1回目の追加に失敗しました: %v", err)
	}

	// 2回目の追加（重複）
	if err := Add(testPath, item); err != nil {
		t.Fatalf("2回目の追加に失敗しました: %v", err)
	}

	// ファイルから読み込んで確認
	dirs, err := Load(testPath)
	if err != nil {
		t.Fatalf("読み込みに失敗しました: %v", err)
	}

	// 重複が除外されて1個のみになることを確認
	if len(dirs) != 1 {
		t.Errorf("期待: 1個, 取得: %d個", len(dirs))
	}
}

func TestAdd_CaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "bookmarks.txt")

	item1 := `C:\Users\Test\Dir1`
	item2 := `c:\users\test\dir1` // 大文字小文字が異なる

	// 1回目の追加
	if err := Add(testPath, item1); err != nil {
		t.Fatalf("1回目の追加に失敗しました: %v", err)
	}

	// 2回目の追加（大文字小文字が異なる）
	if err := Add(testPath, item2); err != nil {
		t.Fatalf("2回目の追加に失敗しました: %v", err)
	}

	// ファイルから読み込んで確認
	dirs, err := Load(testPath)
	if err != nil {
		t.Fatalf("読み込みに失敗しました: %v", err)
	}

	// Windowsでは大文字小文字を区別しないため、1個のみになることを確認
	if len(dirs) != 1 {
		t.Errorf("期待: 1個（大文字小文字の区別なし）, 取得: %d個", len(dirs))
	}
}

func TestAdd_MultipleItems(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "bookmarks.txt")

	items := []string{
		`C:\Users\Test\Dir1`,
		`C:\Users\Test\Dir2`,
		`C:\Users\Test\Dir3`,
	}

	// 複数アイテムを追加
	for _, item := range items {
		if err := Add(testPath, item); err != nil {
			t.Fatalf("追加に失敗しました (%s): %v", item, err)
		}
	}

	// ファイルから読み込んで確認
	dirs, err := Load(testPath)
	if err != nil {
		t.Fatalf("読み込みに失敗しました: %v", err)
	}

	if len(dirs) != len(items) {
		t.Errorf("期待: %d個, 取得: %d個", len(items), len(dirs))
	}
}
