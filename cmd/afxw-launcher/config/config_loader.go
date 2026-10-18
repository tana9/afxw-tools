package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/tana9/afxw-tools/internal/cliutil"
	"github.com/tana9/afxw-tools/internal/cmdutil"
	"github.com/tana9/afxw-tools/internal/configutil"
)

// LoadFrom は指定パスの設定ファイルを読み込んで検証します。
func LoadFrom(path string) (*Config, error) {
	return configutil.LoadValidated[Config](path)
}

// Load はユーザー設定とローカル設定を読み込み、ランチャーの設定を返します。
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
	cfg, err := configutil.TryLoadValidated[Config](userConfigPath)
	if err != nil {
		return nil, err
	}
	if cfg != nil {
		return migrateUserConfig(userConfigPath, cfg)
	}

	cfg, err = configutil.TryLoadValidated[Config](localConfigPath)
	if err != nil {
		return nil, err
	}
	if cfg != nil {
		return cfg, nil
	}

	cfg = DefaultConfig()
	configutil.WriteDefault(userConfigPath, cfg)
	return cfg, nil
}

// migrateUserConfig はユーザー設定に不足している標準メニューを追記し、追加があれば案内を表示します。
func migrateUserConfig(path string, cfg *Config) (*Config, error) {
	items := addMissingStandardMenuItems(cfg)
	if len(items) == 0 {
		return cfg, nil
	}
	if err := appendMenuItems(path, items); err != nil {
		return nil, fmt.Errorf("設定ファイルの更新に失敗しました: %w", err)
	}
	// Enter入力を待ってから返すことで、直後に始まるTUI描画に案内メッセージが埋もれて
	// 見逃されるのを防ぐ。
	cliutil.PressEnterToContinue(fmt.Sprintf("設定ファイルに不足していた標準メニューを追加しました: %s", path))
	return cfg, nil
}

// addMissingStandardMenuItems は設定に存在しない標準メニューを追加し、追加した項目を返します。
func addMissingStandardMenuItems(cfg *Config) []MenuItem {
	standardItems := standardMenuItems()
	missing := make([]MenuItem, 0, len(standardItems))
	for _, standard := range standardItems {
		found := slices.ContainsFunc(cfg.Menu, func(item MenuItem) bool {
			return strings.EqualFold(filepath.Base(item.Command), standard.Command)
		})
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
