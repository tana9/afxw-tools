package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if len(cfg.Programs) == 0 {
		t.Error("デフォルト設定にプログラムが存在しません")
	}
	for i, p := range cfg.Programs {
		if p.Name == "" {
			t.Errorf("Programs[%d].Name が空です", i)
		}
		if p.Command == "" {
			t.Errorf("Programs[%d].Command が空です", i)
		}
	}
}

func TestLoadFrom(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[[program]]
name = "VSCode"
description = "Visual Studio Codeで開く"
command = "code.exe"
args = []

[[program]]
name = "メモ帳"
command = "notepad.exe"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("テスト用設定ファイルの作成に失敗しました: %v", err)
	}

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom() エラー: %v", err)
	}

	if len(cfg.Programs) != 2 {
		t.Fatalf("Programs の件数: got %d, want 2", len(cfg.Programs))
	}

	if cfg.Programs[0].Name != "VSCode" {
		t.Errorf("Programs[0].Name: got %q, want %q", cfg.Programs[0].Name, "VSCode")
	}
	if cfg.Programs[0].Description != "Visual Studio Codeで開く" {
		t.Errorf("Programs[0].Description: got %q, want %q", cfg.Programs[0].Description, "Visual Studio Codeで開く")
	}
	if cfg.Programs[0].Command != "code.exe" {
		t.Errorf("Programs[0].Command: got %q, want %q", cfg.Programs[0].Command, "code.exe")
	}
	if cfg.Programs[1].Name != "メモ帳" {
		t.Errorf("Programs[1].Name: got %q, want %q", cfg.Programs[1].Name, "メモ帳")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
	}{
		{name: "empty name", cfg: Config{Programs: []Program{{Name: " ", Command: "tool.exe"}}}},
		{name: "empty command", cfg: Config{Programs: []Program{{Name: "Tool", Command: "\t"}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cfg.Validate(); err == nil {
				t.Fatal("Validate() error = nil, want error")
			}
		})
	}

	if err := DefaultConfig().Validate(); err != nil {
		t.Fatalf("DefaultConfig().Validate() error = %v", err)
	}
}

func TestLoadFrom_InvalidValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("[[program]]\nname = \" \"\ncommand = \"tool.exe\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadFrom(path); err == nil {
		t.Fatal("LoadFrom() error = nil, want validation error")
	}
}

func TestLoadFrom_NotFound(t *testing.T) {
	_, err := LoadFrom("/non/existent/path.toml")
	if err == nil {
		t.Error("存在しないファイルでエラーが発生しませんでした")
	}
}

func TestLoad_CreatesDefaultConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	localPath := filepath.Join(t.TempDir(), "afxw-open.toml")

	cfg, err := load(configPath, localPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Programs) != len(DefaultConfig().Programs) {
		t.Errorf("got %d programs, want %d", len(cfg.Programs), len(DefaultConfig().Programs))
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("default config was not created: %v", err)
	}
}

func TestLoad_UserConfigTakesPriority(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	localPath := filepath.Join(t.TempDir(), "afxw-open.toml")
	if err := os.WriteFile(configPath, []byte("[[program]]\nname = \"User\"\ncommand = \"user.exe\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(localPath, []byte("[[program]]\nname = \"Local\"\ncommand = \"local.exe\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := load(configPath, localPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Programs) != 1 || cfg.Programs[0].Command != "user.exe" {
		t.Errorf("unexpected programs: %+v", cfg.Programs)
	}
}

func TestLoad_UsesLocalConfigWhenUserConfigMissing(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	localPath := filepath.Join(t.TempDir(), "afxw-open.toml")
	if err := os.WriteFile(localPath, []byte("[[program]]\nname = \"Local\"\ncommand = \"local.exe\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := load(configPath, localPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Programs) != 1 || cfg.Programs[0].Command != "local.exe" {
		t.Errorf("unexpected programs: %+v", cfg.Programs)
	}
}

func TestLoad_InvalidUserConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	localPath := filepath.Join(t.TempDir(), "afxw-open.toml")
	if err := os.WriteFile(configPath, []byte("invalid toml [[["), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := load(configPath, localPath); err == nil {
		t.Error("expected error for invalid user config")
	}
}
