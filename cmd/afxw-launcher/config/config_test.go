package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// menuCommands はメニューからコマンド名を順番に取り出します。
func menuCommands(items []MenuItem) []string {
	commands := make([]string, len(items))
	for i, item := range items {
		commands[i] = item.Command
	}
	return commands
}

// assertMenuCommands はメニューのコマンドと順序を検証します。
func assertMenuCommands(t *testing.T, items []MenuItem, want []string) {
	t.Helper()
	if got := menuCommands(items); !slices.Equal(got, want) {
		t.Fatalf("menu commands = %v, want %v", got, want)
	}
}

// findMenuByCommand はコマンド名に一致するメニューを返します。
func findMenuByCommand(items []MenuItem, command string) (MenuItem, bool) {
	for _, item := range items {
		if item.Command == command {
			return item, true
		}
	}
	return MenuItem{}, false
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assertMenuCommands(t, cfg.Menu, []string{
		"afxw-his.exe", "afxw-zox.exe", "afxw-bm.exe", "afxw-bm.exe",
		"afxw-open.exe", "afxw-wt.exe", "afxw-rg.exe", "afxw-diff.exe",
	})

	open, ok := findMenuByCommand(cfg.Menu, "afxw-open.exe")
	if !ok {
		t.Fatal("afxw-open.exe menu not found")
	}
	if len(open.Args) != 1 || open.Args[0] != "{files}" {
		t.Errorf("unexpected open args: %v", open.Args)
	}
	for _, command := range []string{"afxw-wt.exe", "afxw-rg.exe", "afxw-diff.exe"} {
		item, found := findMenuByCommand(cfg.Menu, command)
		if !found || len(item.Args) != 0 {
			t.Errorf("unexpected %s menu: %+v", command, item)
		}
	}
}

func TestLoad_CreatesDefaultConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	localPath := filepath.Join(t.TempDir(), "local-config.toml")

	cfg, err := load(configPath, localPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertMenuCommands(t, cfg.Menu, menuCommands(DefaultConfig().Menu))
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
	wantCommands := append([]string{"afxw-his.exe"}, menuCommands(standardMenuItems())...)
	assertMenuCommands(t, cfg.Menu, wantCommands)

	persisted, err := LoadFrom(configPath)
	if err != nil {
		t.Fatal(err)
	}
	assertMenuCommands(t, persisted.Menu, wantCommands)
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

func TestLoad_DoesNotDuplicateStandardMenus(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	localPath := filepath.Join(t.TempDir(), "local-config.toml")
	content := `
[[menu]]
name = "ファイルを開く"
command = "AFXW-OPEN.EXE"
args = ["{files}"]

[[menu]]
name = "Terminal"
command = "tools/afxw-wt.exe"
args = []

[[menu]]
name = "Search"
command = "AFXW-RG.EXE"
args = []

[[menu]]
name = "Diff"
command = "tools/AFXW-DIFF.EXE"
args = []
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := load(configPath, localPath)
	if err != nil {
		t.Fatal(err)
	}
	assertMenuCommands(t, cfg.Menu, []string{"AFXW-OPEN.EXE", "tools/afxw-wt.exe", "AFXW-RG.EXE", "tools/AFXW-DIFF.EXE"})
}

func TestLoad_MigratesOnlyMissingWindowsTerminalMenu(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	localPath := filepath.Join(t.TempDir(), "local-config.toml")
	content := `
# 既存設定を保持する
[[menu]]
name = "ファイルを開く"
command = "afxw-open.exe"
args = ["{files}"]

[[menu]]
name = "キーワード検索"
command = "afxw-rg.exe"
args = []

[[menu]]
name = "WinMergeで比較"
command = "afxw-diff.exe"
args = []
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := load(configPath, localPath)
	if err != nil {
		t.Fatal(err)
	}
	wantCommands := []string{"afxw-open.exe", "afxw-rg.exe", "afxw-diff.exe", "afxw-wt.exe"}
	assertMenuCommands(t, cfg.Menu, wantCommands)

	persisted, err := LoadFrom(configPath)
	if err != nil {
		t.Fatal(err)
	}
	assertMenuCommands(t, persisted.Menu, wantCommands)
	persistedContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(persistedContent), "# 既存設定を保持する") {
		t.Error("existing comments were not preserved")
	}
}

func TestLoadFrom(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	toolDir := filepath.Join(tmpDir, "tools")
	if err := os.Mkdir(toolDir, 0755); err != nil {
		t.Fatal(err)
	}

	configContent := fmt.Sprintf(`
[[menu]]
name = "Test Command"
description = "Test Description"
command = "test.exe"
args = ["--flag"]

[settings]
tool_dir = %q
`, filepath.ToSlash(toolDir))

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
	if cfg.Settings.ToolDir != filepath.ToSlash(toolDir) {
		t.Errorf("expected tool_dir %q, got %q", filepath.ToSlash(toolDir), cfg.Settings.ToolDir)
	}
}

func TestValidate(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "tool.exe")
	if err := os.WriteFile(filePath, nil, 0644); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		cfg  Config
	}{
		{name: "empty name", cfg: Config{Menu: []MenuItem{{Name: " ", Command: "tool.exe"}}}},
		{name: "empty command", cfg: Config{Menu: []MenuItem{{Name: "Tool", Command: "\t"}}}},
		{name: "missing tool dir", cfg: Config{Settings: Settings{ToolDir: filepath.Join(t.TempDir(), "missing")}}},
		{name: "tool dir is file", cfg: Config{Settings: Settings{ToolDir: filePath}}},
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
	if err := os.WriteFile(path, []byte("[[menu]]\nname = \"Tool\"\ncommand = \" \"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadFrom(path); err == nil {
		t.Fatal("LoadFrom() error = nil, want validation error")
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
