# afxw-tools

あふw（afxw）用のユーティリティツール集

## ツール一覧

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

## ビルド

```bash
# すべてのツールをビルド
task build

# 個別にビルド
task build-his
task build-bm
task build-zox
```

## テスト

```bash
task test
```