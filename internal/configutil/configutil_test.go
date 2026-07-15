package configutil

import (
	"path/filepath"
	"testing"
)

type testConfig struct {
	Name string `toml:"name"`
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	if exists, err := Exists(path); err != nil || exists {
		t.Fatalf("Exists() = %v, %v; ファイルが無い場合は false, nil を期待", exists, err)
	}

	if err := Write(path, &testConfig{Name: "a"}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if exists, err := Exists(path); err != nil || !exists {
		t.Fatalf("Exists() = %v, %v; ファイルがある場合は true, nil を期待", exists, err)
	}
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

	if _, err := LoadFrom[testConfig](path); err == nil {
		t.Fatal("LoadFrom() error = nil; 存在しないファイルはエラーを期待")
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
