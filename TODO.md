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

## コードレビュー第2弾(2026-07-16)

パフォーマンス・共通化・一貫性観点の全体レビューで洗い出した対応項目。パフォーマンスについては、短命な対話型CLIでありデータ量も小さいため、既対応(zoxideキャッシュ等)以上の最適化は不要と判断。

- [x] **[suggestion] 「一覧→fuzzyfinder選択→EXCD移動」フローを共通化する**
  `cmd/afxw-his/main.go`(run)、`cmd/afxw-bm/main.go`(runSelect)、`cmd/afxw-zox/main.go`(run)で、「候補リスト取得→空なら終了→`f.Find`→`ErrAbort`ならnil→`a.EXCD(選択パス)`」の構造がほぼ同一で3箇所に重複している。`ErrAbort`をキャンセル扱いにするリポジトリ規約も分散している。
  対応: `internal/selectnav` を新設し `SelectAndMove(a afx.AFX, f finder.Finder, dirs []string) error` に共通化。空候補・キャンセル時はnil、EXCD失敗時は「ディレクトリ移動に失敗しました」でラップ。3コマンドとも候補の取得と空メッセージ表示のみ持つ形に書き換え、`selectnav_test.go` を追加。

- [x] **[warning] ブックマークの重複判定・正規化が Load と Add で不整合**
  `cmd/afxw-bm/bookmark/bookmark.go` で、`Add` は `strings.EqualFold` の大文字小文字を無視した比較なのに、`Load` の重複除去(`stringutil.RemoveDuplicates`)は大文字小文字を区別するため、手編集等で `C:\Foo` と `c:\foo` が両方あると選択リストに両方出る。また `Add` は `filepath.Clean` で正規化するが `Load` はしないため、末尾 `\` 付き等の既存行が重複チェックをすり抜ける。
  対応: `readAll`/`parseLines` ヘルパーに分離し、`parseLines` で `filepath.Clean` による正規化と小文字化キーでの重複除去を実施。`Load`/`Add` 双方がこれを共有する形に書き換え。

- [x] **[suggestion] `bookmark.Add` が末尾改行の無いファイルで行を結合してしまう**
  `bookmark.Add` は `newItem + "\n"` を単純追記するため、手編集で末尾改行が無い `bookmarks.txt` に追加すると最終行と新エントリが1行に結合される。
  対応: 追記前にファイル内容が `\n` で終わっているか確認し、無ければ改行を前置してから追記するように変更。

- [x] **[suggestion] afxw-zox の `singleinstance.Acquire` 呼び出し順を他ツールと揃える**
  `afxw-his`/`afxw-bm` は「`Acquire`→`NewOleAFX`」の順だが、`cmd/afxw-zox/main.go` はCOM接続後に `Acquire` しており、二重起動時に無駄にCOM接続してから最大3秒待つ。
  対応: `--import-history`(非対話)の分岐を先に処理し、対話パスでは `Acquire`→`NewOleAFX` の順に変更。

- [x] **[suggestion] launcher設定移行の `DefaultConfig().Menu[4]` マジックインデックスを排除する**
  `cmd/afxw-launcher/config/config_loader.go` の `addOpenMenuItem`/`appendOpenMenuItem` が「ファイルを開く」メニューを `Menu[4]` で参照しており、`DefaultConfig` のメニュー順を変えると移行処理が黙って壊れる。
  対応: `openMenuItem()` を新設し、`DefaultConfig` と移行処理(`addOpenMenuItem`/`appendOpenMenuItem`)の両方から参照するように変更。

- [x] **[suggestion] `cmdutil.Find` がディレクトリを実行ファイルとして返しうる**
  `internal/cmdutil/cmdutil.go` の絶対パス判定と `fileExists` は `os.Stat` の成功のみを見ておりディレクトリでも「見つかった」扱いになる。後段の `exec.Command` で失敗するため実害は小さいがエラーメッセージが分かりにくい。
  対応: `fileExists` に `!info.IsDir()` チェックを追加し、絶対パス分岐も `fileExists` 経由に統一。

- [x] **[suggestion] 細かな改善**
  - [x] `internal/cliutil/cliutil.go`: 「何かキーを押すと終了します」と表示するが `fmt.Scanln()` はEnterが必要。メッセージを「Enterキーを押すと終了します」に直すか、実際に1キー読み取りにする。
    対応: ブックマーク未登録時の案内が読めない問題の修正とあわせて、`cliutil.WaitForEnter()` として共通化しメッセージを「Enterキーを押すと終了します...」に修正。
  - [x] エラーコンテキストの二重ラップ: `afx.Histories` 内でラップ済みのエラーを `afxw-his/main.go` の run で再度ラップしており冗長。
    対応: his run・zox run/runImport・bm runSelect/addBookmark で、下位層が既にコンテキスト付きでラップしているエラーの再ラップを削除。テストの期待メッセージも更新。
  - [x] `configutil.Exists`→`LoadFrom` の2段階を、`LoadFrom` を直接呼んで `os.ErrNotExist` で分岐する形にすれば `Exists` を削減できる(TOCTOUも解消)。
    対応: `configutil.Exists` を削除し、launcher/openの設定ロードを `LoadFrom`+`errors.Is(err, os.ErrNotExist)` 判定に書き換え。`LoadFrom` の `%w` ラップが `os.ErrNotExist` を保持することを保証するテストを追加。

## コードレビュー第3弾(2026-07-16)

第2弾の対応後diffに対するマルチエージェントレビュー(8角度×検証付き)で生存した項目。検証で「大文字小文字を無視したブックマーク統合はNTFSケースセンシティブディレクトリを潰す」という候補は、CLAUDE.md明記の設計方針かつAdd既存動作との整合修正のためREFUTED(対応不要)と判定済み。

- [x] **[warning] afxw-hisだけ空履歴時に無言でウィンドウが閉じる**
  `cmd/afxw-his/main.go` のrunには空履歴時のメッセージ表示+`cliutil.WaitForEnter()` が無く、`selectnav.SelectAndMove` が空candidatesで黙ってnilを返すため、あふwから起動したコンソールが説明なく即閉じする。bm/zoxで直したのと同じUXバグが3ツール中hisにだけ残存。
  対応: `cliutil.Notice`(errorを実装する案内メッセージ型)を新設し、`cliutil.Run` がNoticeを検出したら表示+Enter待ちで正常終了するよう一元化。hisのrunは空履歴時に「フォルダ履歴が見つかりません。」のNoticeを返す。

- [x] **[warning] ブックマークの大文字小文字同一視がToLowerとEqualFoldの2機構に分裂し、Unicodeで実際に判定が食い違う**
  `cmd/afxw-bm/bookmark/bookmark.go` で、parseLinesは `strings.ToLower` キーのmap、Addは `strings.EqualFold` スキャンで同じ規則を別実装している。`strings.ToLower("İ")=="i"` はtrueだが `strings.EqualFold("İ","i")` はfalse(U+0130、実行確認済み)のため、手編集ファイルにU+0130を含むパスがあるとLoadでは統合されるのにAddの重複チェックは素通りする。
  対応: `normKey`(strings.ToLower)を新設し、parseLinesのdedupとAddの重複チェックの両方が同じキーで比較するように統一。U+0130の回帰テスト(`TestAdd_UnicodeCaseFold`)を追加。

- [x] **[suggestion] cliutil.WaitForEnterがユニットテスト対象関数内にあり、TTY接続stdinでテストがハングしうる**
  `cmd/afxw-bm/main.go` runSelect と `cmd/afxw-zox/main.go` run の空ケース分岐が実物の `WaitForEnter`(fmt.Scanln)を呼ぶため、`TestRunSelect_EmptyBookmarks`/`TestRun_EmptyEntries` はテストバイナリ直接実行やpty割当のIDEランナーでEnter入力待ちになりうる(CI/通常のgo testではstdinがnull接続のため即返る)。
  対応: 空ケースの案内をNotice返却に変更し、WaitForEnter呼び出しを `cliutil.Run`(テストから呼ばれないmain経路)へ集約。runSelect/runはブロッキング呼び出しを持たなくなり、テストはNotice型を検証する形に更新。

- [x] **[suggestion] afxw-zoxのNewOleAFX+エラーラップ+defer Closeブロックが2分岐に複製された**
  Acquire順序修正の際に、`cmd/afxw-zox/main.go` Action内のimport分岐(32-37行)と対話分岐(44-48行)に同一の接続ブロックが複製された。his/bm mainの同型ブロックも含めると4箇所以上。
  対応: Acquireを `if !importHistory` でゲートして接続を1回に集約。さらに `afx.Connect()` ヘルパーを新設し、his/bm/zox/launcher(args.go)の接続+日本語ラップを全て共通化。

- [x] **[suggestion] parseLinesの手実装dedupをstringutilのキー付き汎用版に統合する**
  `parseLines` のseen-mapパターンは `stringutil.RemoveDuplicates` と構造的に同一で、キー変換の有無だけが違う。
  対応: `stringutil.RemoveDuplicatesBy[T, K]` を追加し `RemoveDuplicates` をそのラッパー化。parseLinesは「分割・Clean」+`RemoveDuplicatesBy(lines, normKey)` の2パスに書き換え。テスト追加。

- [x] **[suggestion] loadExistingConfigのbool戻り値は cfg != nil と常に一致し冗長**
  `cmd/afxw-launcher/config/config_loader.go` の `(*Config, bool, error)` は全リターンパスでboolが `cfg != nil` から導出可能(検証済み)。同じdiffで簡素化したafxw-open側のload(2値)とスタイルも乖離。
  対応: `loadExistingConfig` 自体を削除し、`configutil.TryLoad[T]`(2値、未存在は nil, nil)への直接呼び出しに置き換え。呼び出し側は `cfg != nil` で分岐。

- [x] **[suggestion] LoadFrom+os.ErrNotExist分岐がlauncher/openの2箇所に重複**
  「LoadFrom→ErrNotExistなら未存在扱い→他エラーは伝播」の3分岐が `config_loader.go` と `afxw-open/config/config.go` に同型で存在する。launcher側はユーザー設定のみmigrateする非対称があるため「最初に見つかったものをロード」型の共通化は不適合だが、パス単位のヘルパーなら両方に適合する(検証済み)。
  対応: `configutil.TryLoad[T](path) (*T, error)` を追加し(boolなしの2値に統一)、launcher/openの両loadから使用。ErrNotExist判定はconfigutil内の1箇所のみに。テスト追加。

- [x] **[suggestion] diffで書き換えた3関数に日本語docコメントが無い(CLAUDE.md規約違反)**
  CLAUDE.md「Add a Japanese comment for every function, including unexported functions」に対し、`bookmark.Add`(bookmark.go:62)、`loadExistingConfig`(config_loader.go:49)、`load`(afxw-open/config/config.go:47)がコメント無し(3箇所とも確認済み)。特にAddは正規化・重複判定・末尾改行と挙動が変わったのに説明が無い。
  対応: `bookmark.Add`・launcher/openの `load` に日本語docコメントを追加(loadExistingConfigは削除により対象外)。Notice/Connect/TryLoad/RemoveDuplicatesBy等の新規関数にも規約どおりコメントを付与。

## コードレビュー第4弾(2026-07-23)

パフォーマンス・UI/UX観点の全体レビューで洗い出した項目。パフォーマンスは短命な対話型CLIでデータ量も小さく、既対応(zoxideキャッシュ等)以上の最適化は不要と再確認。1と2は未コミットの `wtMenuItem()` 追加をコミットする前に対応すべき。

- [x] **[warning] `wtMenuItem()` 追加でlauncher configのテストが失敗する**
  未コミット変更で `DefaultConfig()` のメニューが6件になったが、テストは5件を期待しており `go test ./...` が失敗していた。
  対応: 上流(origin/master)で同等機能が `standardMenuItems()` + 更新済みテストとして実装されたため、ローカル変更はstashに退避して破棄扱い(2026-07-23)。

- [x] **[warning] 既存ユーザー設定に「Windows Terminalを開く」が移行されない**
  対応: 上流の `addMissingStandardMenuItems`(config_loader.go)が標準メニューリスト(`standardMenuItems()`)に対する汎用移行として実装済み。wt/rg/diffも同機構で移行される(2026-07-23確認)。

- [x] **[warning] スペースを含むパスで `{files}` が黙って壊れる**
  `$MFP` のスペース区切りパースが原因で、スペースを含むパスが分断されていた。
  対応: 上流で `$JU$MF`(改行区切り)マクロによるパースに変更され解消(internal/afx/afx.go `getMarkedFiles`/`parseMarkedFiles`、2026-07-23確認)。

- [x] **[suggestion] 重複ブックマーク追加時も「追加しました」と表示される**
  対応: 第5弾(135行目)で対応済み。`bookmark.Add` が `(added bool, err error)` を返すよう変更し、`addBookmark` で出し分け(2026-07-23)。

- [x] **[suggestion] launcher TUIの細かなUI改善**
  - [x] `Description` が空の項目でも `descStyle` の行を無条件描画するため空行が入る問題。
    対応: `strings.TrimSpace(item.Description) != ""` の場合のみ描画するよう修正(`tui.go` View、2026-07-23)。
  - [x] 色が `170`/`241` 固定でライトテーマの端末ではhelp/descriptionが読みにくい問題。
    対応: `lipgloss.AdaptiveColor{Light: ..., Dark: ...}` に置き換え、ライト/ダーク両テーマに対応(2026-07-23)。
  - [ ] メニュー10件以上は数字キーで選択不可(現状6件で実害なし。増やす場合に注意。対応不要と判断し据え置き)。
  - [ ] 起動時の移行メッセージ(`config_loader.go` migrateUserConfig)は直後にTUI描画が始まり見逃されやすい(Notice機構への統合は表示フローの再設計が必要なため未着手)。

- [x] **[suggestion] あふw未起動時のエラーメッセージが不親切**
  対応: `afx.Connect` のエラーメッセージに「（あふwが起動しているか確認してください）」を追記(`internal/afx/afx.go`、2026-07-23)。

## コードレビュー第5弾(2026-07-23) — origin/master取り込み分(ed3f08a..12b9349)のレビュー

afxw-rg/afxw-diffの新規追加、標準メニュー移行の汎用化、`$JU$MF` による空白パス問題の解消、設定ファイルのアトミック書き込み、cliutilのTTY判定、e2eテスト追加など大きな改善を確認。`go build`/`go vet`/`go test ./...` は全通過。その上で以下を検出。

- [x] **[warning] `Validate()` が実行時に一度も呼ばれない(launcher/openとも)**
  `load()` が `configutil.TryLoad[Config]` を直接呼ぶため、`Validate` を実行するパッケージ側 `LoadFrom` ラッパーを経由せず、実際の起動時にはバリデーションが一切効いていなかった。
  対応: `tryLoadValidated(path)` ヘルパー(TryLoad成功時に `cfg.Validate()` を呼び、失敗時は `設定ファイル %q が不正です` でラップ)を launcher/open それぞれの `load()` から使うよう変更。既存の `LoadFrom`(未存在時はエラーにする版)はそのまま。`tool_dir` 不正も含めて起動時ハードエラーとする方針とした(FindCommandのフォールバックがあるとはいえ、ユーザー設定ミスは早期に気付けた方が良いと判断)。既存テストへの影響が無いことを確認済み(load()で意味的に不正だが構文的に正しいTOMLをテストしている箇所は無かった)。`go build`/`go test` で回帰無し(2026-07-23)。

- [x] **[warning] `reportError` の「何かキーを押すと終了します...」が第2弾で修正した文言バグの再発**
  対応: `internal/cliutil/cliutil.go` の `reportError` を「Enterキーを押すと終了します...」に修正し、`reportNotice` と表現を統一。`cliutil_test.go` の期待値も更新。`go test ./internal/cliutil/...` で確認済み(2026-07-23)。

- [x] **[suggestion] デッドコード2件: `configutil.Exists` の再追加と `internal/sliceutil` の重複**
  対応: `configutil.Exists`(呼び出し元なし)とそのテストを削除。`internal/sliceutil` パッケージ(`Unique`が`stringutil.RemoveDuplicates`と重複、自テスト以外に利用者なし)を丸ごと削除。あわせて `internal/configutil/atomic.go` のunexported関数(`atomicWrite`/`createTempFile`/`syncAndClose`/`cleanupTempFile`)に日本語docコメントを追加。`go build ./...`/`go test ./...` で全体の回帰無しを確認済み(2026-07-23)。

- [x] **[suggestion] `bookmark.Add` が重複時も「追加しました」と表示される**
  `Add` は重複時に何もせず `nil` を返すが、`addBookmark` は無条件に「追加しました」と表示していた。
  対応: `Add` のシグネチャを `(added bool, err error)` に変更し、重複時は `false, nil`、新規追加時は `true, nil` を返すように変更。`cmd/afxw-bm/main.go` の `addBookmark` は `added` で「ブックマークに追加しました」/「既に登録されています」を出し分け。呼び出し元テスト(`TestAdd`/`TestAddConcurrentDuplicate`/`TestAdd_UnicodeCaseFold`)も新シグネチャに追随。`go test ./cmd/afxw-bm/...` で確認済み(2026-07-23)。

- [x] **[verified] `bookmark.Add` の `addMu` ミューテックスは実は必要 — 削除提案はREFUTED**
  当初「`Add`はプロセスあたり1回しか呼ばれないため`addMu`は実効性が無い」と判断したが、`cmd/afxw-bm/bookmark/bookmark_test.go` の `TestAddConcurrentDuplicate`(20 goroutineから同一パスを並行`Add`し、最終的に1件のみ残ることを検証)が既に同一プロセス内の並行呼び出し安全性を仕様として要求していることを確認した。`addMu`を削除するとこのテストが指す競合(check-then-write間のレース)が再発するため、削除は誤り。プロセス間排他ではなくプロセス内goroutine安全性のためのミューテックスとして意図通り機能しており、対応不要と判定(2026-07-23)。

- [x] **[suggestion] afxw-rg autoモードは全ファイルを全量読みして文字コード判定するため大きいツリーで遅い**
  対応: `isUTF8File` を先頭 `classifySampleSize`(64KiB)までのサンプリング判定に変更。サンプル内で不正バイト列が見つからなければUTF-8として扱う。既存テストのファイルはいずれもサンプルサイズ未満のため挙動は変わらず、`go test ./cmd/afxw-rg/...` で確認済み(2026-07-23)。

- [x] **[suggestion] `Validate` のdocコメントが英語(CLAUDE.md規約違反)**
  対応: launcher/open両方の `Validate` docコメントを日本語に書き換え。`atomic.go` のコメント追加は上記デッドコード削除の対応に含めて実施(2026-07-23)。

## Go 1.26スタイルチェック

- `internal/afx/afx.go` の古典的カウントループを `for i := range count` に置き換え済み(上記参照)。
- それ以外は `strings.SplitSeq`(bookmark.go, zoxide.go)、`testing.B.Loop()`(zoxide_bench_test.go)、ジェネリクス(`stringutil.RemoveDuplicates`, `configutil.LoadFrom[T]`)など既にモダンな書き方が使われており、`interface{}`・`sort.Slice`・`ioutil`等の古い書き方は見つからなかった。追加対応なし。
