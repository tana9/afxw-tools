package configutil

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// LoadFrom は指定パスのTOMLファイルをデコードして T の値を返します。
func LoadFrom[T any](path string) (*T, error) {
	var cfg T
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("設定ファイルの読み込みに失敗しました (%s): %w", path, err)
	}
	return &cfg, nil
}

// TryLoad は指定パスのTOMLファイルを読み込みます。
// ファイルが存在しない場合はエラーにせず (nil, nil) を返し、呼び出し側が cfg != nil で存在を判定できます。
func TryLoad[T any](path string) (*T, error) {
	cfg, err := LoadFrom[T](path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return cfg, err
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
	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		_ = f.Close()
		return fmt.Errorf("設定の書き込みに失敗しました: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("設定ファイルのクローズに失敗しました: %w", err)
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

	if _, err := f.WriteString("\n"); err != nil {
		_ = f.Close()
		return fmt.Errorf("改行の書き込みに失敗しました: %w", err)
	}
	if err := toml.NewEncoder(f).Encode(v); err != nil {
		_ = f.Close()
		return fmt.Errorf("設定の書き込みに失敗しました: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("設定ファイルのクローズに失敗しました: %w", err)
	}
	return nil
}
