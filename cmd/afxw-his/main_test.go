package main

import (
	"errors"
	"reflect"
	"testing"

	"github.com/tana9/afxw-tools/internal/afx"
	"github.com/tana9/afxw-tools/internal/afxtest"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name         string
		clientMock   *afxtest.MockClient
		finderMock   *afxtest.MockFinder
		expectErr    bool
		expectedErr  string
		expectedPath string
	}{
		{
			name: "normal run",
			clientMock: &afxtest.MockClient{
				DirectoryHistoriesResult: []string{"C:\\Windows", "C:\\Users"},
			},
			finderMock: &afxtest.MockFinder{
				Idx: 0,
			},
			expectedPath: "C:\\Windows",
		},
		{
			name: "normal run with selection",
			clientMock: &afxtest.MockClient{
				DirectoryHistoriesResult: []string{"C:\\Windows", "C:\\Users"},
			},
			finderMock: &afxtest.MockFinder{
				Idx: 1,
			},
			expectedPath: "C:\\Users",
		},
		{
			name: "finder cancelled",
			clientMock: &afxtest.MockClient{
				DirectoryHistoriesResult: []string{"C:\\Windows", "C:\\Users"},
			},
			finderMock: &afxtest.MockFinder{
				Err: errors.New("fuzzyfinder cancelled"),
			},
			expectErr:   true,
			expectedErr: "fuzzyfinder cancelled",
		},
		{
			name: "empty history",
			clientMock: &afxtest.MockClient{
				DirectoryHistoriesResult: []string{},
			},
			finderMock: &afxtest.MockFinder{},
		},
		{
			name: "error from histories",
			clientMock: &afxtest.MockClient{
				DirectoryHistoriesErr: errors.New("histories error"),
			},
			finderMock:  &afxtest.MockFinder{},
			expectErr:   true,
			expectedErr: "履歴の取得に失敗しました: histories error",
		},
		{
			name: "error from excd",
			clientMock: &afxtest.MockClient{
				DirectoryHistoriesResult: []string{"C:\\Windows"},
				ChangeDirectoryErr:       errors.New("excd error"),
			},
			finderMock: &afxtest.MockFinder{
				Idx: 0,
			},
			expectErr:   true,
			expectedErr: "ディレクトリ移動に失敗しました: excd error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := run(tt.clientMock, tt.finderMock, nil)

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected an error, but got none")
				} else if err.Error() != tt.expectedErr {
					t.Errorf("expected error %q, got %q", tt.expectedErr, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.clientMock.ChangedDirectory != tt.expectedPath {
				t.Errorf("expected directory %q, got %q", tt.expectedPath, tt.clientMock.ChangedDirectory)
			}
		})
	}
}

func TestRun_WinsAffectsHistoryResults(t *testing.T) {
	tests := []struct {
		name         string
		wins         []int
		expectedPath string
	}{
		{
			name:         "left only",
			wins:         []int{afx.WindowLeft},
			expectedPath: "C:\\Left",
		},
		{
			name:         "right only",
			wins:         []int{afx.WindowRight},
			expectedPath: "C:\\Right",
		},
		{
			name:         "both windows uses first entry",
			wins:         []int{afx.WindowLeft, afx.WindowRight},
			expectedPath: "C:\\Left",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientMock := &afxtest.MockClient{
				HistoriesByWindow: map[int][]string{
					afx.WindowLeft:  {"C:\\Left"},
					afx.WindowRight: {"C:\\Right"},
				},
			}
			finderMock := &afxtest.MockFinder{Idx: 0}

			err := run(clientMock, finderMock, tt.wins)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(clientMock.RequestedWindows, tt.wins) {
				t.Errorf("expected wins %v, got %v", tt.wins, clientMock.RequestedWindows)
			}
			if clientMock.ChangedDirectory != tt.expectedPath {
				t.Errorf("expected directory %q, got %q", tt.expectedPath, clientMock.ChangedDirectory)
			}
		})
	}
}

func TestParseWindowFlag(t *testing.T) {
	tests := []struct {
		name        string
		window      string
		expectedErr bool
		expected    []int
	}{
		{"left", "left", false, []int{afx.WindowLeft}},
		{"right", "right", false, []int{afx.WindowRight}},
		{"both", "both", false, []int{afx.WindowLeft, afx.WindowRight}},
		{"invalid", "invalid", true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wins, err := parseWindowFlag(tt.window)
			if tt.expectedErr {
				if err == nil {
					t.Errorf("expected error for window=%s, but got none", tt.window)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !reflect.DeepEqual(wins, tt.expected) {
					t.Errorf("expected %v, got %v", tt.expected, wins)
				}
			}
		})
	}
}

func TestRun_WithDuplicates(t *testing.T) {
	// 左右のウィンドウで重複する履歴がある場合のテスト
	clientMock := &afxtest.MockClient{
		DirectoryHistoriesResult: []string{"C:\\Windows", "C:\\Users", "C:\\Windows", "C:\\Temp"},
	}
	finderMock := &afxtest.MockFinder{Idx: 1} // "C:\\Users"を選択

	err := run(clientMock, finderMock, []int{afx.WindowLeft, afx.WindowRight})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 重複が除去されているため、インデックス1は"C:\\Users"を指す
	if clientMock.ChangedDirectory != "C:\\Users" {
		t.Errorf("expected directory %q, got %q", "C:\\Users", clientMock.ChangedDirectory)
	}
}
