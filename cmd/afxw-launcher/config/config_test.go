package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if len(cfg.Menu) != 4 {
		t.Errorf("expected 4 menu items, got %d", len(cfg.Menu))
	}

	// 最初のメニュー項目を確認
	if cfg.Menu[0].Name != "フォルダ履歴から選択" {
		t.Errorf("unexpected first menu name: %s", cfg.Menu[0].Name)
	}

	if cfg.Menu[0].Command != "afxw-his.exe" {
		t.Errorf("unexpected first menu command: %s", cfg.Menu[0].Command)
	}
}

func TestLoad_DefaultWhenNoFile(t *testing.T) {
	// 設定ファイルが存在しない環境で実行
	// デフォルト設定が返されることを確認
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defaultCfg := DefaultConfig()
	if !reflect.DeepEqual(cfg.Menu, defaultCfg.Menu) {
		t.Error("expected default config when no file exists")
	}
}

func TestLoadFrom(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	configContent := `
[[menu]]
name = "Test Command"
description = "Test Description"
command = "test.exe"
args = ["--flag"]

[settings]
tool_dir = "C:\\tools"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	cfg, err := LoadFrom(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Menu) != 1 {
		t.Fatalf("expected 1 menu item, got %d", len(cfg.Menu))
	}
	if cfg.Menu[0].Name != "Test Command" {
		t.Errorf("expected name %q, got %q", "Test Command", cfg.Menu[0].Name)
	}
	if cfg.Menu[0].Command != "test.exe" {
		t.Errorf("expected command %q, got %q", "test.exe", cfg.Menu[0].Command)
	}
	if len(cfg.Menu[0].Args) != 1 || cfg.Menu[0].Args[0] != "--flag" {
		t.Errorf("unexpected args: %v", cfg.Menu[0].Args)
	}
	if cfg.Settings.ToolDir != `C:\tools` {
		t.Errorf("expected tool_dir %q, got %q", `C:\tools`, cfg.Settings.ToolDir)
	}
}

func TestLoadFrom_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	if err := os.WriteFile(configPath, []byte("invalid toml [[["), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	_, err := LoadFrom(configPath)
	if err == nil {
		t.Error("expected error for invalid TOML, got nil")
	}
}

func TestFindCommand_AbsolutePath(t *testing.T) {
	cfg := DefaultConfig()

	// 一時ファイルを作成
	tmpFile := filepath.Join(t.TempDir(), "test.exe")
	if err := os.WriteFile(tmpFile, []byte{}, 0755); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// 絶対パスで検索
	found, err := cfg.FindCommand(tmpFile)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if found != tmpFile {
		t.Errorf("expected %s, got %s", tmpFile, found)
	}
}

func TestFindCommand_NotFound(t *testing.T) {
	cfg := DefaultConfig()

	// 存在しないコマンドを検索
	_, err := cfg.FindCommand("nonexistent-command-12345.exe")
	if err == nil {
		t.Error("expected error for nonexistent command")
	}
}

func TestFindCommand_WithToolDir(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.exe")
	if err := os.WriteFile(tmpFile, []byte{}, 0755); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cfg := &Config{
		Settings: Settings{
			ToolDir: tmpDir,
		},
	}

	found, err := cfg.FindCommand("test.exe")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if found != tmpFile {
		t.Errorf("expected %s, got %s", tmpFile, found)
	}
}
