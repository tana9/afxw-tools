package afx

import (
	"errors"
	"reflect"
	"testing"
)

func TestClose_NilFields(t *testing.T) {
	// afxw, unknown が両方 nil の状態で Close() を呼んでもパニックしないことを確認
	a := &oleAFX{afxw: nil, unknown: nil}
	a.Close()
}

func TestEnsureTrailingBackslash(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`C:\Users\Test`, `C:\Users\Test\`},
		{`C:\Users\Test\`, `C:\Users\Test\`},
		{`C:\`, `C:\`},
		{`C:`, `C:\`},
		{``, `\`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ensureTrailingBackslash(tt.input)
			if got != tt.expected {
				t.Errorf("期待: %q, 取得: %q", tt.expected, got)
			}
		})
	}
}

func TestParseMarkedFiles(t *testing.T) {
	t.Run("marked files", func(t *testing.T) {
		got, err := parseMarkedFiles("C:\\a.txt C:\\b.txt", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{`C:\a.txt`, `C:\b.txt`}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("no marks uses current file", func(t *testing.T) {
		got, err := parseMarkedFiles("  ", func() (string, error) {
			return `C:\current.txt`, nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(got, []string{`C:\current.txt`}) {
			t.Errorf("unexpected files: %v", got)
		}
	})

	t.Run("current file error", func(t *testing.T) {
		_, err := parseMarkedFiles("", func() (string, error) {
			return "", errors.New("current file failed")
		})
		if err == nil || err.Error() != "current file failed" {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
