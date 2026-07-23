# TODO

コードレビューで洗い出した対応項目。完了済みは「完了履歴」に要約して記録し、このファイルで実際に手を付けるべきなのは「残課題」のみ。

## 残課題

現時点でオープンな項目は無し。

## 完了履歴

### 第6弾(2026-07-23) — 残課題2件の対応
- [x] **launcher TUI: メニュー10件以上は数字キーで選択不可**
  対応: `tui.go` に数字入力バッファ(`numberInput`)を追加し、複数桁のメニュー番号に対応。もう1桁足すと必ずメニュー総数を超え候補が一意に定まる場合は即座に選択(9件以下では常にこの条件を満たすため、従来の1キー即時選択と完全互換)。曖昧な場合はEnterで確定。入力中の番号はヘルプ行下に表示。`cmd/afxw-launcher/tui_test.go` を新設し、単一桁即時確定・曖昧2桁の自動確定・Enterでの確定・範囲外/先頭ゼロの無視・カーソル移動でのバッファクリアを検証。

- [x] **launcher TUI: 起動時の設定移行メッセージが見逃されやすい**
  対応: `internal/cliutil.PressEnterToContinue`(対話端末なら表示後にEnter入力を待つ、非対話ならテスト等でハングしない)を新設し、`config_loader.go` の `migrateUserConfig` から利用するよう変更。TUI描画開始前に必ずメッセージが読まれることを保証。`cliutil_test.go` にテストを追加。

### 第1弾(初回レビュー) — パフォーマンス・保守性
- [x] CIに `go test`/lint ワークフローを追加(`.github/workflows/ci.yaml`、windows-latest)
- [x] TOML設定ロードの重複を `internal/configutil`(Exists/LoadFrom[T]/Write/Append)に共通化
- [x] zoxide実行ファイル探索を `sync.Once` でプロセス内キャッシュ
- [x] COM戻り値(VARIANT)への直接型アサーションのpanicを `toInt` ヘルパーで解消

### 第2弾(2026-07-16) — パフォーマンス・共通化・一貫性
- [x] 「一覧→fuzzyfinder選択→EXCD移動」フローを `internal/selectnav.SelectAndMove` に共通化
- [x] ブックマークのLoad/Addで重複判定・正規化(大文字小文字・末尾`\`)の不整合を解消
- [x] `bookmark.Add` が末尾改行なしファイルで既存行と結合してしまう不具合を修正
- [x] afxw-zoxの `singleinstance.Acquire` 呼び出し順をhis/bmと統一(二重起動時の無駄なCOM接続を回避)
- [x] launcher設定移行の `Menu[4]` マジックインデックスを `openMenuItem()` 参照に置換
- [x] `cmdutil.Find` がディレクトリを実行ファイルとして返しうる不具合を修正
- [x] 細かな改善: cliutil案内文言修正・エラー二重ラップ解消・`configutil.Exists`+`LoadFrom`の2段階をTryLoad相当に整理

### 第3弾(2026-07-16) — マルチエージェントレビュー(8角度×検証付き)
- [x] afxw-hisの空履歴時に無言でウィンドウが閉じる問題を `cliutil.Notice` で解消(bm/zoxと同じUXに統一)
- [x] ブックマーク大文字小文字同一視がToLower/EqualFoldの2機構に分裂しUnicode(U+0130)で食い違う問題を `normKey` に統一
- [x] `cliutil.WaitForEnter` がテスト対象関数内にありハングしうる問題をNotice分離で解消
- [x] afxw-zoxの接続(NewOleAFX+エラーラップ+defer Close)重複を `afx.Connect()` に共通化
- [x] `parseLines` の手実装dedupを `stringutil.RemoveDuplicatesBy` に統合
- [x] `loadExistingConfig` の冗長なbool戻り値を削除し `configutil.TryLoad[T]` に統一
- [x] 日本語docコメント漏れ(Add/load等)を補完
- **REFUTED**: 「大文字小文字を無視したブックマーク統合はNTFSケースセンシティブディレクトリを潰す」という指摘は、CLAUDE.md明記の設計方針かつAdd既存動作との整合修正のため対応不要と判定済み。

### 第4弾(2026-07-23) — パフォーマンス・UI/UX、origin/master取り込み前後
- [x] `wtMenuItem()` 追加によるテスト失敗・既存設定への移行漏れ・`{files}`の空白パス分断 → いずれもorigin/master取り込み(`standardMenuItems()`, `$JU$MF`)で解消済みと確認
- [x] 重複ブックマーク追加時の誤表示、launcher TUIの表示改善、あふw未起動時エラーメッセージ → 第5弾の対応と統合済み(下記参照)
- 上記のうち2件は「残課題」へ据え置き(数字キー選択の上限、移行メッセージの表示タイミング)

### 第5弾(2026-07-23) — origin/master取り込み分(ed3f08a..12b9349)のレビュー
- [x] `Validate()` が実行時に一度も呼ばれず設定検証が無効だった問題を `tryLoadValidated` ヘルパーで解消(launcher/open両方)
- [x] `reportError` の案内文言「何かキーを押すと」→「Enterキーを押すと」の再発(第2弾で一度修正済みだった)を再修正
- [x] デッドコード(`configutil.Exists`, `internal/sliceutil`)を削除、`atomic.go` に日本語docコメント追加
- [x] 重複ブックマーク追加時の誤表示を `bookmark.Add` の戻り値拡張(`added bool`)で解消
- [x] afxw-rgのUTF-8判定を全量読みから先頭64KiBサンプリングに変更(大きいツリーでrg本体より遅くなる問題)
- [x] `Validate` のdocコメントを日本語化(CLAUDE.md規約)
- [x] launcher TUIの空Description行抑制、AdaptiveColorによるライト/ダーク両対応
- [x] あふw未起動時のエラーメッセージにヒント(「あふwが起動しているか確認してください」)を追加
- **REFUTED**: `bookmark.Add` の `addMu` ミューテックス削除提案 — `TestAddConcurrentDuplicate` が同一プロセス内の並行呼び出し安全性を仕様として要求しており必要と判明、削除しない。

## Go 1.26スタイルチェック

- `internal/afx/afx.go` の古典的カウントループを `for i := range count` に置き換え済み。
- それ以外は `strings.SplitSeq`、`testing.B.Loop()`、ジェネリクス(`stringutil.RemoveDuplicates`, `configutil.LoadFrom[T]`)など既にモダンな書き方が使われており、`interface{}`・`sort.Slice`・`ioutil`等の古い書き方は見つからなかった。追加対応なし。
