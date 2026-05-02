package bookmark

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_NonExistentFile(t *testing.T) {
	dirs, err := Load(filepath.Join(t.TempDir(), "non_existent.txt"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dirs) != 0 {
		t.Errorf("expected empty slice, got %v", dirs)
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.txt")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	dirs, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dirs) != 0 {
		t.Errorf("expected empty slice, got %v", dirs)
	}
}

func TestLoad_WithContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bookmarks.txt")
	content := "C:\\Users\\Test\\Dir1\nC:\\Users\\Test\\Dir2\nC:\\Users\\Test\\Dir3\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	dirs, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dirs) != 3 {
		t.Errorf("expected 3 items, got %d", len(dirs))
	}
	if dirs[0] != `C:\Users\Test\Dir1` {
		t.Errorf("unexpected first item: %s", dirs[0])
	}
}

func TestLoad_WithDuplicates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bookmarks.txt")
	content := "C:\\Users\\Test\\Dir1\nC:\\Users\\Test\\Dir2\nC:\\Users\\Test\\Dir1\nC:\\Users\\Test\\Dir3\nC:\\Users\\Test\\Dir2\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	dirs, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dirs) != 3 {
		t.Errorf("expected 3 items after dedup, got %d", len(dirs))
	}
}

func TestLoad_WithEmptyLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bookmarks.txt")
	content := "C:\\Users\\Test\\Dir1\n\nC:\\Users\\Test\\Dir2\n\nC:\\Users\\Test\\Dir3\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	dirs, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dirs) != 3 {
		t.Errorf("expected 3 items, got %d", len(dirs))
	}
}

func TestAdd_NewItem(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bookmarks.txt")
	newItem := `C:\Users\Test\NewDir`

	if err := Add(path, newItem); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dirs, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dirs) != 1 {
		t.Errorf("expected 1 item, got %d", len(dirs))
	}
	if dirs[0] != filepath.Clean(newItem) {
		t.Errorf("expected %s, got %s", filepath.Clean(newItem), dirs[0])
	}
}

func TestAdd_DuplicateItem(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bookmarks.txt")
	item := `C:\Users\Test\Dir1`

	if err := Add(path, item); err != nil {
		t.Fatal(err)
	}
	if err := Add(path, item); err != nil {
		t.Fatal(err)
	}

	dirs, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dirs) != 1 {
		t.Errorf("expected 1 item, got %d", len(dirs))
	}
}

func TestAdd_CaseInsensitive(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bookmarks.txt")

	if err := Add(path, `C:\Users\Test\Dir1`); err != nil {
		t.Fatal(err)
	}
	if err := Add(path, `c:\users\test\dir1`); err != nil {
		t.Fatal(err)
	}

	dirs, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Windowsパスは大文字小文字を区別しないため1件のみ
	if len(dirs) != 1 {
		t.Errorf("expected 1 item, got %d", len(dirs))
	}
}

func TestAdd_MultipleItems(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bookmarks.txt")
	items := []string{`C:\Users\Test\Dir1`, `C:\Users\Test\Dir2`, `C:\Users\Test\Dir3`}

	for _, item := range items {
		if err := Add(path, item); err != nil {
			t.Fatalf("Add(%s): %v", item, err)
		}
	}

	dirs, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dirs) != len(items) {
		t.Errorf("expected %d items, got %d", len(items), len(dirs))
	}
}
