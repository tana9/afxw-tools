//go:build integration

package main

import (
	"testing"

	"github.com/tana9/afxw-tools/internal/finder"
)

func TestGoFuzzyFinder(t *testing.T) {
	// これは GoFuzzyFinder が期待通りに動作することを保証するための統合テストです。
	// 単体テストではありませんが、あると便利です。
	f := &finder.GoFuzzyFinder{}
	_, err := f.Find([]string{"a", "b"})
	if err == nil {
		t.Errorf("expected an error from fuzzyfinder, but got none")
	}
}
