package main

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/tana9/afxw-tools/cmd/afxw-open/config"
)

func TestRunWithConfig(t *testing.T) {
	program := config.Program{Name: "Editor", Command: "editor.exe"}
	cfg := &config.Config{Programs: []config.Program{program}}

	t.Run("opens selected program", func(t *testing.T) {
		var gotProgram config.Program
		var gotFiles []string
		err := runWithConfig(
			[]string{`C:\a.txt`, `C:\b.txt`},
			cfg,
			func([]config.Program) (int, error) { return 0, nil },
			func(p config.Program, files []string) error {
				gotProgram = p
				gotFiles = files
				return nil
			},
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(gotProgram, program) || !reflect.DeepEqual(gotFiles, []string{`C:\a.txt`, `C:\b.txt`}) {
			t.Errorf("opened %+v with %v", gotProgram, gotFiles)
		}
	})

	t.Run("empty program configuration", func(t *testing.T) {
		err := runWithConfig(nil, &config.Config{}, nil, nil)
		if err == nil || !strings.Contains(err.Error(), "プログラムが設定されていません") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("selection cancelled", func(t *testing.T) {
		err := runWithConfig(nil, cfg, func([]config.Program) (int, error) {
			return 0, fuzzyfinder.ErrAbort
		}, func(config.Program, []string) error {
			t.Fatal("opener must not be called after cancellation")
			return nil
		})
		if err != nil {
			t.Fatalf("cancellation returned error: %v", err)
		}
	})

	t.Run("selection error", func(t *testing.T) {
		err := runWithConfig(nil, cfg, func([]config.Program) (int, error) {
			return 0, errors.New("selector failed")
		}, nil)
		if err == nil || !strings.Contains(err.Error(), "プログラムの選択に失敗しました") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestOpenFilesWith(t *testing.T) {
	program := config.Program{Command: "editor.exe", Args: []string{"--wait"}}

	t.Run("passes program and file arguments", func(t *testing.T) {
		var gotPath string
		var gotArgs []string
		err := openFilesWith(
			program,
			[]string{`C:\a.txt`, `C:\b.txt`},
			func(string) (string, error) { return `C:\tools\editor.exe`, nil },
			func(path string, args []string) error {
				gotPath = path
				gotArgs = args
				return nil
			},
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotPath != `C:\tools\editor.exe` {
			t.Errorf("got path %q", gotPath)
		}
		if want := []string{"--wait", `C:\a.txt`, `C:\b.txt`}; !reflect.DeepEqual(gotArgs, want) {
			t.Errorf("got args %v, want %v", gotArgs, want)
		}
	})

	t.Run("command lookup error", func(t *testing.T) {
		err := openFilesWith(program, nil, func(string) (string, error) {
			return "", errors.New("not found")
		}, nil)
		if err == nil || err.Error() != "not found" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("process start error", func(t *testing.T) {
		err := openFilesWith(program, nil, func(string) (string, error) {
			return "editor.exe", nil
		}, func(string, []string) error {
			return errors.New("start failed")
		})
		if err == nil || !strings.Contains(err.Error(), "プログラムの起動に失敗しました") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
