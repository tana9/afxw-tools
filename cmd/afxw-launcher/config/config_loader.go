package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tana9/afxw-tools/internal/cmdutil"
	"github.com/tana9/afxw-tools/internal/configutil"
)

func LoadFrom(path string) (*Config, error) {
	return configutil.LoadFrom[Config](path)
}

func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("ホームディレクトリの取得に失敗しました: %w", err)
	}
	return load(
		filepath.Join(home, ".config", "afxw-launcher", "config.toml"),
		filepath.Join(cmdutil.ExecDir(), "config.toml"),
	)
}

func load(userConfigPath, localConfigPath string) (*Config, error) {
	cfg, exists, err := loadExistingConfig(userConfigPath)
	if err != nil {
		return nil, err
	}
	if exists {
		return migrateUserConfig(userConfigPath, cfg)
	}

	cfg, exists, err = loadExistingConfig(localConfigPath)
	if err != nil {
		return nil, err
	}
	if exists {
		return cfg, nil
	}

	return createDefaultConfig(userConfigPath), nil
}

func loadExistingConfig(path string) (*Config, bool, error) {
	exists, err := configutil.Exists(path)
	if err != nil || !exists {
		return nil, exists, err
	}
	cfg, err := LoadFrom(path)
	return cfg, true, err
}

func migrateUserConfig(path string, cfg *Config) (*Config, error) {
	if !addOpenMenuItem(cfg) {
		return cfg, nil
	}
	if err := appendOpenMenuItem(path); err != nil {
		return nil, fmt.Errorf("設定ファイルの更新に失敗しました: %w", err)
	}
	fmt.Fprintf(os.Stderr, "設定ファイルに「ファイルを開く」メニューを追加しました: %s\n", path)
	return cfg, nil
}

func createDefaultConfig(path string) *Config {
	cfg := DefaultConfig()
	if err := configutil.Write(path, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "警告: 設定ファイルの作成に失敗しました: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "デフォルト設定ファイルを作成しました: %s\n", path)
	}
	return cfg
}

func addOpenMenuItem(cfg *Config) bool {
	for _, item := range cfg.Menu {
		if strings.EqualFold(filepath.Base(item.Command), "afxw-open.exe") {
			return false
		}
	}
	cfg.Menu = append(cfg.Menu, DefaultConfig().Menu[4])
	return true
}

func appendOpenMenuItem(path string) error {
	openMenu := DefaultConfig().Menu[4]
	return configutil.Append(path, struct {
		Menu []MenuItem `toml:"menu"`
	}{Menu: []MenuItem{openMenu}})
}
