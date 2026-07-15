package configutil

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Exists は指定パスの設定ファイルが存在するかどうかを返します。
// 存在確認自体が失敗した場合（権限エラー等、ファイル不在以外）はエラーを返します。
func Exists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("設定ファイルの確認に失敗しました (%s): %w", path, err)
	}
	return false, nil
}

// LoadFrom は指定パスのTOMLファイルをデコードして T の値を返します。
func LoadFrom[T any](path string) (*T, error) {
	var cfg T
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("設定ファイルの読み込みに失敗しました (%s): %w", path, err)
	}
	return &cfg, nil
}

// Write は cfg をTOML形式でpathに書き込みます。親ディレクトリが無ければ作成します。
// 既存ファイルがある場合は上書きします。
func Write(path string, cfg any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("ファイルの作成に失敗しました: %w", err)
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		return fmt.Errorf("設定の書き込みに失敗しました: %w", err)
	}
	return nil
}

// Append は既存の設定ファイルに v をTOMLとして追記します。
// 追記の前に改行を1つ挟むため、既存内容と区切られた新しいテーブルの追加に使えます。
func Append(path string, v any) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("ファイルのオープンに失敗しました: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString("\n"); err != nil {
		return fmt.Errorf("改行の書き込みに失敗しました: %w", err)
	}
	if err := toml.NewEncoder(f).Encode(v); err != nil {
		return fmt.Errorf("設定の書き込みに失敗しました: %w", err)
	}
	return nil
}
