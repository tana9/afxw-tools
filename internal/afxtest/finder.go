package afxtest

// MockFinder は finder.Finder インターフェースのテスト用モックです。
type MockFinder struct {
	Idx int
	Err error
}

func (m *MockFinder) Find(items []string) (int, error) {
	return m.Idx, m.Err
}
