package cmdutil

import (
	"os"
	"path/filepath"
	"testing"
)

func makeExe(t *testing.T, dir, name string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte{}, 0755); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestFind_AbsolutePath_Exists(t *testing.T) {
	p := makeExe(t, t.TempDir(), "tool.exe")
	got, err := Find(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != p {
		t.Errorf("got %q, want %q", got, p)
	}
}

func TestFind_AbsolutePath_NotFound(t *testing.T) {
	p := filepath.Join(t.TempDir(), "nonexistent.exe")
	if _, err := Find(p); err == nil {
		t.Error("expected error for non-existent absolute path")
	}
}

func TestFind_ExtraDir_Found(t *testing.T) {
	dir := t.TempDir()
	makeExe(t, dir, "mytool.exe")

	got, err := Find("mytool.exe", dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != filepath.Join(dir, "mytool.exe") {
		t.Errorf("got %q, want %q", got, filepath.Join(dir, "mytool.exe"))
	}
}

func TestFind_ExtraDir_TakePriorityOverLater(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	makeExe(t, dir1, "tool.exe")
	makeExe(t, dir2, "tool.exe")

	got, err := Find("tool.exe", dir1, dir2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != filepath.Join(dir1, "tool.exe") {
		t.Errorf("expected dir1 to take priority, got %q", got)
	}
}

func TestFind_NotFoundAnywhere(t *testing.T) {
	if _, err := Find("this-tool-does-not-exist-afxw.exe"); err == nil {
		t.Error("expected error for unknown command")
	}
}

func TestFind_RelativeNotFoundInExtraDir(t *testing.T) {
	emptyDir := t.TempDir()
	// 存在しないディレクトリを extra dir に指定しても、PATH/exe dir で見つからなければエラー
	if _, err := Find("nonexistent-afxw-tool.exe", emptyDir); err == nil {
		t.Error("expected error when tool not found in extra dir or PATH")
	}
}
