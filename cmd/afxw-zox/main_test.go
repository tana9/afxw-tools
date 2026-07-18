package main

import (
	"context"
	"errors"
	"testing"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/tana9/afxw-tools/cmd/afxw-zox/zoxide"
	"github.com/tana9/afxw-tools/internal/afxtest"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name          string
		entries       []zoxide.Entry
		queryErr      error
		client        *afxtest.MockClient
		finder        *afxtest.MockFinder
		wantDirectory string
		wantErr       string
	}{
		{
			name: "selects entry",
			entries: []zoxide.Entry{
				{Path: `C:\Users\Test`, Score: 10},
				{Path: `C:\Projects`, Score: 20},
			},
			client:        &afxtest.MockClient{},
			finder:        &afxtest.MockFinder{Idx: 1},
			wantDirectory: `C:\Projects`,
		},
		{
			name:    "empty entries",
			entries: []zoxide.Entry{},
			client:  &afxtest.MockClient{},
			finder:  &afxtest.MockFinder{},
		},
		{
			name:     "query error",
			queryErr: errors.New("query error"),
			client:   &afxtest.MockClient{},
			finder:   &afxtest.MockFinder{},
			wantErr:  "zoxideデータベースの取得に失敗しました: query error",
		},
		{
			name:    "selection cancelled",
			entries: []zoxide.Entry{{Path: `C:\Users\Test`, Score: 10}},
			client:  &afxtest.MockClient{},
			finder:  &afxtest.MockFinder{Err: fuzzyfinder.ErrAbort},
		},
		{
			name:    "selection error",
			entries: []zoxide.Entry{{Path: `C:\Users\Test`, Score: 10}},
			client:  &afxtest.MockClient{},
			finder:  &afxtest.MockFinder{Err: errors.New("finder error")},
			wantErr: "finder error",
		},
		{
			name:    "directory change error",
			entries: []zoxide.Entry{{Path: `C:\Users\Test`, Score: 10}},
			client: &afxtest.MockClient{
				ChangeDirectoryErr: errors.New("change directory error"),
			},
			finder:  &afxtest.MockFinder{Idx: 0},
			wantErr: "ディレクトリ移動に失敗しました: change directory error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := func() ([]zoxide.Entry, error) { return tt.entries, tt.queryErr }
			err := run(tt.client, tt.finder, query)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("run() error = %v", err)
				}
			} else if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("run() error = %v, want %q", err, tt.wantErr)
			}
			if tt.client.ChangedDirectory != tt.wantDirectory {
				t.Errorf("ChangedDirectory = %q, want %q", tt.client.ChangedDirectory, tt.wantDirectory)
			}
		})
	}
}

func TestBuildZFormat(t *testing.T) {
	got := buildZFormat([]string{`C:\Users\Test`, `C:\Projects`}, 1234567890)
	want := "C:\\Users\\Test|1|1234567890\nC:\\Projects|1|1234567890\n"
	if got != want {
		t.Errorf("buildZFormat() = %q, want %q", got, want)
	}
}

func TestRunImport(t *testing.T) {
	tests := []struct {
		name    string
		client  *afxtest.MockClient
		wantErr string
	}{
		{
			name:    "history error",
			client:  &afxtest.MockClient{DirectoryHistoriesErr: errors.New("history error")},
			wantErr: "履歴の取得に失敗しました: history error",
		},
		{
			name:   "empty history",
			client: &afxtest.MockClient{DirectoryHistoriesResult: []string{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runImport(context.Background(), tt.client)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("runImport() error = %v", err)
				}
			} else if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("runImport() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}
