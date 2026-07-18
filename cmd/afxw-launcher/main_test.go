package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tana9/afxw-tools/cmd/afxw-launcher/config"
)

func newTestModel() model {
	return model{
		cfg: &config.Config{
			Menu: []config.MenuItem{
				{Name: "Item1", Command: "cmd1.exe"},
				{Name: "Item2", Command: "cmd2.exe"},
				{Name: "Item3", Command: "cmd3.exe"},
			},
		},
	}
}

func TestUpdate_NumberKey_SelectsItem(t *testing.T) {
	tests := []struct {
		key            string
		expectedCursor int
	}{
		{"1", 0},
		{"2", 1},
		{"3", 2},
	}

	for _, tt := range tests {
		t.Run("key="+tt.key, func(t *testing.T) {
			m := newTestModel()
			result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			got := result.(model)

			if !got.selected {
				t.Error("selected が true になるべきです")
			}
			if got.cursor != tt.expectedCursor {
				t.Errorf("cursor: 期待=%d, 取得=%d", tt.expectedCursor, got.cursor)
			}
		})
	}
}

func TestUpdate_NumberKey_OutOfRange(t *testing.T) {
	m := newTestModel() // メニューは3件

	// 範囲外の "9" を押しても選択されない
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("9")})
	got := result.(model)

	if got.selected {
		t.Error("範囲外の番号で selected が true になるべきではありません")
	}
}

func TestUpdate_ArrowKeys(t *testing.T) {
	m := newTestModel()

	// 下に移動
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = result.(model)
	if m.cursor != 1 {
		t.Errorf("j キー後の cursor: 期待=1, 取得=%d", m.cursor)
	}

	// 上に移動
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = result.(model)
	if m.cursor != 0 {
		t.Errorf("k キー後の cursor: 期待=0, 取得=%d", m.cursor)
	}
}
