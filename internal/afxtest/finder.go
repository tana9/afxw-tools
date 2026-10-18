package afxtest

// MockFinder は finder.Finder インターフェースのテスト用モックです。
type MockFinder struct {
	Idx int
	Err error
}

// Find は設定された選択インデックスとエラーを返します。
func (m *MockFinder) Find(items []string) (int, error) {
	return m.Idx, m.Err
}
