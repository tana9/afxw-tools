package main

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/tana9/afxw-tools/cmd/afxw-launcher/config"
)

func TestExecuteCommandWith(t *testing.T) {
	item := config.MenuItem{Command: "tool.exe", Args: []string{"--flag"}}

	t.Run("runs resolved command", func(t *testing.T) {
		var gotPath string
		var gotArgs []string
		err := executeCommandWith(
			item,
			func(string) (string, error) { return `C:\tools\tool.exe`, nil },
			func(args []string) ([]string, error) { return append(args, "value"), nil },
			func(path string, args []string) error {
				gotPath = path
				gotArgs = args
				return nil
			},
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotPath != `C:\tools\tool.exe` {
			t.Errorf("got path %q", gotPath)
		}
		if want := []string{"--flag", "value"}; !reflect.DeepEqual(gotArgs, want) {
			t.Errorf("got args %v, want %v", gotArgs, want)
		}
	})

	t.Run("command lookup error", func(t *testing.T) {
		err := executeCommandWith(item, func(string) (string, error) {
			return "", errors.New("not found")
		}, nil, nil)
		if err == nil || err.Error() != "not found" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("argument resolution error", func(t *testing.T) {
		err := executeCommandWith(item, func(string) (string, error) {
			return "tool.exe", nil
		}, func([]string) ([]string, error) {
			return nil, errors.New("argument failed")
		}, nil)
		if err == nil || err.Error() != "argument failed" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("command execution error", func(t *testing.T) {
		err := executeCommandWith(item, func(string) (string, error) {
			return "tool.exe", nil
		}, func(args []string) ([]string, error) {
			return args, nil
		}, func(string, []string) error {
			return errors.New("execution failed")
		})
		if err == nil || !strings.Contains(err.Error(), "コマンドの実行に失敗しました") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
