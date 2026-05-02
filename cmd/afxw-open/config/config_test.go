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

func TestLoadFrom_NotFound(t *testing.T) {
	_, err := LoadFrom("/non/existent/path.toml")
	if err == nil {
		t.Error("存在しないファイルでエラーが発生しませんでした")
	}
}
