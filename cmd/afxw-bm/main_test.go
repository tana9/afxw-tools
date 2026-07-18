package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/tana9/afxw-tools/internal/afx"
	"github.com/tana9/afxw-tools/internal/afxtest"
	"github.com/tana9/afxw-tools/internal/finder"
)

func preserveBookmarkDependencies(t *testing.T) {
	t.Helper()
	originalAcquire := acquireBookmarkInstance
	originalClient := newAFXClient
	originalPath := defaultBookmarkPath
	originalAdd := addBookmarkEntry
	originalFinder := newBookmarkFinder
	t.Cleanup(func() {
		acquireBookmarkInstance = originalAcquire
		newAFXClient = originalClient
		defaultBookmarkPath = originalPath
		addBookmarkEntry = originalAdd
		newBookmarkFinder = originalFinder
	})
}

func TestResolveBookmarkTarget(t *testing.T) {
	tests := []struct {
		name      string
		target    string
		client    afx.Client
		clientErr error
		want      string
	}{
		{name: "explicit target", target: `C:\Explicit`, want: `C:\Explicit`},
		{name: "active path", target: ".", client: &afxtest.MockClient{ActivePathResult: `C:\Active`}, want: `C:\Active`},
		{name: "connection failure", target: "", clientErr: errors.New("not running"), want: "."},
		{name: "active path failure", target: ".", client: &afxtest.MockClient{ActivePathErr: errors.New("failed")}, want: "."},
		{name: "empty active path", target: ".", client: &afxtest.MockClient{}, want: "."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preserveBookmarkDependencies(t)
			newAFXClient = func() (afx.Client, error) {
				if tt.client == nil && tt.clientErr == nil {
					t.Fatal("AFX client should not be created")
				}
				return tt.client, tt.clientErr
			}
			if got := resolveBookmarkTarget(tt.target); got != tt.want {
				t.Errorf("resolveBookmarkTarget(%q) = %q, want %q", tt.target, got, tt.want)
			}
		})
	}
}

func TestRunBookmarkCommand(t *testing.T) {
	t.Run("acquire error", func(t *testing.T) {
		preserveBookmarkDependencies(t)
		wantErr := errors.New("busy")
		acquireBookmarkInstance = func(string) error { return wantErr }
		if err := runBookmarkCommand(false, ""); !errors.Is(err, wantErr) {
			t.Fatalf("runBookmarkCommand() error = %v, want %v", err, wantErr)
		}
	})

	t.Run("add mode", func(t *testing.T) {
		preserveBookmarkDependencies(t)
		acquireBookmarkInstance = func(string) error { return nil }
		defaultBookmarkPath = func() (string, error) { return `C:\bin\bookmarks.txt`, nil }
		var gotPath, gotItem string
		addBookmarkEntry = func(path, item string) error {
			gotPath, gotItem = path, item
			return nil
		}
		if err := runBookmarkCommand(true, `C:\Target`); err != nil {
			t.Fatalf("runBookmarkCommand() error = %v", err)
		}
		if gotPath != `C:\bin\bookmarks.txt` || gotItem != `C:\Target` {
			t.Errorf("Add(%q, %q)", gotPath, gotItem)
		}
	})

	t.Run("select mode", func(t *testing.T) {
		preserveBookmarkDependencies(t)
		path := filepath.Join(t.TempDir(), "bookmarks.txt")
		if err := os.WriteFile(path, []byte("C:\\Selected\n"), 0644); err != nil {
			t.Fatal(err)
		}
		client := &afxtest.MockClient{}
		acquireBookmarkInstance = func(string) error { return nil }
		defaultBookmarkPath = func() (string, error) { return path, nil }
		newAFXClient = func() (afx.Client, error) { return client, nil }
		newBookmarkFinder = func() finder.Finder { return &afxtest.MockFinder{Idx: 0} }
		if err := runBookmarkCommand(false, ""); err != nil {
			t.Fatalf("runBookmarkCommand() error = %v", err)
		}
		if client.ChangedDirectory != `C:\Selected` {
			t.Errorf("ChangedDirectory = %q", client.ChangedDirectory)
		}
	})
}

func TestRunSelect(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		client        *afxtest.MockClient
		finder        *afxtest.MockFinder
		wantDirectory string
		wantErr       string
	}{
		{
			name:          "selects bookmark",
			content:       "C:\\Users\\Test\\Dir1\nC:\\Users\\Test\\Dir2\n",
			client:        &afxtest.MockClient{},
			finder:        &afxtest.MockFinder{Idx: 1},
			wantDirectory: `C:\Users\Test\Dir2`,
		},
		{
			name:   "empty bookmarks",
			client: &afxtest.MockClient{},
			finder: &afxtest.MockFinder{},
		},
		{
			name:    "selection cancelled",
			content: "C:\\Users\\Test\\Dir1\n",
			client:  &afxtest.MockClient{},
			finder:  &afxtest.MockFinder{Err: fuzzyfinder.ErrAbort},
		},
		{
			name:    "selection error",
			content: "C:\\Users\\Test\\Dir1\n",
			client:  &afxtest.MockClient{},
			finder:  &afxtest.MockFinder{Err: errors.New("finder error")},
			wantErr: "finder error",
		},
		{
			name:    "directory change error",
			content: "C:\\Users\\Test\\Dir1\n",
			client: &afxtest.MockClient{
				ChangeDirectoryErr: errors.New("change directory error"),
			},
			finder:  &afxtest.MockFinder{Idx: 0},
			wantErr: "ディレクトリ移動に失敗しました: change directory error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookmarkPath := filepath.Join(t.TempDir(), "bookmarks.txt")
			if tt.content != "" {
				if err := os.WriteFile(bookmarkPath, []byte(tt.content), 0644); err != nil {
					t.Fatal(err)
				}
			}

			err := runSelect(tt.client, tt.finder, bookmarkPath)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("runSelect() error = %v", err)
				}
			} else if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("runSelect() error = %v, want %q", err, tt.wantErr)
			}
			if tt.client.ChangedDirectory != tt.wantDirectory {
				t.Errorf("ChangedDirectory = %q, want %q", tt.client.ChangedDirectory, tt.wantDirectory)
			}
		})
	}
}
