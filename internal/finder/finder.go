package finder

import "github.com/ktr0731/go-fuzzyfinder"

// Finder はアイテムの検索と選択を行うためのインターフェースを定義します。
type Finder interface {
	Find(items []string) (int, error)
}

// FuzzyFinder は go-fuzzyfinder を用いて Finder を実装します。
type FuzzyFinder struct{}

// Find はファジーファインダーで項目を選択させ、選択された項目のインデックスを返します。
func (f *FuzzyFinder) Find(items []string) (int, error) {
	return fuzzyfinder.Find(items, func(i int) string {
		return items[i]
	})
}
