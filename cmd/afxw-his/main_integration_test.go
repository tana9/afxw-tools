//go:build integration

package main

import (
	"testing"

	"github.com/tana9/afxw-tools/internal/finder"
)

func TestFuzzyFinder(t *testing.T) {
	// これは FuzzyFinder が期待通りに動作することを保証するための統合テストです。
	// 単体テストではありませんが、あると便利です。
	f := &finder.FuzzyFinder{}
	_, err := f.Find([]string{"a", "b"})
	if err == nil {
		t.Errorf("expected an error from fuzzyfinder, but got none")
	}
}
