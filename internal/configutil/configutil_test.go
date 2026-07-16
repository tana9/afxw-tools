package configutil

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type testConfig struct {
	Name string `toml:"name"`
}

func TestWriteAndLoadFrom(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "config.toml")

	if err := Write(path, &testConfig{Name: "afxw"}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	cfg, err := LoadFrom[testConfig](path)
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}
	if cfg.Name != "afxw" {
		t.Errorf("cfg.Name = %q, want %q", cfg.Name, "afxw")
	}
}

func TestLoadFromNotExist(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.toml")

	_, err := LoadFrom[testConfig](path)
	if err == nil {
		t.Fatal("LoadFrom() error = nil; 存在しないファイルはエラーを期待")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("errors.Is(err, os.ErrNotExist) = false; err = %v", err)
	}
}

func TestAppend(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	if err := Write(path, &testConfig{Name: "first"}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	type extra struct {
		Extra string `toml:"extra"`
	}
	if err := Append(path, &extra{Extra: "second"}); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	type combined struct {
		Name  string `toml:"name"`
		Extra string `toml:"extra"`
	}
	cfg, err := LoadFrom[combined](path)
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}
	if cfg.Name != "first" || cfg.Extra != "second" {
		t.Errorf("cfg = %+v, want {Name:first Extra:second}", cfg)
	}
}

func TestAppendNotExist(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.toml")

	if err := Append(path, &testConfig{Name: "a"}); err == nil {
		t.Fatal("Append() error = nil; 存在しないファイルはエラーを期待")
	}
}

func TestTryLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	// 存在しない場合は (nil, nil)
	cfg, err := TryLoad[testConfig](path)
	if err != nil || cfg != nil {
		t.Fatalf("TryLoad() = %v, %v; 存在しないファイルは nil, nil を期待", cfg, err)
	}

	if err := Write(path, &testConfig{Name: "afxw"}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	cfg, err = TryLoad[testConfig](path)
	if err != nil || cfg == nil || cfg.Name != "afxw" {
		t.Fatalf("TryLoad() = %+v, %v; 読み込み成功を期待", cfg, err)
	}
}
