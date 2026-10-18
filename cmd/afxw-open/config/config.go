package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tana9/afxw-tools/internal/cmdutil"
	"github.com/tana9/afxw-tools/internal/configutil"
)

// Program は選択肢として表示する外部プログラム（表示名・説明・コマンド・引数）を表します。
type Program struct {
	Name        string   `toml:"name"`
	Description string   `toml:"description"`
	Command     string   `toml:"command"`
	Args        []string `toml:"args"`
}

// Config は afxw-open の設定（プログラム一覧）を表します。
type Config struct {
	Programs []Program `toml:"program"`
}

// DefaultConfig は新規設定ファイルの生成に使う既定のプログラム一覧を返します。
func DefaultConfig() *Config {
	return &Config{
		Programs: []Program{
			{Name: "メモ帳", Description: "メモ帳で開く", Command: "notepad.exe", Args: []string{}},
			{Name: "エクスプローラー", Description: "エクスプローラーで開く", Command: "explorer.exe", Args: []string{}},
		},
	}
}

// Validate は、プログラムを実際に使ったときにしか判明しない不正な値がないかを検証します。
func (c *Config) Validate() error {
	for i, program := range c.Programs {
		if strings.TrimSpace(program.Name) == "" {
			return fmt.Errorf("program[%d].name が空です", i)
		}
		if strings.TrimSpace(program.Command) == "" {
			return fmt.Errorf("program[%d].command が空です", i)
		}
	}
	return nil
}

// LoadFrom は指定パスの設定ファイルを読み込んで検証します。
func LoadFrom(path string) (*Config, error) {
	return configutil.LoadValidated[Config](path)
}

// Load はユーザー設定とローカル設定を読み込み、afxw-open の設定を返します。
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
		cfg, err := configutil.TryLoadValidated[Config](path)
		if err != nil {
			return nil, err
		}
		if cfg != nil {
			return cfg, nil
		}
	}

	cfg := DefaultConfig()
	configutil.WriteDefault(configPath, cfg)
	return cfg, nil
}

// FindCommand は実行ファイルと同じフォルダおよびPATHからコマンドの実行ファイルを探します。
func FindCommand(command string) (string, error) {
	return cmdutil.Find(command)
}
