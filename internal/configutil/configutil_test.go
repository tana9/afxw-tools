package configutil

import (
	"os"
	"path/filepath"
	"strings"
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

func TestWriteEncodeErrorKeepsExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	want := []byte("name = \"original\"\n")
	if err := os.WriteFile(path, want, 0644); err != nil {
		t.Fatal(err)
	}

	if err := Write(path, map[string]any{"unsupported": make(chan int)}); err == nil {
		t.Fatal("Write() error = nil; エンコードできない値はエラーを期待")
	}
	assertFileAndNoTemporaryFiles(t, path, want)
}

func TestAppendEncodeErrorKeepsExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	want := []byte("name = \"original\"\n")
	if err := os.WriteFile(path, want, 0644); err != nil {
		t.Fatal(err)
	}

	if err := Append(path, map[string]any{"unsupported": make(chan int)}); err == nil {
		t.Fatal("Append() error = nil; エンコードできない値はエラーを期待")
	}
	assertFileAndNoTemporaryFiles(t, path, want)
}

func assertFileAndNoTemporaryFiles(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Errorf("file content = %q, want %q", got, want)
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "."+filepath.Base(path)+".tmp-") {
			t.Errorf("temporary file remains: %s", entry.Name())
		}
	}
}
