# リポジトリガイドライン

## プロジェクト構成 & モジュール構成

このリポジトリには、あふw(Afxw)ファイラーと連携するためのWindows向けユーティリティが含まれています。各実行ファイルは `cmd/` 以下に個別のパッケージを持ちます: `afxw-launcher`、`afxw-his`、`afxw-bm`、`afxw-zox`、`afxw-open`、`afxw-wt`、`afxw-rg`、`afxw-diff`。コマンド固有のコードとテストは、対応する `main.go` の近くに置いてください。再利用可能なパッケージは `internal/` に配置し、あふwのOLEアクセス、設定、コマンド、フィルタ検索、スライスヘルパーなどが含まれます。ビルド成果物は(gitignore対象の)`bin/` に出力され、リリース成果物は `dist/` を使用します。CIおよびリリースのワークフローは `.github/workflows/` にあります。

## アーキテクチャ

8つの実行ファイルはすべて、OLE/COM経由で起動中のあふwインスタンスと連携します。ランチャーはBubble Tea製のTUIメニューで、設定済みのコマンドを解決し、選択されたコマンドを子プロセスとして実行します。他の実行ファイルは直接呼び出すこともできます(例: あふwのマクロ/外部プログラムキーとして割り当てる)。

`internal/afx` は、起動中のあふwインスタンスとの境界です。`NewOleAFX` は `afxw.obj` COMオブジェクトを生成し、OSスレッドをロックしたうえで `afx.AFX` インターフェースを返します。このインターフェースを通じて、フォルダ履歴の読み取り、ディレクトリ変更(`EXCD`)、アクティブ/カーソル/マーク済みファイルの取得を行います。`NewOleAFX` の呼び出しに成功した直後には必ず `Close` を defer し、COMの終了処理とOSスレッドのアンロックを行ってください。COMの `VARIANT` 型の整数戻り値(例: `HisDirCount`)は、あふwが異なる整数サブタイプを返す場合があるため、直接の型アサーションではなくローカルの `toInt` ヘルパーを介して変換します。

各 `cmd/*` パッケージは、依存先を直接インスタンス化するのではなく、インターフェースを受け取るテスト可能な関数(例: `run(a afx.AFX, f finder.Finder, ...)`)としてコアロジックを構成します。本番コードでは `internal/afx.NewOleAFX` と `internal/finder.FuzzyFinder` を組み込み、ユニットテストでは `internal/afxtest` のフェイク(`MockAFX`、`MockFinder`)を使用します。コマンド固有の実装パッケージ(`bookmark`、`config`、`zoxide` など)は、そのコマンドの `cmd/` ディレクトリ配下に置きます。`internal/` は複数のコマンドで共有されるコード(`afx`、`afxtest`、`cliutil`、`cmdutil`、`configutil`、`finder`、`singleinstance`、`stringutil`)専用です。

設定の扱い:
- `afxw-launcher` と `afxw-open` は、まず `%USERPROFILE%\.config\<tool>\config.toml` からTOML設定を読み込み、見つからない場合は実行ファイルと同じ場所の設定ファイルにフォールバックし、どちらも存在しない場合はデフォルト値でユーザー設定を作成します。どちらも `toml` パッケージを直接呼び出すのではなく、共有の `internal/configutil` ヘルパー(`Exists`/`LoadFrom[T]`/`Write`/`Append`)を経由します。
- `afxw-launcher` のメニュー `args` は `{file}`/`{files}` プレースホルダーをサポートしており、実行時にあふwのアクティブ/マーク済みファイルから解決されます。
- `afxw-bm` は代わりに、ブックマークを自身の実行ファイルと同じ場所のプレーンテキスト(`bookmarks.txt`)として、1行1パスで、大文字小文字を区別せずに比較して保存します。
- `afxw-zox` には設定ファイルがなく、frecencyデータベースの問い合わせや履歴インポートのために、インストール済みの `zoxide` コマンド(`internal/cmdutil.Find` で解決。scoop/wingetの一般的なインストール場所も探索します)を呼び出します。

## ビルド・テスト・開発コマンド

リポジトリルートから [Task](https://taskfile.dev/) を使用します:

- `task build` は8つのWindows実行ファイルすべてを `bin/` にビルドします。
- `task build-launcher`(または `build-his`、`build-bm`、`build-zox`、`build-open`、`build-wt`、`build-rg`、`build-diff`)は1つのコマンドをビルドします。
- `task test` は `go test ./...` を実行します。
- `task integration-test` には `integration` ビルドタグで保護されたテストが含まれ、一部は対話的またはWindows固有の動作を伴います。
- `task e2e-test` は、あふwを必要とせずに8つの実行ファイルすべてをビルド・実行します。
- `task lint` は `golangci-lint run ./...` を実行します。

`go test ./internal/configutil` のように特定のパッケージパスを指定すれば、そのパッケージだけを対象にした短いテストサイクルを回せます。このモジュールは `go.mod` で宣言されているとおりGo 1.27を要求します。本プロジェクトはWindows専用(`go-ole`、`golang.org/x/sys/windows` を使用)であり、Windows上でのみ正しくビルド・実行できます。CI(`.github/workflows/ci.yaml`)は `master` へのpushとPRに対して `windows-latest` 上で `go test ./...` と `golangci-lint-action` を実行します。リリース成果物はGoReleaserにより `windows/amd64` 向けにのみ生成され、`v<major>.<minor>.<patch>` 形式のタグpushで `.github/workflows/release.yaml` がトリガーされます。

## コーディングスタイル & 命名規則

標準的なGoの規約に従ってください: 変更したファイルは `gofmt` でフォーマットし、フォーマッタが出力するタブをそのまま使い、パッケージ名は短く小文字にします。エクスポートする識別子は `PascalCase`、非エクスポートの識別子は `camelCase` を使用します。コマンド間で動作を重複させるより、`internal/` に小さな共有ヘルパーを置くことを優先してください。Windowsのパス操作やOLEとのやり取りは、テスト可能な関数やインターフェースの背後に隠してください。提出前に `task lint` を実行してください。

## リポジトリ固有の規約

- ユーザー向けの文字列・エラー・コメント・関数ドキュメントは日本語で記述します。非エクスポート関数を含め、すべての関数に日本語コメントを付けてください。
- 操作エラーは `%w` を使って日本語の文脈でラップし、呼び出し元やテストのために元のエラーを保持してください。
- 対話的な選択処理では `fuzzyfinder.ErrAbort`(Esc/Ctrl+C)を通常のキャンセルとして扱い、`nil` を返します。新しい選択フローでもこの挙動を維持してください。
- 対話的なUIフローの前には、ツールウィンドウの重複を防ぐために `singleinstance.Acquire` で名前付きミューテックスを取得してください。ブックマークの `-a`(追加)処理は意図的に非対話的であり、ミューテックスを取得しません。
- あふw固有のパス処理はWindows互換を保ってください: 必要な箇所ではバックスラッシュを使用し、`internal/afx.EXCD` の末尾バックスラッシュ正規化を保持し、ブックマークのパス比較は大文字小文字を区別しないようにします。
- このコードベースで既に使われているモダンなGo 1.22+のイディオム(`for i := range n`、`strings.SplitSeq`、`testing.B.Loop()`、ジェネリクス(`stringutil.RemoveDuplicates`、`configutil.LoadFrom[T]` など))を、従来型のカウントループ、`strings.Split`、`interface{}` よりも優先してください。

## テストガイドライン

テストはGoの `testing` パッケージを使用し、`*_test.go` として本番コードのそばに置きます。テスト関数は `TestName`、ベンチマークは `BenchmarkName` という命名規則に従います。修正や新機能に対してはユニットテストを追加し、有用な場合はテーブル駆動テストを使用してください。対話的な統合チェックには `//go:build integration` タグを付けたままにします。
実行ファイル単位のテストは `e2e/` にあり、`e2e` ビルドタグを使用し、ローカルにインストールされたあふwインスタンスを必要としてはいけません。

## コミット & プルリクエストガイドライン

最近のコミットは `feat:`、`fix:`、`add:`、`refactor:` のような簡潔なプレフィックスの後に、具体的な日本語の要約を続ける形式を使用しています。各コミットは焦点を絞ってください。プルリクエストでは、ユーザーに見える変更点を説明し、影響を受けるコマンドを明記し、関連するissueをリンクし、実施した検証内容を列挙してください。TUIの挙動が変わる場合はターミナル出力やスクリーンショットを含め、生成された `bin/`、`dist/`、ローカル設定ファイルはコミットしないでください。
