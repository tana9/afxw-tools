package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/tana9/afxw-tools/internal/cmdutil"
	"path/filepath"
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
		},
		Settings: Settings{ToolDir: ""},
	}
}

func LoadFrom(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("設定ファイルの読み込みに失敗しました (%s): %w", path, err)
	}
	return &cfg, nil
}

func Load() (*Config, error) {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".config", "afxw-launcher", "config.toml")
	localPath := filepath.Join(cmdutil.ExecDir(), "config.toml")

	for _, path := range []string{configPath, localPath} {
		if _, err := os.Stat(path); err == nil {
			return LoadFrom(path)
		}
	}

	cfg := DefaultConfig()
	if err := createDefaultConfigFile(configPath, cfg); err != nil {
		// 作成に失敗してもデフォルト設定を返す（エラーにしない）
		fmt.Fprintf(os.Stderr, "警告: 設定ファイルの作成に失敗しました: %v\n", err)
	}
	return cfg, nil
}

func createDefaultConfigFile(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("ファイルの作成に失敗しました: %w", err)
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		return fmt.Errorf("設定の書き込みに失敗しました: %w", err)
	}
	fmt.Fprintf(os.Stderr, "デフォルト設定ファイルを作成しました: %s\n", path)
	return nil
}

func (c *Config) FindCommand(command string) (string, error) {
	if c.Settings.ToolDir != "" {
		return cmdutil.Find(command, c.Settings.ToolDir)
	}
	return cmdutil.Find(command)
}
