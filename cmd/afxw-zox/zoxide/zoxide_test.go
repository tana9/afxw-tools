package zoxide

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestQueryWithCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := queryWith(ctx, "cmd.exe")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("queryWith() error = %v, want context.Canceled", err)
	}
}

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

func TestParseQueryOutput(t *testing.T) {
	dirs := makeTestDirs(t, "a", "b", "my folder", "another path")
	nonExistent := filepath.Join(t.TempDir(), "no_such_dir")
	tests := []struct {
		name  string
		input string
		want  []Entry
	}{
		{
			name:  "normal",
			input: fmt.Sprintf("12.5 %s\n10.0 %s\n", dirs[0], dirs[1]),
			want:  []Entry{{Path: dirs[0], Score: 12.5}, {Path: dirs[1], Score: 10}},
		},
		{
			name:  "paths with spaces",
			input: fmt.Sprintf("15.0 %s\n8.5 %s\n", dirs[2], dirs[3]),
			want:  []Entry{{Path: dirs[2], Score: 15}, {Path: dirs[3], Score: 8.5}},
		},
		{
			name:  "nonexistent path preserved",
			input: fmt.Sprintf("20.0 %s\n", nonExistent),
			want:  []Entry{{Path: nonExistent, Score: 20}},
		},
		{name: "empty", input: "", want: nil},
		{
			name:  "blank lines skipped",
			input: fmt.Sprintf("12.5 %s\n\n10.0 %s\n", dirs[0], dirs[1]),
			want:  []Entry{{Path: dirs[0], Score: 12.5}, {Path: dirs[1], Score: 10}},
		},
		{
			name:  "invalid lines skipped",
			input: fmt.Sprintf("notanumber %s\n10.0 %s\n", dirs[0], dirs[0]),
			want:  []Entry{{Path: dirs[0], Score: 10}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseQueryOutput(tt.input)
			if err != nil {
				t.Fatalf("parseQueryOutput() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseQueryOutput() = %#v, want %#v", got, tt.want)
			}
		})
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
