# afxw-tools

あふw（afxw）用のユーティリティツール集です。

## ツール一覧

| ツール | 概要 |
|--------|------|
| [afxw-launcher](#afxw-launcher) | ツールランチャー（起点）。メニューから各ツールを呼び出す |
| [afxw-his](#afxw-his) | あふwのフォルダ履歴から選択して移動 |
| [afxw-bm](#afxw-bm) | ブックマークを管理し、選択したフォルダに移動 |
| [afxw-zox](#afxw-zox) | zoxideのデータベースから選択して移動 |
| [afxw-open](#afxw-open) | カーソル上のファイルをプログラム選択して開く |
| [afxw-wt](#afxw-wt) | アクティブなペインのフォルダをWindows Terminalで開く |
| [afxw-rg](#afxw-rg) | 現在のフォルダ以下をripgrepでキーワード検索 |
| [afxw-diff](#afxw-diff) | マークした2項目をWinMergeで比較 |

すべてのツールはあふwが起動中の状態で使用します（OLE経由であふwと連携します）。

## セットアップ

### 1. 実行ファイルを配置する

[Releases](../../releases) からダウンロードするか、[自分でビルド](#ビルド)して `.exe` を任意のフォルダに配置します。

### 2. あふwからランチャーを呼び出せるようにする

あふwの設定でキーにランチャーを割り当てます。あふwのマクロ設定（`AFX32.INI` や設定ダイアログ）から外部プログラムとして登録してください。

```
afxw-launcher.exe
```

1つのキーでランチャーが開き、そこから各ツールを選べます。

### 3. 設定ファイルを確認する

各ツールは初回起動時にデフォルト設定ファイルを自動生成します。必要に応じて編集してください。

| ツール | 設定ファイル |
|--------|-------------|
| afxw-launcher | `~/.config/afxw-launcher/config.toml` |
| afxw-open | `~/.config/afxw-open/config.toml` |
| afxw-bm | 実行ファイルと同じフォルダの `bookmarks.txt` |

---

## ツール

### afxw-launcher

メニューから各ツールを選択して実行するランチャーです。あふwとの連携の起点として使います。

**キー操作:**

| キー | 動作 |
|------|------|
| ↑ / k | 上に移動 |
| ↓ / j | 下に移動 |
| 1〜9 | 番号で直接選択・実行 |
| Enter | 選択して実行 |
| q / Esc / Ctrl+C | 終了 |

**設定ファイル:** `~/.config/afxw-launcher/config.toml`（実行ファイルと同フォルダの `config.toml` も可）

ユーザー設定ファイルには、更新後の初回起動時に不足している標準メニューが自動追加されます。

```toml
[[menu]]
name = "フォルダ履歴から選択"
description = "あふwのフォルダ履歴から選択して移動"
command = "afxw-his.exe"
args = []

[[menu]]
name = "zoxideから選択"
description = "zoxideのfrecencyデータベースから選択して移動"
command = "afxw-zox.exe"
args = []

[[menu]]
name = "ブックマークから選択"
description = "ブックマークから選択して移動"
command = "afxw-bm.exe"
args = []

[[menu]]
name = "ブックマークを追加"
description = "現在のディレクトリをブックマークに追加"
command = "afxw-bm.exe"
args = ["-a", ""]

[[menu]]
name = "ファイルを開く"
description = "選択ファイルをプログラム選択して開く"
command = "afxw-open.exe"
args = ["{files}"]

[[menu]]
name = "Windows Terminalで開く"
description = "現在のフォルダをWindows Terminalで開く"
command = "afxw-wt.exe"
args = []

[[menu]]
name = "キーワード検索"
description = "現在のフォルダ以下をripgrepで検索"
command = "afxw-rg.exe"
args = []

[[menu]]
name = "WinMergeで比較"
description = "マークした2項目をWinMergeで比較"
command = "afxw-diff.exe"
args = []

[settings]
tool_dir = ""  # ツールの検索パス（省略時は実行ファイルと同じディレクトリ）
```

**引数のプレースホルダー:**

`args` にはあふwの状態を参照するプレースホルダーを使えます。

| プレースホルダー | 展開内容 |
|-----------------|---------|
| `{file}` | アクティブウィンドウのカーソル位置のファイルのフルパス |
| `{files}` | マーク済みファイルのフルパス一覧（マークなしの場合はカーソルファイル）。1引数が複数引数に展開される |

---

### afxw-his

あふwのフォルダ履歴をfuzzyfinderで絞り込み、選択したフォルダに移動します。

```bash
# 両ウィンドウの履歴から選択して移動
afxw-his.exe

# 左窓の履歴のみ
afxw-his.exe --window left

# 右窓の履歴のみ
afxw-his.exe --window right
```

---

### afxw-bm

フォルダのブックマークを管理します。ブックマークをfuzzyfinderで絞り込み、選択したフォルダに移動します。

```bash
# ブックマークから選択して移動
afxw-bm.exe

# あふwのアクティブパスをブックマークに追加
afxw-bm.exe -a

# 指定したパスをブックマークに追加
afxw-bm.exe -a "C:\path\to\directory"
```

ブックマークは実行ファイルと同じフォルダの `bookmarks.txt` に1行1フォルダで保存されます。

---

### afxw-zox

[zoxide](https://github.com/ajeetdsouza/zoxide) のfrecency（頻度×最近性）データベースから選択してあふwで移動します。ターミナルでよく使うフォルダにすばやくジャンプできます。

**前提:** zoxideがインストールされ、データベースが構築されていること。PATHにzoxideが無い場合でも、scoopやwingetでインストールされていれば既定のインストール先（scoop: `%SCOOP%\shims` または `%USERPROFILE%\scoop\shims`、グローバルインストールなら `%SCOOP_GLOBAL%\shims` または `%ProgramData%\scoop\shims` / winget: `%LOCALAPPDATA%\Microsoft\WinGet\Links` またはPackages配下）を自動的に探して実行します。

```bash
# zoxideのデータベースから選択して移動
afxw-zox.exe

# あふwの履歴をzoxideデータベースにインポート（初回推奨）
afxw-zox.exe --import-history
```

---

### afxw-open

あふwで選択したファイルをfuzzyfinderでプログラムを選んで開きます。プログラムは非同期で起動し、ツール自体はすぐに終了します。

```bash
# カーソル上のファイルを開く（あふwのマクロから: $F = カーソルファイルのフルパス）
afxw-open.exe "$F"

# マーク済みファイルをまとめて開く（$MF = マーク済みファイルのフルパス一覧）
afxw-open.exe $MF
```

afxw-launcherから呼び出す場合は `args = ["{files}"]` を設定します（[上記参照](#afxw-launcher)）。

**設定ファイル:** `~/.config/afxw-open/config.toml`（実行ファイルと同フォルダの `afxw-open.toml` も可）

```toml
[[program]]
name = "VSCode"
description = "Visual Studio Codeで開く"
command = "code.exe"
args = []

[[program]]
name = "サクラエディタ"
description = "サクラエディタで開く"
command = "sakura.exe"
args = []

[[program]]
name = "7-Zip (解凍)"
description = "カレントディレクトリに解凍"
command = "7z.exe"
args = ["x", "-y"]
```

---

### afxw-rg

あふwのアクティブなペインのフォルダ以下を [ripgrep](https://github.com/BurntSushi/ripgrep) で検索し、結果をfuzzy finderから選んで該当ファイルにカーソルを移動します。

**前提:** `rg.exe` が実行ファイルと同じフォルダまたはPATHにあること。

```bash
# キーワードを対話入力（ランチャーからの標準動作）
afxw-rg.exe

# キーワードを直接指定
afxw-rg.exe TODO

# 固定文字列、隠しファイル、globを指定
afxw-rg.exe -F --hidden -g "*.go" "error handling"

# 文字コードを限定
afxw-rg.exe --encoding sjis "検索文字列"
```

`--encoding auto`（既定）では、検索対象をUTF-8・UTF-8 BOMとShift_JISへ分類して検索結果を統合します。`utf-8`、`utf-8bom`、`sjis`（`shift_jis`も可）を指定すると文字コードを限定できます。

---

### afxw-diff

あふwのアクティブなペインでマークした2つのファイルまたはフォルダを [WinMerge](https://winmerge.org/) で比較します。空白や日本語を含むパスにも対応します。

**前提:** WinMergeがインストールされていること。`WinMergeU.exe` をPATH、ユーザーの標準インストール先、`Program Files`、`Program Files (x86)` の順に探します。

```bash
# あふwで2項目をマークして比較
afxw-diff.exe

# パスを直接指定して比較
afxw-diff.exe "C:\path\before" "C:\path\after"
```

選択数が2つでない場合はWinMergeを起動せず、現在の選択数を案内します。ランチャーの新規設定と既存ユーザー設定には標準メニューとして追加されます。

---

### afxw-wt

あふwのアクティブなペイン（ウィンドウ）のフォルダを開始ディレクトリとして [Windows Terminal](https://github.com/microsoft/terminal) を開きます。Windows Terminalは非同期で起動し、ツール自体はすぐに終了します。

**前提:** Windows Terminalがインストールされていること。PATHに `wt.exe` が無い場合でも、既定のエイリアス（`%LOCALAPPDATA%\Microsoft\WindowsApps\wt.exe`）を自動的に探して実行します。

```bash
# アクティブなペインのフォルダでWindows Terminalを開く
afxw-wt.exe
```

既にWindows Terminalが開いている場合は、直近で使用したウィンドウに新しいタブを追加します。開いていない場合は新しいウィンドウを作成します。

afxw-launcherの新規設定には標準メニューとして含まれます。既存のユーザー設定にも起動時に自動追加されます。手動で設定する場合は以下を追加します。

```toml
[[menu]]
name = "Windows Terminalを開く"
description = "アクティブなペインのフォルダをWindows Terminalで開く"
command = "afxw-wt.exe"
args = []
```

---

## ビルド

```bash
# すべてのツールをビルド（bin/ に出力）
task build

# 個別にビルド
task build-launcher
task build-his
task build-bm
task build-zox
task build-open
task build-wt
task build-rg
task build-diff
```

## テスト

```bash
task test
```
