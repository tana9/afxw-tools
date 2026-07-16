package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tana9/afxw-tools/internal/cmdutil"
	"github.com/tana9/afxw-tools/internal/configutil"
)

type Program struct {
	Name        string   `toml:"name"`
	Description string   `toml:"description"`
	Command     string   `toml:"command"`
	Args        []string `toml:"args"`
}

type Config struct {
	Programs []Program `toml:"program"`
}

func DefaultConfig() *Config {
	return &Config{
		Programs: []Program{
			{Name: "メモ帳", Description: "メモ帳で開く", Command: "notepad.exe", Args: []string{}},
			{Name: "エクスプローラー", Description: "エクスプローラーで開く", Command: "explorer.exe", Args: []string{}},
		},
	}
}

func LoadFrom(path string) (*Config, error) {
	return configutil.LoadFrom[Config](path)
}

func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("ホームディレクトリの取得に失敗しました: %w", err)
	}
	configPath := filepath.Join(home, ".config", "afxw-open", "config.toml")
	localPath := filepath.Join(cmdutil.ExecDir(), "afxw-open.toml")
	return load(configPath, localPath)
}

// load はユーザー設定→ローカル設定の順に読み込み、どちらも無ければデフォルト設定を生成します。
func load(configPath, localPath string) (*Config, error) {
	for _, path := range []string{configPath, localPath} {
		cfg, err := configutil.TryLoad[Config](path)
		if err != nil {
			return nil, err
		}
		if cfg != nil {
			return cfg, nil
		}
	}

	cfg := DefaultConfig()
	if err := configutil.Write(configPath, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "警告: 設定ファイルの作成に失敗しました: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "デフォルト設定ファイルを作成しました: %s\n", configPath)
	}
	return cfg, nil
}

func FindCommand(command string) (string, error) {
	return cmdutil.Find(command)
}
