package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tana9/afxw-tools/cmd/afxw-launcher/config"
)

// keyMsg はテスト用にmsg.String()が期待通りの文字列を返すtea.KeyMsgを組み立てます。
func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func menuOfSize(n int) *config.Config {
	items := make([]config.MenuItem, n)
	for i := range items {
		items[i] = config.MenuItem{Name: "item", Command: "tool.exe"}
	}
	return &config.Config{Menu: items}
}

func TestNumberSelection_SingleDigitAutoCommitsWhenUnambiguous(t *testing.T) {
	m := model{cfg: menuOfSize(6)}

	updated, cmd := m.Update(keyMsg("3"))
	got := updated.(model)

	if !got.selected || got.cursor != 2 {
		t.Fatalf("selected=%v cursor=%d, want selected=true cursor=2", got.selected, got.cursor)
	}
	if cmd == nil {
		t.Error("expected tea.Quit command")
	}
}

func TestNumberSelection_OutOfRangeDigitIsIgnored(t *testing.T) {
	m := model{cfg: menuOfSize(6)}

	updated, _ := m.Update(keyMsg("9"))
	got := updated.(model)

	if got.selected || got.cursor != 0 {
		t.Fatalf("selected=%v cursor=%d, want unchanged", got.selected, got.cursor)
	}
}

func TestNumberSelection_LeadingZeroIsIgnored(t *testing.T) {
	m := model{cfg: menuOfSize(6)}

	updated, _ := m.Update(keyMsg("0"))
	got := updated.(model)

	if got.selected || got.numberInput != "" {
		t.Fatalf("selected=%v numberInput=%q, want no state change", got.selected, got.numberInput)
	}
}

func TestNumberSelection_AmbiguousMultiDigitWaitsThenAutoCommits(t *testing.T) {
	m := model{cfg: menuOfSize(12)}

	updated, cmd := m.Update(keyMsg("1"))
	got := updated.(model)
	if got.selected || got.numberInput != "1" {
		t.Fatalf("after first digit: selected=%v numberInput=%q, want waiting with buffer \"1\"", got.selected, got.numberInput)
	}
	if cmd != nil {
		t.Error("expected no command while ambiguous")
	}

	updated, cmd = got.Update(keyMsg("2"))
	got = updated.(model)
	if !got.selected || got.cursor != 11 {
		t.Fatalf("after second digit: selected=%v cursor=%d, want selected=true cursor=11 (item 12)", got.selected, got.cursor)
	}
	if cmd == nil {
		t.Error("expected tea.Quit command once unambiguous")
	}
}

func TestNumberSelection_EnterCommitsAmbiguousPartialNumber(t *testing.T) {
	m := model{cfg: menuOfSize(12)}

	updated, _ := m.Update(keyMsg("1"))
	got := updated.(model)

	updated, cmd := got.Update(keyMsg("enter"))
	got = updated.(model)
	if !got.selected || got.cursor != 0 {
		t.Fatalf("selected=%v cursor=%d, want selected=true cursor=0 (item 1)", got.selected, got.cursor)
	}
	if cmd == nil {
		t.Error("expected tea.Quit command")
	}
}

func TestNumberSelection_NavigationClearsBuffer(t *testing.T) {
	m := model{cfg: menuOfSize(12)}

	updated, _ := m.Update(keyMsg("1"))
	got := updated.(model)
	if got.numberInput != "1" {
		t.Fatalf("numberInput = %q, want \"1\"", got.numberInput)
	}

	updated, _ = got.Update(keyMsg("down"))
	got = updated.(model)
	if got.numberInput != "" {
		t.Errorf("numberInput = %q, want cleared after navigation", got.numberInput)
	}
	if got.cursor != 1 {
		t.Errorf("cursor = %d, want 1", got.cursor)
	}
}
