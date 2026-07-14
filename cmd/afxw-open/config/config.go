package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/tana9/afxw-tools/internal/cmdutil"
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
	configPath := filepath.Join(home, ".config", "afxw-open", "config.toml")
	localPath := filepath.Join(cmdutil.ExecDir(), "afxw-open.toml")
	return load(configPath, localPath)
}

func load(configPath, localPath string) (*Config, error) {
	if _, err := os.Stat(configPath); err == nil {
		return LoadFrom(configPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("設定ファイルの確認に失敗しました (%s): %w", configPath, err)
	}

	if _, err := os.Stat(localPath); err == nil {
		return LoadFrom(localPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("設定ファイルの確認に失敗しました (%s): %w", localPath, err)
	}

	cfg := DefaultConfig()
	if err := createDefaultConfigFile(configPath, cfg); err != nil {
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

func FindCommand(command string) (string, error) {
	return cmdutil.Find(command)
}
