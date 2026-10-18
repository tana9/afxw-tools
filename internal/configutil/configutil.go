package configutil

import (
	"errors"
	"fmt"
	"io"
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

// TryLoad はTOMLファイルを読み込み、存在しない場合は (nil, nil) を返します。
func TryLoad[T any](path string) (*T, error) {
	cfg, err := LoadFrom[T](path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return cfg, err
}

// validatable は Validate メソッドを持つ *T を表す型制約です。
type validatable[T any] interface {
	*T
	Validate() error
}

// LoadValidated はTOMLファイルを読み込み、Validateで検証した値を返します。
func LoadValidated[T any, PT validatable[T]](path string) (*T, error) {
	cfg, err := LoadFrom[T](path)
	if err != nil {
		return nil, err
	}
	return validate[T, PT](path, cfg)
}

// TryLoadValidated はTOMLファイルを読み込んで検証し、ファイルが存在しない場合は (nil, nil) を返します。
// 検証エラーがos.ErrNotExistをラップしていてもファイル不在と誤判定しないよう、不在判定は読み込み時のみ行います。
func TryLoadValidated[T any, PT validatable[T]](path string) (*T, error) {
	cfg, err := TryLoad[T](path)
	if err != nil || cfg == nil {
		return cfg, err
	}
	return validate[T, PT](path, cfg)
}

// validate は読み込んだ設定を検証し、不正な場合はファイルパス付きのエラーを返します。
func validate[T any, PT validatable[T]](path string, cfg *T) (*T, error) {
	if err := PT(cfg).Validate(); err != nil {
		return nil, fmt.Errorf("設定ファイル %q が不正です: %w", path, err)
	}
	return cfg, nil
}

// WriteDefault はデフォルト設定をpathに書き込み、結果を標準エラー出力へ表示します。
// 書き込みに失敗してもデフォルト設定で動作を継続できるよう、エラーは警告表示に留めます。
func WriteDefault(path string, cfg any) {
	if err := Write(path, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "警告: 設定ファイルの作成に失敗しました: %v\n", err)
		return
	}
	fmt.Fprintf(os.Stderr, "デフォルト設定ファイルを作成しました: %s\n", path)
}

// Write は cfg をTOML形式でpathに書き込みます。親ディレクトリが無ければ作成します。
// 既存ファイルがある場合は上書きします。
func Write(path string, cfg any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
	}
	return atomicWrite(path, func(w io.Writer) error {
		if err := toml.NewEncoder(w).Encode(cfg); err != nil {
			return fmt.Errorf("設定の書き込みに失敗しました: %w", err)
		}
		return nil
	})
}

// Append は既存の設定ファイルに v をTOMLとして追記します。
// 追記の前に改行を1つ挟むため、既存内容と区切られた新しいテーブルの追加に使えます。
func Append(path string, v any) error {
	existing, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("ファイルのオープンに失敗しました: %w", err)
	}

	return atomicWrite(path, func(w io.Writer) error {
		if _, err := w.Write(existing); err != nil {
			return fmt.Errorf("既存設定の書き込みに失敗しました: %w", err)
		}
		if _, err := io.WriteString(w, "\n"); err != nil {
			return fmt.Errorf("改行の書き込みに失敗しました: %w", err)
		}
		if err := toml.NewEncoder(w).Encode(v); err != nil {
			return fmt.Errorf("設定の書き込みに失敗しました: %w", err)
		}
		return nil
	})
}
