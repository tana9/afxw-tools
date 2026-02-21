# afxw-tools

あふw（afxw）用のユーティリティツール集

## ツール一覧

### afxw-launcher
あふw用ツールランチャー - メニューから各ツールを選択して実行

**使い方:**
```bash
# メニューから選択して実行
afxw-launcher.exe
```

**設定ファイル:**
初回起動時に `~/.config/afxw-launcher/config.toml` が自動作成されます。
または実行ファイルと同じディレクトリに `config.toml` を配置することもできます。

```toml
[[menu]]
name = "フォルダ履歴から選択"
description = "あふwのフォルダ履歴から選択して移動"
command = "afxw-his.exe"
args = []

[[menu]]
name = "ブックマークから選択"
description = "ブックマークから選択して移動"
command = "afxw-bm.exe"
args = []

[[menu]]
name = "zoxideから選択"
description = "zoxideのfrecencyデータベースから選択して移動"
command = "afxw-zox.exe"
args = []

[[menu]]
name = "ブックマークを追加"
description = "現在のディレクトリをブックマークに追加"
command = "afxw-bm.exe"
args = ["-a"]

# カスタムツールの追加例
[[menu]]
name = "独自スクリプト"
description = "カスタムツールを実行"
command = "my-tool.exe"
args = []

[settings]
tool_dir = ""  # ツールの検索パス（省略時は実行ファイルと同じディレクトリ）
```

### afxw-his
あふwのフォルダ履歴から選択して移動するツール

**使い方:**
```bash
# 両方のウィンドウの履歴から選択
afxw-his.exe

# 左窓の履歴のみから選択
afxw-his.exe --window left

# 右窓の履歴のみから選択
afxw-his.exe --window right
```

### afxw-bm
ブックマーク管理ツール

**使い方:**
```bash
# ブックマークから選択して移動
afxw-bm.exe

# 現在のディレクトリをブックマークに追加
afxw-bm.exe --add

# 指定したパスをブックマークに追加
afxw-bm.exe --add C:\path\to\directory
```

### afxw-zox
zoxideのfrecency（頻度×最近性）データベースから選択してあふwで移動するツール

**前提条件:**
- [zoxide](https://github.com/ajeetdsouza/zoxide)がインストールされていること
- ターミナルでzoxideを使用してディレクトリデータベースが構築されていること

**使い方:**
```bash
# zoxideのデータベースから選択して移動
afxw-zox.exe
```

## 推奨設定

あふwから `afxw-launcher.exe` を1つのキーで呼び出すように設定すると便利です。

## ビルド

```bash
# すべてのツールをビルド
task build

# 個別にビルド
task build-his
task build-bm
task build-zox
task build-launcher
```

## テスト

```bash
task test
```