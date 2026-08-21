# afxw-tools Copilot Instructions

## Commands

- Install Go **1.27** as declared in `go.mod`; this is a Windows-focused project.
- Build every executable into `bin/`: `task build`
- Build one executable: `task build-his`, `task build-bm`, `task build-zox`, `task build-launcher`, or `task build-open`
- Run all unit tests: `task test` (equivalent to `go test ./...`)
- Run integration tests, which can open an interactive fuzzy finder: `task integration-test`
- Run one test or package directly with Go, for example:
  - `go test ./cmd/afxw-his -run '^TestRun$'`
  - `go test ./cmd/afxw-zox/zoxide -run '^TestParseQueryOutput$'`
- Lint with the repository's existing linter: `task lint` (requires `golangci-lint`)

## Architecture

This module builds five independent Windows executables under `cmd/`: `afxw-launcher`, `afxw-his`, `afxw-bm`, `afxw-zox`, and `afxw-open`. The launcher is a Bubble Tea menu that resolves configured commands and runs each selected executable as a child process; the other binaries can also be invoked directly.

`internal/afx` is the boundary to the running あふw instance. It creates `afxw.obj` through OLE/COM and exposes the `afx.AFX` interface for histories, directory changes, and file selections. `NewOleAFX` locks the OS thread and `Close` must be deferred after a successful creation so COM is released and the thread is unlocked.

Commands that need the あふw state accept interfaces in their core functions (for example, `run(a afx.AFX, f finder.Finder, ...)`) instead of directly instantiating dependencies. Production code passes `internal/finder.FuzzyFinder`; unit tests use `internal/afxtest.MockAFX` and `MockFinder`. Put command-specific implementation packages below that command's `cmd/` directory (such as `bookmark`, `config`, and `zoxide`); reserve `internal/` for code shared by multiple commands.

`afxw-launcher` and `afxw-open` load TOML configuration from `%USERPROFILE%\.config\...` before an executable-local configuration file, and create a user configuration with defaults when neither exists. Launcher arguments support `{file}` and `{files}` placeholders, which obtain values from あふw. `afxw-bm` instead stores `bookmarks.txt` beside its executable. `afxw-zox` delegates database queries and imports to the installed `zoxide` command.

## Repository Conventions

- User-facing strings, errors, comments, and function documentation are Japanese. Add a Japanese comment for every function, including unexported functions.
- Wrap operational errors with Japanese context using `%w`, preserving the underlying error for callers and tests.
- Interactive selection treats `fuzzyfinder.ErrAbort` (Esc/Ctrl+C) as normal cancellation, returning `nil`; retain this behavior in new selection flows.
- Acquire the named mutex with `singleinstance.Acquire` before interactive UI flows to prevent duplicate tool windows. The bookmark `-a` path is intentionally non-interactive and does not acquire it.
- Keep afxw-specific path behavior Windows-compatible: use backslashes where required, preserve `internal/afx.EXCD` trailing-backslash normalization, and compare bookmark paths case-insensitively.
- Release artifacts are produced by GoReleaser only for `windows/amd64`; tag pushes matching `v<major>.<minor>.<patch>` trigger `.github/workflows/release.yaml`.
