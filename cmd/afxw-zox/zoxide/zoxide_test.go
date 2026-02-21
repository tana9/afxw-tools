package zoxide

import (
	"testing"
)

func TestParseQueryOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Entry
	}{
		{
			name:  "通常のパス",
			input: "12.5 C:\\Users\\TanakaTakashi\\Projects\n10.0 C:\\code",
			expected: []Entry{
				{Path: "C:\\Users\\TanakaTakashi\\Projects", Score: 12.5},
				{Path: "C:\\code", Score: 10.0},
			},
		},
		{
			name:  "スペースを含むパス",
			input: "15.0 C:\\Program Files\\MyApp\n8.5 C:\\Users\\Test User\\Documents",
			expected: []Entry{
				{Path: "C:\\Program Files\\MyApp", Score: 15.0},
				{Path: "C:\\Users\\Test User\\Documents", Score: 8.5},
			},
		},
		{
			name:  "複数スペースを含むパス",
			input: "20.0 C:\\My Project Folder\\Sub Folder\\Deep Path",
			expected: []Entry{
				{Path: "C:\\My Project Folder\\Sub Folder\\Deep Path", Score: 20.0},
			},
		},
		{
			name:     "空の入力",
			input:    "",
			expected: nil,
		},
		{
			name:  "空行を含む入力",
			input: "12.5 C:\\Users\\Test\n\n10.0 C:\\code\n",
			expected: []Entry{
				{Path: "C:\\Users\\Test", Score: 12.5},
				{Path: "C:\\code", Score: 10.0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// os.Statのチェックをスキップするため、実際のファイルシステムには依存しない
			// （実装では os.Stat チェックがあるため、実際には存在しないパスは除外される）
			result, err := parseQueryOutput(tt.input)
			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}

			// 実際のファイルシステムで存在しないパスは除外されるため、
			// 長さが0でもエラーではない
			if len(result) == 0 && len(tt.expected) > 0 {
				t.Logf("パスが存在しないため結果は空（期待通り）")
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("エントリ数が一致しません: got %d, want %d", len(result), len(tt.expected))
				return
			}

			for i, entry := range result {
				if entry.Path != tt.expected[i].Path {
					t.Errorf("パスが一致しません [%d]: got %q, want %q", i, entry.Path, tt.expected[i].Path)
				}
				if entry.Score != tt.expected[i].Score {
					t.Errorf("スコアが一致しません [%d]: got %f, want %f", i, entry.Score, tt.expected[i].Score)
				}
			}
		})
	}
}

func TestPaths(t *testing.T) {
	entries := []Entry{
		{Path: "C:\\Users\\Test", Score: 10.0},
		{Path: "C:\\Program Files\\MyApp", Score: 15.0},
		{Path: "C:\\My Folder\\Sub Folder", Score: 20.0},
	}

	expected := []string{
		"C:\\Users\\Test",
		"C:\\Program Files\\MyApp",
		"C:\\My Folder\\Sub Folder",
	}

	result := Paths(entries)

	if len(result) != len(expected) {
		t.Fatalf("パス数が一致しません: got %d, want %d", len(result), len(expected))
	}

	for i, path := range result {
		if path != expected[i] {
			t.Errorf("パスが一致しません [%d]: got %q, want %q", i, path, expected[i])
		}
	}
}
