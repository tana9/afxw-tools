package afx

import (
	"testing"
)

func TestClose_NilFields(t *testing.T) {
	// afxw, unknown が両方 nil の状態で Close() を呼んでもパニックしないことを確認
	a := &oleAFX{afxw: nil, unknown: nil}
	a.Close()
}
