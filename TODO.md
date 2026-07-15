# TODO

コードレビュー(パフォーマンス・保守性観点)で洗い出した対応項目。

- [x] **[warning] CIにテスト/lintワークフローを追加する**
  現状 `.github/workflows/release.yaml` はタグpush時のGoReleaserのみで、push/PR時に `go test` や lint を自動実行するワークフローが無い。テスト自体は各パッケージに存在するのに、masterへの変更で壊れても自動検知できない。
  対応: `.github/workflows/ci.yaml` を追加し、`push`(master)/`pull_request` トリガーで `go test ./...` と `golangci-lint-action` を実行。Windows専用API(go-ole, x/sys/windows)に依存するため `windows-latest` ランナーで実行するようにした。

- [ ] **[suggestion] TOML設定ロード処理の重複を共通化する**
  `cmd/afxw-open/config/config.go`(33-85行目)と `cmd/afxw-launcher/config/config_loader.go`(14-103行目)で、「ユーザー設定パス→ローカル設定パス→無ければデフォルト生成」のフォールバック構造と `LoadFrom`(`toml.DecodeFile`)・`createDefaultConfigFile`(`os.MkdirAll`→`os.Create`→`toml.NewEncoder().Encode`)がほぼ同一実装で重複している。片方だけ直して直し忘れるリスクあり。
  対応: `internal/tomlconfig`(仮)のような共有ヘルパーに切り出し、各ツールは型とデフォルト値のみ渡す形にする。

- [ ] **[suggestion] zoxide実行ファイル探索のGlob結果をキャッシュ検討**
  `cmd/afxw-zox/zoxide/zoxide.go`(19-47行目)の `candidateExecutablePaths()` は呼び出しのたびにWinGetパッケージ配下へ `filepath.Glob` を2回実行する。現状は `ResolveExecutable()` がプロセスあたり1回しか呼ばれないため実害はないが、将来呼び出し回数が増える場合に備えて結果キャッシュの余地あり。優先度は低い。
