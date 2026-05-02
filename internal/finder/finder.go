package finder

import "github.com/ktr0731/go-fuzzyfinder"

// Finder はアイテムの検索と選択を行うためのインターフェースを定義します。
type Finder interface {
	Find(items []string) (int, error)
}

type FuzzyFinder struct{}

func (f *FuzzyFinder) Find(items []string) (int, error) {
	return fuzzyfinder.Find(items, func(i int) string {
		return items[i]
	})
}
