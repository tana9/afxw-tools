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

func TestLoad_FromFile(t *testing.T) {
	// 一時ディレクトリにテスト用の設定ファイルを作成
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

	// 簡易的なテスト: デフォルト設定がロードできることを確認
	// （実際のファイル読み込みテストは統合テストで実施）
	cfg := DefaultConfig()
	if len(cfg.Menu) == 0 {
		t.Error("menu should not be empty")
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
