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

func TestResolveExecutable(t *testing.T) {
	t.Run("PATH takes priority", func(t *testing.T) {
		pathDir := t.TempDir()
		pathExecutable := makeExe(t, pathDir, "tool.exe")
		candidate := makeExe(t, t.TempDir(), "tool.exe")
		t.Setenv("PATH", pathDir)

		if got := ResolveExecutable("tool.exe", candidate); got != pathExecutable {
			t.Fatalf("ResolveExecutable() = %q, want PATH executable %q", got, pathExecutable)
		}
	})

	t.Run("candidate", func(t *testing.T) {
		t.Setenv("PATH", "")
		executable := filepath.Join(t.TempDir(), "tool.exe")
		if err := os.WriteFile(executable, nil, 0644); err != nil {
			t.Fatal(err)
		}
		if got := ResolveExecutable("tool.exe", "", executable); got != executable {
			t.Fatalf("ResolveExecutable() = %q, want %q", got, executable)
		}
	})

	t.Run("first existing candidate", func(t *testing.T) {
		t.Setenv("PATH", "")
		first := makeExe(t, t.TempDir(), "first.exe")
		second := makeExe(t, t.TempDir(), "second.exe")
		if got := ResolveExecutable("tool.exe", first, second); got != first {
			t.Fatalf("ResolveExecutable() = %q, want %q", got, first)
		}
	})

	t.Run("directory is ignored", func(t *testing.T) {
		t.Setenv("PATH", "")
		directory := t.TempDir()
		executable := makeExe(t, t.TempDir(), "tool.exe")
		if got := ResolveExecutable("tool.exe", directory, executable); got != executable {
			t.Fatalf("ResolveExecutable() = %q, want %q", got, executable)
		}
	})

	t.Run("fallback", func(t *testing.T) {
		t.Setenv("PATH", "")
		if got := ResolveExecutable("missing-tool.exe", "", filepath.Join(t.TempDir(), "missing.exe")); got != "missing-tool.exe" {
			t.Fatalf("ResolveExecutable() = %q", got)
		}
	})
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
