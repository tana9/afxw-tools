package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/tana9/afxw-tools/internal/cmdutil"
)

// MenuItem はランチャーのメニュー1項目（表示名・説明・実行コマンド・引数）を表します。
type MenuItem struct {
	Name        string   `toml:"name"`
	Description string   `toml:"description"`
	Command     string   `toml:"command"`
	Args        []string `toml:"args"`
}

// Settings は設定ファイルの [settings] セクションを表します。
type Settings struct {
	ToolDir string `toml:"tool_dir"`
}

// Config はランチャーの設定全体（メニュー一覧と設定）を表します。
type Config struct {
	Menu     []MenuItem `toml:"menu"`
	Settings Settings   `toml:"settings"`
}

// DefaultConfig は新規設定ファイルの生成に使う既定の設定を返します。
func DefaultConfig() *Config {
	menu := []MenuItem{
		{Name: "フォルダ履歴から選択", Description: "あふwのフォルダ履歴から選択して移動", Command: "afxw-his.exe", Args: []string{}},
		{Name: "zoxideから選択", Description: "zoxideのfrecencyデータベースから選択して移動", Command: "afxw-zox.exe", Args: []string{}},
		{Name: "ブックマークから選択", Description: "ブックマークから選択して移動", Command: "afxw-bm.exe", Args: []string{}},
		{Name: "ブックマークを追加", Description: "現在のディレクトリをブックマークに追加", Command: "afxw-bm.exe", Args: []string{"-a", ""}},
	}
	menu = append(menu, standardMenuItems()...)
	return &Config{
		Menu:     menu,
		Settings: Settings{ToolDir: ""},
	}
}

// Validate は、メニューを実際に使ったときにしか判明しない不正な値がないかを検証します。
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

// FindCommand はtool_dirが設定されていればそのフォルダも含めて、コマンドの実行ファイルを探します。
func (c *Config) FindCommand(command string) (string, error) {
	if c.Settings.ToolDir != "" {
		return cmdutil.Find(command, c.Settings.ToolDir)
	}
	return cmdutil.Find(command)
}

// standardMenuItems は新規設定と既存設定への移行で共有する標準メニューを返します。
// 呼び出し側が変更しても互いに影響しないよう、呼び出しごとに新しいスライスを返します。
func standardMenuItems() []MenuItem {
	return []MenuItem{
		{Name: "ファイルを開く", Description: "選択したファイルをプログラムで開く", Command: "afxw-open.exe", Args: []string{"{files}"}},
		{Name: "Windows Terminalで開く", Description: "現在のフォルダをWindows Terminalで開く", Command: "afxw-wt.exe", Args: []string{}},
		{Name: "キーワード検索", Description: "現在のフォルダ以下をripgrepで検索", Command: "afxw-rg.exe", Args: []string{}},
		{Name: "WinMergeで比較", Description: "マークした2項目をWinMergeで比較", Command: "afxw-diff.exe", Args: []string{}},
	}
}
