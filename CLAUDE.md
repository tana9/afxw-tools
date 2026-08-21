# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

- Go **1.27** is required (declared in `go.mod`); this is a Windows-focused project (uses `go-ole`, `golang.org/x/sys/windows`) and only builds/runs correctly on Windows.
- Build every executable into `bin/`: `task build`
- Build one executable: `task build-his`, `task build-bm`, `task build-zox`, `task build-launcher`, `task build-open`, `task build-wt`, `task build-rg`, or `task build-diff`
- Run all unit tests: `task test` (equivalent to `go test ./...`)
- Run integration tests, which can open an interactive fuzzy finder: `task integration-test` (equivalent to `go test -tags=integration ./...`)
- Run one test or package directly with Go, for example:
  - `go test ./cmd/afxw-his -run '^TestRun$'`
  - `go test ./cmd/afxw-zox/zoxide -run '^TestParseQueryOutput$'`
- Lint with the repository's existing linter: `task lint` (requires `golangci-lint`)
- CI (`.github/workflows/ci.yaml`) runs `go test ./...` and `golangci-lint-action` on `windows-latest` for pushes to `master` and PRs, since the code depends on Windows-only APIs.
- Release artifacts are produced by GoReleaser only for `windows/amd64`; tag pushes matching `v<major>.<minor>.<patch>` trigger `.github/workflows/release.yaml`.

## Architecture

This module builds eight independent Windows executables under `cmd/`: `afxw-launcher`, `afxw-his`, `afxw-bm`, `afxw-zox`, `afxw-open`, `afxw-wt`, `afxw-rg`, and `afxw-diff`. All of them interoperate with a running instance of あふw (afxw), a Windows file manager, via OLE/COM. The launcher is a Bubble Tea TUI menu that resolves configured commands and runs each selected executable as a child process; the other binaries can also be invoked directly (e.g. wired up as あふw macro/external-program keys).

`internal/afx` is the boundary to the running あふw instance. `NewOleAFX` creates the `afxw.obj` COM object, locks the OS thread, and returns the `afx.AFX` interface for reading folder histories, changing directory (`EXCD`), and reading the active/cursor/marked files. `Close` must always be deferred right after a successful `NewOleAFX` call so COM is uninitialized and the OS thread is unlocked. COM `VARIANT` integer return values (e.g. `HisDirCount`) are converted through the local `toInt` helper rather than a direct type assertion, since あふw may return different integer subtypes.

Each `cmd/*` package structures its core logic as testable functions that accept interfaces (for example `run(a afx.AFX, f finder.Finder, ...)`) instead of instantiating dependencies directly. Production code wires in `internal/afx.NewOleAFX` and `internal/finder.FuzzyFinder`; unit tests use the fakes in `internal/afxtest` (`MockAFX`, `MockFinder`). Command-specific implementation packages (such as `bookmark`, `config`, `zoxide`) live below that command's `cmd/` directory; `internal/` is reserved for code shared across multiple commands (`afx`, `afxtest`, `cliutil`, `cmdutil`, `configutil`, `finder`, `singleinstance`, `stringutil`).

Configuration handling:
- `afxw-launcher` and `afxw-open` load TOML config from `%USERPROFILE%\.config\<tool>\config.toml` first, falling back to an executable-local config file, and create a user config with defaults if neither exists. Both go through the shared `internal/configutil` helpers (`Exists`/`LoadFrom[T]`/`Write`/`Append`) rather than calling the `toml` package directly.
- `afxw-launcher` menu `args` support `{file}`/`{files}` placeholders that are resolved from あふw's active/marked files at execution time.
- `afxw-bm` instead stores bookmarks as plain text (`bookmarks.txt`) beside its own executable, one path per line, compared case-insensitively.
- `afxw-zox` has no config file; it shells out to the installed `zoxide` command (resolved via `internal/cmdutil.Find`, which also probes common scoop/winget install locations) for frecency database queries and history import.

## Repository Conventions

- User-facing strings, errors, comments, and function documentation are Japanese. Add a Japanese comment for every function, including unexported functions.
- Wrap operational errors with Japanese context using `%w`, preserving the underlying error for callers and tests.
- Interactive selection treats `fuzzyfinder.ErrAbort` (Esc/Ctrl+C) as normal cancellation, returning `nil`; retain this behavior in new selection flows.
- Acquire the named mutex with `singleinstance.Acquire` before interactive UI flows to prevent duplicate tool windows. The bookmark `-a` (add) path is intentionally non-interactive and does not acquire it.
- Keep afxw-specific path behavior Windows-compatible: use backslashes where required, preserve `internal/afx.EXCD`'s trailing-backslash normalization, and compare bookmark paths case-insensitively.
- Prefer modern Go 1.22+ idioms already used in this codebase: `for i := range n`, `strings.SplitSeq`, `testing.B.Loop()`, and generics (e.g. `stringutil.RemoveDuplicates`, `configutil.LoadFrom[T]`) over classic counting loops, `strings.Split`, or `interface{}`.
