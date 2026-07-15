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

func TestParseQueryOutput_PreservesNonExistentPaths(t *testing.T) {
	dirs := makeTestDirs(t, "exists")
	nonExistent := filepath.Join(t.TempDir(), "no_such_dir")
	input := fmt.Sprintf("20.0 %s\n5.0 %s\n", dirs[0], nonExistent)

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
	if got[1].Path != nonExistent {
		t.Errorf("expected %q, got %q", nonExistent, got[1].Path)
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

func TestCandidateExecutablePaths_ScoopEnv(t *testing.T) {
	t.Setenv("SCOOP", `C:\custom\scoop`)
	t.Setenv("SCOOP_GLOBAL", `C:\custom\scoop-global`)
	t.Setenv("LOCALAPPDATA", "")

	got := candidateExecutablePaths()
	want := []string{
		`C:\custom\scoop\shims\zoxide.exe`,
		`C:\custom\scoop-global\shims\zoxide.exe`,
	}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d]: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCandidateExecutablePaths_FallbackToDefaults(t *testing.T) {
	t.Setenv("SCOOP", "")
	t.Setenv("SCOOP_GLOBAL", "")
	t.Setenv("ProgramData", `C:\ProgramData`)
	t.Setenv("LOCALAPPDATA", "")

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}

	got := candidateExecutablePaths()
	want := []string{
		filepath.Join(home, "scoop", "shims", "zoxide.exe"),
		`C:\ProgramData\scoop\shims\zoxide.exe`,
	}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d]: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCandidateExecutablePaths_WingetLinks(t *testing.T) {
	localAppData := t.TempDir()
	t.Setenv("LOCALAPPDATA", localAppData)

	got := candidateExecutablePaths()
	want := filepath.Join(localAppData, "Microsoft", "WinGet", "Links", "zoxide.exe")
	found := false
	for _, c := range got {
		if c == want {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected %q among candidates, got %v", want, got)
	}
}

func TestCandidateExecutablePaths_WingetPackagesGlob(t *testing.T) {
	t.Setenv("SCOOP", "")
	t.Setenv("SCOOP_GLOBAL", "")
	t.Setenv("ProgramData", "")
	localAppData := t.TempDir()
	t.Setenv("LOCALAPPDATA", localAppData)

	nestedDir := filepath.Join(localAppData, "Microsoft", "WinGet", "Packages", "ajeetdsouza.zoxide_Microsoft.Winget.Source_8wekyb3d8bbwe", "zoxide-0.9.4-x86_64-pc-windows-msvc")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatal(err)
	}
	exePath := filepath.Join(nestedDir, "zoxide.exe")
	if err := os.WriteFile(exePath, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	got := candidateExecutablePaths()
	found := false
	for _, c := range got {
		if c == exePath {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected %q among candidates, got %v", exePath, got)
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
