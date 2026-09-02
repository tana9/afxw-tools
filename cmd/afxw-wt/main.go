package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tana9/afxw-tools/internal/afx"
	"github.com/tana9/afxw-tools/internal/cliutil"
	"github.com/tana9/afxw-tools/internal/cmdutil"
	"github.com/urfave/cli/v3"
)

var version = "dev"

func main() {
	cmd := &cli.Command{
		Name:    "afxw-wt",
		Usage:   "あふwのアクティブなペインのフォルダをWindows Terminalで開く",
		Version: version,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			a, err := afx.Connect()
			if err != nil {
				return err
			}
			defer a.Close()

			path, err := a.GetActivePath()
			if err != nil {
				return fmt.Errorf("アクティブパスの取得に失敗しました: %w", err)
			}

			return run(path, resolveWt(), cmdutil.StartCommand)
		},
	}

	cliutil.Run(cmd)
}

// run は直近のWindows Terminalウィンドウに、指定フォルダを開始ディレクトリとするタブを非同期で開きます。
// 既存のウィンドウがない場合は新しいウィンドウが作成されます。
func run(dir string, wtPath string, start func(string, []string) error) error {
	if dir == "" {
		return fmt.Errorf("アクティブパスが取得できませんでした")
	}

	if err := start(wtPath, []string{"-w", "0", "-d", dir}); err != nil {
		return fmt.Errorf("wt.exeの起動に失敗しました: %w", err)
	}

	return nil
}

// resolveWt はwt.exeのパスを解決します。
// PATHに無い場合はWindowsAppsの既定エイリアスを確認し、どちらにも無ければ "wt.exe" をそのまま返して実行時にエラーを報告させます。
func resolveWt() string {
	var alias string
	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		alias = filepath.Join(localAppData, "Microsoft", "WindowsApps", "wt.exe")
	}
	return cmdutil.ResolveExecutable("wt.exe", alias)
}
