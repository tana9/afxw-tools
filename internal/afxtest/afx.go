package afxtest

import "github.com/tana9/afxw-tools/internal/afx"

// MockAFX は afx.AFX インターフェースのテスト用モックです。
type MockAFX struct {
	HistoriesResult   []string
	ExcdPath          string
	ExcdFilePath      string
	HistoriesErr      error
	ExcdErr           error
	ExcdFileErr       error
	CurrentFileResult string
	CurrentFileErr    error
	MarkedFilesResult []string
	MarkedFilesErr    error
	// HistoriesByWin はウィンドウ番号ごとの履歴を設定します。
	// 設定されている場合、HistoriesResult より優先されます。
	HistoriesByWin map[int][]string
	// ReceivedWins は Histories に渡された wins 引数を記録します。
	ReceivedWins []int
}

// インターフェースの実装を保証するコンパイル時チェック
var _ afx.AFX = (*MockAFX)(nil)

// Histories は設定された履歴を返し、渡された wins を ReceivedWins に記録します。
func (m *MockAFX) Histories(wins []int) ([]string, error) {
	m.ReceivedWins = wins
	if m.HistoriesErr != nil {
		return nil, m.HistoriesErr
	}
	if m.HistoriesByWin != nil {
		var dirs []string
		for _, win := range wins {
			dirs = append(dirs, m.HistoriesByWin[win]...)
		}
		return dirs, nil
	}
	return m.HistoriesResult, nil
}

// EXCD はエラー設定が無ければ渡されたパスを ExcdPath に記録します。
func (m *MockAFX) EXCD(path string) error {
	if m.ExcdErr != nil {
		return m.ExcdErr
	}
	m.ExcdPath = path
	return nil
}

// EXCDFile はエラー設定が無ければ渡されたパスを ExcdFilePath に記録します。
func (m *MockAFX) EXCDFile(path string) error {
	if m.ExcdFileErr != nil {
		return m.ExcdFileErr
	}
	m.ExcdFilePath = path
	return nil
}

// GetActivePath は常に空のパスを返します。
func (m *MockAFX) GetActivePath() (string, error) {
	return "", nil
}

// GetCurrentFile は設定されたカーソル位置のファイルとエラーを返します。
func (m *MockAFX) GetCurrentFile() (string, error) {
	if m.CurrentFileErr != nil {
		return "", m.CurrentFileErr
	}
	return m.CurrentFileResult, nil
}

// GetMarkedFiles は設定されたマーク済みファイルとエラーを返します。
func (m *MockAFX) GetMarkedFiles() ([]string, error) {
	if m.MarkedFilesErr != nil {
		return nil, m.MarkedFilesErr
	}
	return m.MarkedFilesResult, nil
}

// Close は何もしません。
func (m *MockAFX) Close() {}
