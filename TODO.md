# TODO

コードレビュー(パフォーマンス・保守性観点)で洗い出した対応項目。

- [x] **[warning] CIにテスト/lintワークフローを追加する**
  現状 `.github/workflows/release.yaml` はタグpush時のGoReleaserのみで、push/PR時に `go test` や lint を自動実行するワークフローが無い。テスト自体は各パッケージに存在するのに、masterへの変更で壊れても自動検知できない。
  対応: `.github/workflows/ci.yaml` を追加し、`push`(master)/`pull_request` トリガーで `go test ./...` と `golangci-lint-action` を実行。Windows専用API(go-ole, x/sys/windows)に依存するため `windows-latest` ランナーで実行するようにした。

- [x] **[suggestion] TOML設定ロード処理の重複を共通化する**
  `cmd/afxw-open/config/config.go`(33-85行目)と `cmd/afxw-launcher/config/config_loader.go`(14-103行目)で、「ユーザー設定パス→ローカル設定パス→無ければデフォルト生成」のフォールバック構造と `LoadFrom`(`toml.DecodeFile`)・`createDefaultConfigFile`(`os.MkdirAll`→`os.Create`→`toml.NewEncoder().Encode`)がほぼ同一実装で重複している。片方だけ直して直し忘れるリスクあり。
  対応: `internal/configutil` を新設し、`Exists`/`LoadFrom[T]`/`Write`/`Append` の汎用ヘルパーに切り出した。`afxw-open/config` と `afxw-launcher/config` の両方をこのヘルパー経由に書き換え、`toml` パッケージへの直接依存とボイラープレートを削除。テスト(`internal/configutil/configutil_test.go`)も追加し、`go build ./...` / `go test ./...` / `go vet ./...` で回帰が無いことを確認済み。

- [x] **[suggestion] zoxide実行ファイル探索のGlob結果をキャッシュ検討**
  `cmd/afxw-zox/zoxide/zoxide.go` の `candidateExecutablePaths()` は呼び出しのたびにWinGetパッケージ配下へ `filepath.Glob` を2回実行する。現状 `ResolveExecutable()` はプロセスあたり1回しか呼ばれないため実害はなかった(ユーザー判断により、将来の呼び出し増加に備えて実装)。
  対応: `sync.Once` で `ResolveExecutable()` の結果をプロセス内キャッシュするようにした。既存テストは `candidateExecutablePaths()` を直接呼んでおり `ResolveExecutable()` 経由ではないため、テスト分離への影響なし。`go build ./...` / `go vet ./...` / `go test ./...` で回帰が無いことを確認済み。

- [x] **[warning] `internal/afx/afx.go:87` — COM戻り値への型アサーションがpanicしうる**
  `HisDirCount` の戻り値(VARIANT)を `res.Value().(int32)` と直接アサーションしており、`afxw.obj` 側の実装差異で別の整数サブタイプ(int16/int64等)が返るとpanicする。同ファイルの他の呼び出し(`extract`)は `fmt.Sprint` で型を問わず安全に処理しているのに対し、ここだけ一貫性が無かった。
  対応: `toInt(v any) (int, error)` を追加し、int/int16/int32/int64を許容してpanicせずエラーを返すように変更。あわせて `for i := 0; i < int(count); i++` を Go 1.22以降の `for i := range count` に置き換え。`toInt` の単体テストを `internal/afx/afx_test.go` に追加。

## 未対応のレビュー指摘

- [x] **[warning] 空白を含むマーク済みファイルのパスを正しく解析する**
  `internal/afx/afx.go` の `parseMarkedFiles` は `$MFP` の展開結果を `strings.Fields` で分割するため、`C:\My Files\a.txt` のようなパスが複数の引数に壊れる。`afxw-open` の `{files}` から通常利用される経路なので優先度は高い。
  対応: 公式仕様の区切り制御マクロを使い、`$JU$QN$MF` でLF区切り・引用符なしのパス一覧を取得して行単位で解析するよう変更。空白を含む複数パスとCRLFのテストを追加した。

- [x] **[warning] ブックマーク追加を並行実行に対して安全にする**
  `afxw-bm -a` は単一起動制御を取得せず、`bookmark.Add` が「読み込み→重複確認→追記」を別操作で行う。同時起動すると重複登録や書き込み競合が起こりうる。
  対応: CLIの追加処理にも名前付きmutexを適用し、`bookmark.Add` 内にもプロセス内mutexを追加。同じパスを20 goroutineから同時追加して1行だけ保存されるraceテストを追加した。

- [x] **[suggestion] `EXCD` のCOM戻り値を解放する**
  `internal/afx/afx.go` の `EXCD` は `oleutil.CallMethod` が返す `VARIANT` を破棄しており、他のCOM呼び出しと異なり `Clear()`していない。短命プロセスのため影響は限定的だが、リソース管理を統一する。
  対応: 戻り値を受け取って `Clear()`するよう変更し、解放とエラーラップをテストした。

- [x] **[suggestion] COM境界のテストを拡充する**
  `internal/afx` のステートメントカバレッジは現状22.7%で、実COM呼び出し周辺の回帰を検出しにくい。COM呼び出しを差し替え可能な境界に分離し、正常系・エラー系・リソース解放を単体テストできる構造を検討する。
  対応: `oleutil.CallMethod` を差し替え可能な関数境界にまとめ、`EXCD` のメソッド名・引数・エラー伝播・VARIANT解放を実COMなしで検証できるテストを追加した。

## Go 1.26スタイルチェック

- `internal/afx/afx.go` の古典的カウントループを `for i := range count` に置き換え済み(上記参照)。
- それ以外は `strings.SplitSeq`(bookmark.go, zoxide.go)、`testing.B.Loop()`(zoxide_bench_test.go)、ジェネリクス(`sliceutil.Unique`, `configutil.LoadFrom[T]`)など既にモダンな書き方が使われており、`interface{}`・`sort.Slice`・`ioutil`等の古い書き方は見つからなかった。追加対応なし。

## 追加の改善対応

- [x] **設定ファイル書き込みを原子化する**
  `configutil.Write` / `Append` を、同一ディレクトリの一時ファイルへ書き込み、`Sync`・`Close`後に置換する方式へ変更。Windowsでは `MoveFileEx`、その他のOSでは `os.Rename` を使用し、失敗時に既存ファイルを維持するテストを追加した。

- [x] **設定値を読み込み時に検証する**
  launcher/openの設定に `Validate` を追加し、空の名前・コマンドと、不正な `tool_dir` を設定ファイルの読み込み時点で報告するようにした。

- [x] **同期実行する外部コマンドへcontextを伝播する**
  launcherとzoxideのquery/importを `exec.CommandContext` 化。非同期起動が仕様の `afxw-open` は、Action終了時のキャンセルで子プロセスを停止させないため対象外とした。

- [x] **CLIエラー時のキー入力待ちをTTYに限定する**
  stdinが端末の場合だけ終了待ちを行い、リダイレクト・スクリプト・CIでは即時終了するよう変更。対話・非対話の両ケースをテストした。

## 構造・命名の改善

- [x] **汎用スライス処理の名前を実装に合わせる**
  `internal/stringutil.RemoveDuplicates` を `internal/sliceutil.Unique` へ変更し、文字列以外にも使えるジェネリック関数であることを名前に反映した。

- [x] **afxw-bmのCLI Actionを分割する**
  コマンド分岐、追加対象の解決、ブックマーク選択を `runBookmarkCommand`、`resolveBookmarkTarget`、`selectBookmark` へ抽出した。

- [x] **設定の原子書き込み処理を分割する**
  一時ファイルの生成、同期・クローズ、後始末を小さな関数へ分け、原子書き込み固有の処理を `internal/configutil/atomic.go` に移した。

- [x] **afx APIを利用目的に沿った名前へ変更する**
  `afx.AFX` / `oleAFX` / `NewOleAFX` を `afx.Client` / `oleClient` / `NewOLEClient` に変更し、操作名を `DirectoryHistories`、`ChangeDirectory`、`ActivePath`、`CurrentFile`、`MarkedFiles` に統一。テスト用モックと変数名も同じ語彙へ揃えた。

## テスト構成の整理

- [x] **重複するテストケースをテーブル駆動へ統合する**
  launcherの引数展開、ブックマークの読み書き、履歴・zoxide選択、zoxide出力解析を、振る舞い単位の親テストと名前付きサブテストへ整理した。共通のセットアップと検証を一箇所にまとめ、各ケースでは入力・依存モック・期待結果だけを宣言する構造に変更した。

- [x] **未検証だった境界とコマンド分岐のテストを追加する**
  COM履歴・Extract・VARIANT解放、afxw-bmの追加／選択／フォールバック、launcherのAFX引数解決、contextキャンセル、TUIの上下端・Enter・終了キーを追加。実COMや外部zoxideを必要としないよう依存生成を小さな境界で差し替え可能にした。

## 依存関係の更新

- [x] **GoモジュールとGitHub Actionsを更新する**
  `go get -u ./...` と `go mod tidy` で直接・実行時の間接依存を更新。`urfave/cli/v3`、`x/sys`、`x/term` とTUI関連の間接依存を更新し、CIでは `actions/checkout@v6`、`actions/setup-go@v6`、`golangci-lint-action@v9` を使用するよう変更した。
