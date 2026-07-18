package main

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tana9/afxw-tools/cmd/afxw-launcher/config"
)

func TestRunCommandCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runCommand(ctx, "cmd.exe", []string{"/c", "exit", "0"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("runCommand() error = %v, want context.Canceled", err)
	}
}

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

func TestUpdateNavigationBounds(t *testing.T) {
	m := newTestModel()
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if got := result.(model).cursor; got != 0 {
		t.Errorf("cursor after moving above first item = %d, want 0", got)
	}

	m.cursor = len(m.cfg.Menu) - 1
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if got := result.(model).cursor; got != len(m.cfg.Menu)-1 {
		t.Errorf("cursor after moving below last item = %d, want %d", got, len(m.cfg.Menu)-1)
	}
}

func TestUpdateEnter(t *testing.T) {
	t.Run("selects current item", func(t *testing.T) {
		m := newTestModel()
		m.cursor = 1
		result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		got := result.(model)
		if !got.selected || got.cursor != 1 {
			t.Errorf("model = %+v", got)
		}
	})

	t.Run("empty menu is ignored", func(t *testing.T) {
		m := model{cfg: &config.Config{}}
		result, command := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		got := result.(model)
		if got.selected || command != nil {
			t.Errorf("selected = %v, command = %v", got.selected, command)
		}
	})
}

func TestUpdateQuitKeys(t *testing.T) {
	for _, key := range []string{"q", "esc", "ctrl+c"} {
		t.Run(key, func(t *testing.T) {
			m := newTestModel()
			result, command := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
			got := result.(model)
			if !got.quitting || got.selected || command == nil {
				t.Errorf("model = %+v, command nil = %v", got, command == nil)
			}
		})
	}
}
