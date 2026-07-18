package main

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/tana9/afxw-tools/internal/cliutil"
)

func TestRunStartsWinMergeWithTwoPaths(t *testing.T) {
	paths := []string{`C:\a file.txt`, `C:\folder b`}
	var gotExecutable string
	var gotArgs []string
	err := run(paths, `C:\WinMerge\WinMergeU.exe`, func(path string, args []string) error {
		gotExecutable = path
		gotArgs = args
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotExecutable != `C:\WinMerge\WinMergeU.exe` || !reflect.DeepEqual(gotArgs, paths) {
		t.Fatalf("executable = %q, args = %v", gotExecutable, gotArgs)
	}
}

func TestRunRequiresExactlyTwoPaths(t *testing.T) {
	for _, paths := range [][]string{nil, {`C:\one`}, {`C:\one`, `C:\two`, `C:\three`}} {
		err := run(paths, "WinMergeU.exe", func(string, []string) error {
			t.Fatal("WinMerge should not be started")
			return nil
		})
		var notice *cliutil.Notice
		if !errors.As(err, &notice) {
			t.Fatalf("paths = %v, error = %v, want Notice", paths, err)
		}
	}
}

func TestRunStartError(t *testing.T) {
	err := run([]string{"a", "b"}, "WinMergeU.exe", func(string, []string) error {
		return errors.New("start failed")
	})
	if err == nil || !strings.Contains(err.Error(), "WinMergeの起動に失敗しました") {
		t.Fatalf("error = %v", err)
	}
}

func TestResolveWinMergeStandardLocations(t *testing.T) {
	t.Setenv("PATH", "")
	t.Setenv("ProgramFiles", "")
	t.Setenv("ProgramFiles(x86)", "")
	localAppData := t.TempDir()
	t.Setenv("LOCALAPPDATA", localAppData)

	executable := filepath.Join(localAppData, "Programs", "WinMerge", "WinMergeU.exe")
	if err := os.MkdirAll(filepath.Dir(executable), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(executable, nil, 0644); err != nil {
		t.Fatal(err)
	}
	if got := resolveWinMerge(); got != executable {
		t.Fatalf("resolveWinMerge() = %q, want %q", got, executable)
	}
}
