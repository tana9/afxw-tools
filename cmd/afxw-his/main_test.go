package main

import (
	"errors"
	"reflect"
	"testing"

	"github.com/tana9/afxw-tools/internal/afx"
	"github.com/tana9/afxw-tools/internal/afxtest"
	"github.com/tana9/afxw-tools/internal/finder"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name             string
		afxMock          *afxtest.MockAFX
		finderMock       *finder.MockFinder
		expectErr        bool
		expectedErr      string
		expectedExcdPath string
	}{
		{
			name: "normal run",
			afxMock: &afxtest.MockAFX{
				HistoriesResult: []string{"C:\\Windows", "C:\\Users"},
			},
			finderMock: &finder.MockFinder{
				Idx: 0,
			},
			expectedExcdPath: "C:\\Windows",
		},
		{
			name: "normal run with selection",
			afxMock: &afxtest.MockAFX{
				HistoriesResult: []string{"C:\\Windows", "C:\\Users"},
			},
			finderMock: &finder.MockFinder{
				Idx: 1,
			},
			expectedExcdPath: "C:\\Users",
		},
		{
			name: "finder cancelled",
			afxMock: &afxtest.MockAFX{
				HistoriesResult: []string{"C:\\Windows", "C:\\Users"},
			},
			finderMock: &finder.MockFinder{
				Err: errors.New("fuzzyfinder cancelled"),
			},
			expectErr:   true,
			expectedErr: "fuzzyfinder cancelled",
		},
		{
			name: "empty history",
			afxMock: &afxtest.MockAFX{
				HistoriesResult: []string{},
			},
			finderMock: &finder.MockFinder{},
		},
		{
			name: "error from histories",
			afxMock: &afxtest.MockAFX{
				HistoriesErr: errors.New("histories error"),
			},
			finderMock:  &finder.MockFinder{},
			expectErr:   true,
			expectedErr: "履歴の取得に失敗しました: histories error",
		},
		{
			name: "error from excd",
			afxMock: &afxtest.MockAFX{
				HistoriesResult: []string{"C:\\Windows"},
				ExcdErr:         errors.New("excd error"),
			},
			finderMock: &finder.MockFinder{
				Idx: 0,
			},
			expectErr:   true,
			expectedErr: "ディレクトリ移動に失敗しました: excd error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := run(tt.afxMock, tt.finderMock, nil)

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected an error, but got none")
				} else if err.Error() != tt.expectedErr {
					t.Errorf("expected error %q, got %q", tt.expectedErr, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.afxMock.ExcdPath != tt.expectedExcdPath {
				t.Errorf("expected excd path %q, got %q", tt.expectedExcdPath, tt.afxMock.ExcdPath)
			}
		})
	}
}

func TestRun_WinsAffectsHistoryResults(t *testing.T) {
	tests := []struct {
		name             string
		wins             []int
		expectedExcdPath string
	}{
		{
			name:             "left only",
			wins:             []int{afx.WindowLeft},
			expectedExcdPath: "C:\\Left",
		},
		{
			name:             "right only",
			wins:             []int{afx.WindowRight},
			expectedExcdPath: "C:\\Right",
		},
		{
			name:             "both windows uses first entry",
			wins:             []int{afx.WindowLeft, afx.WindowRight},
			expectedExcdPath: "C:\\Left",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			afxMock := &afxtest.MockAFX{
				HistoriesByWin: map[int][]string{
					afx.WindowLeft:  {"C:\\Left"},
					afx.WindowRight: {"C:\\Right"},
				},
			}
			finderMock := &finder.MockFinder{Idx: 0}

			err := run(afxMock, finderMock, tt.wins)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(afxMock.ReceivedWins, tt.wins) {
				t.Errorf("expected wins %v, got %v", tt.wins, afxMock.ReceivedWins)
			}
			if afxMock.ExcdPath != tt.expectedExcdPath {
				t.Errorf("expected excd path %q, got %q", tt.expectedExcdPath, afxMock.ExcdPath)
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

func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no duplicates",
			input:    []string{"C:\\Windows", "C:\\Users", "C:\\Temp"},
			expected: []string{"C:\\Windows", "C:\\Users", "C:\\Temp"},
		},
		{
			name:     "with duplicates",
			input:    []string{"C:\\Windows", "C:\\Users", "C:\\Windows", "C:\\Temp"},
			expected: []string{"C:\\Windows", "C:\\Users", "C:\\Temp"},
		},
		{
			name:     "all duplicates",
			input:    []string{"C:\\Windows", "C:\\Windows", "C:\\Windows"},
			expected: []string{"C:\\Windows"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "single element",
			input:    []string{"C:\\Windows"},
			expected: []string{"C:\\Windows"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeDuplicates(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRun_WithDuplicates(t *testing.T) {
	// 左右のウィンドウで重複する履歴がある場合のテスト
	afxMock := &afxtest.MockAFX{
		HistoriesResult: []string{"C:\\Windows", "C:\\Users", "C:\\Windows", "C:\\Temp"},
	}
	finderMock := &finder.MockFinder{Idx: 1} // "C:\\Users"を選択

	err := run(afxMock, finderMock, []int{afx.WindowLeft, afx.WindowRight})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 重複が除去されているため、インデックス1は"C:\\Users"を指す
	if afxMock.ExcdPath != "C:\\Users" {
		t.Errorf("expected excd path %q, got %q", "C:\\Users", afxMock.ExcdPath)
	}
}
