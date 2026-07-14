package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/tana9/afxw-tools/internal/cmdutil"
)

func LoadFrom(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("設定ファイルの読み込みに失敗しました (%s): %w", path, err)
	}
	return &cfg, nil
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
	if _, err := os.Stat(path); err == nil {
		cfg, err := LoadFrom(path)
		return cfg, true, err
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, false, fmt.Errorf("設定ファイルの確認に失敗しました (%s): %w", path, err)
	}
	return nil, false, nil
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
	if err := createDefaultConfigFile(path, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "警告: 設定ファイルの作成に失敗しました: %v\n", err)
	}
	return cfg
}

func createDefaultConfigFile(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
	}
	if err := writeConfig(path, cfg); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "デフォルト設定ファイルを作成しました: %s\n", path)
	return nil
}

func writeConfig(path string, cfg *Config) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("ファイルの作成に失敗しました: %w", err)
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		return fmt.Errorf("設定の書き込みに失敗しました: %w", err)
	}
	return nil
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
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("ファイルのオープンに失敗しました: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString("\n"); err != nil {
		return fmt.Errorf("改行の書き込みに失敗しました: %w", err)
	}

	openMenu := DefaultConfig().Menu[4]
	if err := toml.NewEncoder(f).Encode(struct {
		Menu []MenuItem `toml:"menu"`
	}{Menu: []MenuItem{openMenu}}); err != nil {
		return fmt.Errorf("設定の書き込みに失敗しました: %w", err)
	}
	return nil
}
