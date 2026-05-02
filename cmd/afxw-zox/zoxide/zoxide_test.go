package zoxide

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func makeTestDirs(t *testing.T, names ...string) []string {
	t.Helper()
	base := t.TempDir()
	paths := make([]string, len(names))
	for i, name := range names {
		p := filepath.Join(base, name)
		if err := os.MkdirAll(p, 0755); err != nil {
			t.Fatal(err)
		}
		paths[i] = p
	}
	return paths
}

func TestParseQueryOutput_Normal(t *testing.T) {
	dirs := makeTestDirs(t, "a", "b")
	input := fmt.Sprintf("12.5 %s\n10.0 %s\n", dirs[0], dirs[1])

	got, err := parseQueryOutput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	if got[0].Path != dirs[0] || got[0].Score != 12.5 {
		t.Errorf("entry[0]: got {%s, %v}, want {%s, 12.5}", got[0].Path, got[0].Score, dirs[0])
	}
	if got[1].Path != dirs[1] || got[1].Score != 10.0 {
		t.Errorf("entry[1]: got {%s, %v}, want {%s, 10.0}", got[1].Path, got[1].Score, dirs[1])
	}
}

func TestParseQueryOutput_PathWithSpaces(t *testing.T) {
	dirs := makeTestDirs(t, "my folder", "another path")
	input := fmt.Sprintf("15.0 %s\n8.5 %s\n", dirs[0], dirs[1])

	got, err := parseQueryOutput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	if got[0].Path != dirs[0] {
		t.Errorf("expected %q, got %q", dirs[0], got[0].Path)
	}
}

func TestParseQueryOutput_NonExistentPathFiltered(t *testing.T) {
	dirs := makeTestDirs(t, "exists")
	nonExistent := filepath.Join(t.TempDir(), "no_such_dir")
	input := fmt.Sprintf("20.0 %s\n5.0 %s\n", dirs[0], nonExistent)

	got, err := parseQueryOutput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 entry (non-existent filtered), got %d", len(got))
	}
	if got[0].Path != dirs[0] {
		t.Errorf("expected %q, got %q", dirs[0], got[0].Path)
	}
}

func TestParseQueryOutput_Empty(t *testing.T) {
	got, err := parseQueryOutput("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty result, got %d entries", len(got))
	}
}

func TestParseQueryOutput_BlankLinesSkipped(t *testing.T) {
	dirs := makeTestDirs(t, "x", "y")
	input := fmt.Sprintf("12.5 %s\n\n10.0 %s\n", dirs[0], dirs[1])

	got, err := parseQueryOutput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 entries, got %d", len(got))
	}
}

func TestParseQueryOutput_InvalidLinesSkipped(t *testing.T) {
	dirs := makeTestDirs(t, "valid")
	input := fmt.Sprintf("notanumber %s\n10.0 %s\n", dirs[0], dirs[0])

	got, err := parseQueryOutput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 entry (invalid score line skipped), got %d", len(got))
	}
}

func TestPaths(t *testing.T) {
	entries := []Entry{
		{Path: `C:\a`, Score: 10.0},
		{Path: `C:\b`, Score: 15.0},
		{Path: `C:\c`, Score: 20.0},
	}
	got := Paths(entries)
	if len(got) != len(entries) {
		t.Fatalf("expected %d paths, got %d", len(entries), len(got))
	}
	for i, e := range entries {
		if got[i] != e.Path {
			t.Errorf("[%d]: got %q, want %q", i, got[i], e.Path)
		}
	}
}
