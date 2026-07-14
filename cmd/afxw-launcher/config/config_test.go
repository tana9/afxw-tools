package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if len(cfg.Menu) != 5 {
		t.Errorf("expected 5 menu items, got %d", len(cfg.Menu))
	}

	if cfg.Menu[0].Name != "フォルダ履歴から選択" {
		t.Errorf("unexpected first menu name: %s", cfg.Menu[0].Name)
	}

	if cfg.Menu[0].Command != "afxw-his.exe" {
		t.Errorf("unexpected first menu command: %s", cfg.Menu[0].Command)
	}

	open := cfg.Menu[4]
	if open.Command != "afxw-open.exe" {
		t.Errorf("unexpected open command: %s", open.Command)
	}
	if len(open.Args) != 1 || open.Args[0] != "{files}" {
		t.Errorf("unexpected open args: %v", open.Args)
	}
}

func TestLoad_CreatesDefaultConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	localPath := filepath.Join(t.TempDir(), "local-config.toml")

	cfg, err := load(configPath, localPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Menu) != 5 {
		t.Errorf("expected 5 menu items, got %d", len(cfg.Menu))
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("expected default config to be created: %v", err)
	}
}

func TestLoad_MigratesExistingUserConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	localPath := filepath.Join(t.TempDir(), "local-config.toml")
	legacyConfig := `
# このコメントは保持する
[[menu]]
name = "履歴"
command = "afxw-his.exe"
args = []
`
	if err := os.WriteFile(configPath, []byte(legacyConfig), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := load(configPath, localPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Menu) != 2 {
		t.Fatalf("expected 2 menu items, got %d", len(cfg.Menu))
	}
	if cfg.Menu[1].Command != "afxw-open.exe" {
		t.Errorf("expected open command, got %q", cfg.Menu[1].Command)
	}

	persisted, err := LoadFrom(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(persisted.Menu) != 2 {
		t.Errorf("expected migrated config to be persisted, got %d items", len(persisted.Menu))
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "# このコメントは保持する") {
		t.Error("expected existing comments to be preserved")
	}
}

func TestLoad_PreservesLocalConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	localPath := filepath.Join(t.TempDir(), "local-config.toml")
	localConfig := `
[[menu]]
name = "独自コマンド"
command = "custom.exe"
args = []
`
	if err := os.WriteFile(localPath, []byte(localConfig), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := load(configPath, localPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Menu) != 1 || cfg.Menu[0].Command != "custom.exe" {
		t.Errorf("unexpected local config: %+v", cfg.Menu)
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

	tmpFile := filepath.Join(t.TempDir(), "test.exe")
	if err := os.WriteFile(tmpFile, []byte{}, 0755); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

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
