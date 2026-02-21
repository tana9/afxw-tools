package afx

import (
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
