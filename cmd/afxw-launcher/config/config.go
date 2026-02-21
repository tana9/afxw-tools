package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// MenuItem はメニュー項目を表します。
type MenuItem struct {
	Name        string   `toml:"name"`
	Description string   `toml:"description"`
	Command     string   `toml:"command"`
	Args        []string `toml:"args"`
}

// Settings はツールの設定を表します。
type Settings struct {
	ToolDir string `toml:"tool_dir"`
}

// Config はアプリケーション設定を表します。
type Config struct {
	Menu     []MenuItem `toml:"menu"`
	Settings Settings   `toml:"settings"`
}

// DefaultConfig はデフォルト設定を返します。
func DefaultConfig() *Config {
	return &Config{
		Menu: []MenuItem{
			{
				Name:        "フォルダ履歴から選択",
				Description: "あふwのフォルダ履歴から選択して移動",
				Command:     "afxw-his.exe",
				Args:        []string{},
			},
			{
				Name:        "zoxideから選択",
				Description: "zoxideのfrecencyデータベースから選択して移動",
				Command:     "afxw-zox.exe",
				Args:        []string{},
			},
			{
				Name:        "ブックマークから選択",
				Description: "ブックマークから選択して移動",
				Command:     "afxw-bm.exe",
				Args:        []string{},
			},
			{
				Name:        "ブックマークを追加",
				Description: "現在のディレクトリをブックマークに追加",
				Command:     "afxw-bm.exe",
				Args:        []string{"-a", ""},
			},
		},
		Settings: Settings{
			ToolDir: "",
		},
	}
}

// LoadFrom は指定されたパスの設定ファイルを読み込みます。
func LoadFrom(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("設定ファイルの読み込みに失敗しました (%s): %w", path, err)
	}
	return &cfg, nil
}

// Load は設定ファイルを読み込みます。
// 設定ファイルが見つからない場合はデフォルト設定を作成して返します。
func Load() (*Config, error) {
	configPath := filepath.Join(os.Getenv("USERPROFILE"), ".config", "afxw-launcher", "config.toml")
	localPath := filepath.Join(getExecutableDir(), "config.toml")

	for _, path := range []string{configPath, localPath} {
		if _, err := os.Stat(path); err == nil {
			return LoadFrom(path)
		}
	}

	// 設定ファイルが見つからない場合はデフォルト設定を作成
	cfg := DefaultConfig()
	if err := createDefaultConfigFile(configPath, cfg); err != nil {
		// 作成に失敗してもデフォルト設定を返す（エラーにしない）
		fmt.Fprintf(os.Stderr, "警告: 設定ファイルの作成に失敗しました: %v\n", err)
	}

	return cfg, nil
}

// createDefaultConfigFile はデフォルト設定ファイルを作成します。
func createDefaultConfigFile(path string, cfg *Config) error {
	// ディレクトリを作成
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
	}

	// 設定ファイルを作成
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("ファイルの作成に失敗しました: %w", err)
	}
	defer f.Close()

	// TOMLフォーマットで書き込み
	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("設定の書き込みに失敗しました: %w", err)
	}

	fmt.Printf("デフォルト設定ファイルを作成しました: %s\n", path)
	return nil
}

// FindCommand はコマンドのフルパスを検索します。
func (c *Config) FindCommand(command string) (string, error) {
	// 絶対パスの場合はそのまま返す
	if filepath.IsAbs(command) {
		if _, err := os.Stat(command); err == nil {
			return command, nil
		}
		return "", fmt.Errorf("コマンドが見つかりません: %s", command)
	}

	// 検索パスのリスト
	searchPaths := []string{}

	// tool_dirが設定されている場合
	if c.Settings.ToolDir != "" {
		searchPaths = append(searchPaths, c.Settings.ToolDir)
	}

	// 実行ファイルと同じディレクトリ
	searchPaths = append(searchPaths, getExecutableDir())

	// 各ディレクトリで検索
	for _, dir := range searchPaths {
		fullPath := filepath.Join(dir, command)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}

	// PATH環境変数から検索
	fullPath, err := exec.LookPath(command)
	if err == nil {
		return fullPath, nil
	}

	return "", fmt.Errorf("コマンドが見つかりません: %s", command)
}

// getExecutableDir は実行ファイルのディレクトリを返します。
func getExecutableDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}
