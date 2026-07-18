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
	cfg, err := configutil.LoadFrom[Config](path)
	if err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("設定ファイル %q が不正です: %w", path, err)
	}
	return cfg, nil
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

// load はユーザー設定→ローカル設定の順に読み込み、どちらも無ければデフォルト設定を生成します。
// ユーザー設定が存在する場合のみ、不足メニューの移行(migrateUserConfig)を行います。
func load(userConfigPath, localConfigPath string) (*Config, error) {
	cfg, err := configutil.TryLoad[Config](userConfigPath)
	if err != nil {
		return nil, err
	}
	if cfg != nil {
		return migrateUserConfig(userConfigPath, cfg)
	}

	cfg, err = configutil.TryLoad[Config](localConfigPath)
	if err != nil {
		return nil, err
	}
	if cfg != nil {
		return cfg, nil
	}

	return createDefaultConfig(userConfigPath), nil
}

func migrateUserConfig(path string, cfg *Config) (*Config, error) {
	items := addMissingStandardMenuItems(cfg)
	if len(items) == 0 {
		return cfg, nil
	}
	if err := appendMenuItems(path, items); err != nil {
		return nil, fmt.Errorf("設定ファイルの更新に失敗しました: %w", err)
	}
	_, _ = fmt.Fprintf(os.Stderr, "設定ファイルに不足していた標準メニューを追加しました: %s\n", path)
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

// addMissingStandardMenuItems は設定に存在しない標準メニューを追加し、追加した項目を返します。
func addMissingStandardMenuItems(cfg *Config) []MenuItem {
	standardItems := standardMenuItems()
	missing := make([]MenuItem, 0, len(standardItems))
	for _, standard := range standardItems {
		found := false
		for _, item := range cfg.Menu {
			if strings.EqualFold(filepath.Base(item.Command), standard.Command) {
				found = true
				break
			}
		}
		if !found {
			cfg.Menu = append(cfg.Menu, standard)
			missing = append(missing, standard)
		}
	}
	return missing
}

// appendMenuItems は既存設定ファイルを維持したままメニュー項目を追記します。
func appendMenuItems(path string, items []MenuItem) error {
	return configutil.Append(path, struct {
		Menu []MenuItem `toml:"menu"`
	}{Menu: items})
}
