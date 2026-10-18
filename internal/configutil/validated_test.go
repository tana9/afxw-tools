package configutil

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type validatedConfig struct {
	Name string `toml:"name"`
}

// Validate は名前が空の場合に os.ErrNotExist をラップしたエラーを返します。
// 検証エラーがファイル不在と誤判定されないことを確認するためのテスト用実装です。
func (c *validatedConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name が空です: %w", os.ErrNotExist)
	}
	return nil
}

func TestTryLoadValidated_MissingFile(t *testing.T) {
	cfg, err := TryLoadValidated[validatedConfig](filepath.Join(t.TempDir(), "missing.toml"))
	if err != nil || cfg != nil {
		t.Fatalf("TryLoadValidated() = (%v, %v), want (nil, nil)", cfg, err)
	}
}

func TestTryLoadValidated_ValidationErrorIsNotTreatedAsMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("name = \"\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := TryLoadValidated[validatedConfig](path)
	if err == nil {
		t.Fatalf("TryLoadValidated() = (%v, nil), want validation error", cfg)
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("検証エラーの原因が保持されていません: %v", err)
	}
}

func TestLoadValidated(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := Write(path, &validatedConfig{Name: "afxw"}); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadValidated[validatedConfig](path)
	if err != nil {
		t.Fatalf("LoadValidated() error = %v", err)
	}
	if cfg.Name != "afxw" {
		t.Errorf("cfg.Name = %q, want %q", cfg.Name, "afxw")
	}
}
