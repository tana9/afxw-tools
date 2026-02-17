//go:build integration

package main

import (
	"testing"
)

func TestGoFuzzyFinder(t *testing.T) {
	// これは goFuzzyFinder が期待通りに動作することを保証するための統合テストです。
	// 単体テストではありませんが、あると便利です。
	finder := &goFuzzyFinder{}
	_, err := finder.Find([]string{"a", "b"})
	if err == nil {
		t.Errorf("expected an error from fuzzyfinder, but got none")
	}
}
