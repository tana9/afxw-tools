package main

import (
	"testing"
)

func TestRun(t *testing.T) {
	// run()関数は外部依存が多く、統合テストとして実行する必要があります。
	// ここでは関数が存在することを確認するのみとします。
	t.Skip("run()は外部依存（afx.NewOleAFX、zoxide.Query、finder）があるため、単体テストではスキップします")
}
