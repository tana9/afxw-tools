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

func TestToInt(t *testing.T) {
	t.Run("supported types", func(t *testing.T) {
		tests := []struct {
			name  string
			input any
			want  int
		}{
			{"int", int(3), 3},
			{"int16", int16(3), 3},
			{"int32", int32(3), 3},
			{"int64", int64(3), 3},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := toInt(tt.input)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != tt.want {
					t.Errorf("got %d, want %d", got, tt.want)
				}
			})
		}
	})

	t.Run("unsupported type", func(t *testing.T) {
		if _, err := toInt("3"); err == nil {
			t.Error("expected error for unsupported type, got nil")
		}
	})
}

func TestParseMarkedFiles(t *testing.T) {
	t.Run("marked files", func(t *testing.T) {
		got, err := parseMarkedFiles("\"C:\\a file.txt\"\r\n\"C:\\folder b\"", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{`C:\a file.txt`, `C:\folder b`}
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

func TestMarkedFilesMacroDoesNotAppendPaneLiteral(t *testing.T) {
	if markedFilesMacro != "$JU$MF" {
		t.Fatalf("markedFilesMacro = %q", markedFilesMacro)
	}
}
