package afxtest

import "github.com/tana9/afxw-tools/internal/afx"

// MockClient は afx.Client インターフェースのテスト用モックです。
type MockClient struct {
	DirectoryHistoriesResult []string
	ChangedDirectory         string
	DirectoryHistoriesErr    error
	ChangeDirectoryErr       error
	ActivePathResult         string
	ActivePathErr            error
	CurrentFileResult        string
	CurrentFileErr           error
	MarkedFilesResult        []string
	MarkedFilesErr           error
	// HistoriesByWindow はウィンドウ番号ごとの履歴を設定します。
	// 設定されている場合、DirectoryHistoriesResult より優先されます。
	HistoriesByWindow map[int][]string
	// RequestedWindows は DirectoryHistories に渡された引数を記録します。
	RequestedWindows []int
}

// インターフェースの実装を保証するコンパイル時チェック
var _ afx.Client = (*MockClient)(nil)

func (m *MockClient) DirectoryHistories(windows []int) ([]string, error) {
	m.RequestedWindows = windows
	if m.DirectoryHistoriesErr != nil {
		return nil, m.DirectoryHistoriesErr
	}
	if m.HistoriesByWindow != nil {
		var dirs []string
		for _, window := range windows {
			dirs = append(dirs, m.HistoriesByWindow[window]...)
		}
		return dirs, nil
	}
	return m.DirectoryHistoriesResult, nil
}

func (m *MockClient) ChangeDirectory(path string) error {
	if m.ChangeDirectoryErr != nil {
		return m.ChangeDirectoryErr
	}
	m.ChangedDirectory = path
	return nil
}

func (m *MockClient) ActivePath() (string, error) {
	return m.ActivePathResult, m.ActivePathErr
}

func (m *MockClient) CurrentFile() (string, error) {
	if m.CurrentFileErr != nil {
		return "", m.CurrentFileErr
	}
	return m.CurrentFileResult, nil
}

func (m *MockClient) MarkedFiles() ([]string, error) {
	if m.MarkedFilesErr != nil {
		return nil, m.MarkedFilesErr
	}
	return m.MarkedFilesResult, nil
}

func (m *MockClient) Close() {}
