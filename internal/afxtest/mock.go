package afxtest

import "github.com/tana9/afxw-tools/internal/afx"

// MockAFX は afx.AFX インターフェースのテスト用モックです。
type MockAFX struct {
	HistoriesResult []string
	ExcdPath        string
	HistoriesErr    error
	ExcdErr         error
	// HistoriesByWin はウィンドウ番号ごとの履歴を設定します。
	// 設定されている場合、HistoriesResult より優先されます。
	HistoriesByWin map[int][]string
	// ReceivedWins は Histories に渡された wins 引数を記録します。
	ReceivedWins []int
}

// インターフェースの実装を保証するコンパイル時チェック
var _ afx.AFX = (*MockAFX)(nil)

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

func (m *MockAFX) EXCD(path string) error {
	if m.ExcdErr != nil {
		return m.ExcdErr
	}
	m.ExcdPath = path
	return nil
}

func (m *MockAFX) GetActivePath() (string, error) {
	return "", nil
}

func (m *MockAFX) Close() {}
