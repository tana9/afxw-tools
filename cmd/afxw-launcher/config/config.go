package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/tana9/afxw-tools/internal/cmdutil"
)

type MenuItem struct {
	Name        string   `toml:"name"`
	Description string   `toml:"description"`
	Command     string   `toml:"command"`
	Args        []string `toml:"args"`
}

type Settings struct {
	ToolDir string `toml:"tool_dir"`
}

type Config struct {
	Menu     []MenuItem `toml:"menu"`
	Settings Settings   `toml:"settings"`
}

func DefaultConfig() *Config {
	return &Config{
		Menu: []MenuItem{
			{Name: "フォルダ履歴から選択", Description: "あふwのフォルダ履歴から選択して移動", Command: "afxw-his.exe", Args: []string{}},
			{Name: "zoxideから選択", Description: "zoxideのfrecencyデータベースから選択して移動", Command: "afxw-zox.exe", Args: []string{}},
			{Name: "ブックマークから選択", Description: "ブックマークから選択して移動", Command: "afxw-bm.exe", Args: []string{}},
			{Name: "ブックマークを追加", Description: "現在のディレクトリをブックマークに追加", Command: "afxw-bm.exe", Args: []string{"-a", ""}},
			openMenuItem(),
		},
		Settings: Settings{ToolDir: ""},
	}
}

// Validate checks values that would otherwise fail only when a menu is used.
func (c *Config) Validate() error {
	for i, item := range c.Menu {
		if strings.TrimSpace(item.Name) == "" {
			return fmt.Errorf("menu[%d].name が空です", i)
		}
		if strings.TrimSpace(item.Command) == "" {
			return fmt.Errorf("menu[%d].command が空です", i)
		}
	}

	toolDir := strings.TrimSpace(c.Settings.ToolDir)
	if toolDir == "" {
		return nil
	}
	info, err := os.Stat(toolDir)
	if err != nil {
		return fmt.Errorf("settings.tool_dir %q を確認できません: %w", c.Settings.ToolDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("settings.tool_dir %q はディレクトリではありません", c.Settings.ToolDir)
	}
	return nil
}

func (c *Config) FindCommand(command string) (string, error) {
	if c.Settings.ToolDir != "" {
		return cmdutil.Find(command, c.Settings.ToolDir)
	}
	return cmdutil.Find(command)
}

// openMenuItem は「ファイルを開く」の標準メニュー項目を返します。
// DefaultConfigと既存ユーザー設定への移行処理の両方から参照するため、メニュー順に依存しない独立した関数にしています。
func openMenuItem() MenuItem {
	return MenuItem{Name: "ファイルを開く", Description: "選択したファイルをプログラムで開く", Command: "afxw-open.exe", Args: []string{"{files}"}}
}
