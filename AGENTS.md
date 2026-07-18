# Repository Guidelines

## Project Structure & Module Organization

This repository contains Windows utilities for integrating with the Afxw file manager. Each executable has its own package under `cmd/`: `afxw-launcher`, `afxw-his`, `afxw-bm`, `afxw-zox`, `afxw-open`, and `afxw-wt`. Keep command-specific code and tests beside the relevant `main.go`. Reusable packages belong in `internal/`, including Afxw OLE access, configuration, command, finder, and slice helpers. Build output goes to the ignored `bin/` directory; release artifacts use `dist/`. CI and release workflows are in `.github/workflows/`.

## Build, Test, and Development Commands

Use [Task](https://taskfile.dev/) from the repository root:

- `task build` builds all six Windows executables into `bin/`.
- `task build-launcher` (or `build-his`, `build-bm`, `build-zox`, `build-open`, `build-wt`) builds one command.
- `task test` runs `go test ./...`.
- `task integration-test` includes tests guarded by the `integration` build tag; some exercise interactive or Windows-specific behavior.
- `task e2e-test` builds and exercises all six executables on Windows without requiring Afxw.
- `task lint` runs `golangci-lint run ./...`.

Run `go test ./internal/configutil` or another package path for a focused test cycle. The module requires Go 1.26 with the toolchain version declared in `go.mod`.

## Coding Style & Naming Conventions

Follow standard Go conventions: format changed files with `gofmt`, use tabs as emitted by the formatter, and keep package names short and lowercase. Exported identifiers use `PascalCase`; unexported identifiers use `camelCase`. Prefer small shared helpers in `internal/` over duplicating behavior across commands. Keep Windows paths and OLE interactions testable behind focused functions or interfaces. Run `task lint` before submitting.

## Testing Guidelines

Tests use Go's `testing` package and live beside production code as `*_test.go`; test functions follow `TestName`, with `BenchmarkName` for benchmarks. Add unit tests for fixes and new behavior, using table-driven cases where useful. Keep interactive integration checks tagged with `//go:build integration`. CI requires the standard test suite and golangci-lint to pass.
Executable-level tests live in `e2e/`, use the `e2e` build tag, and must not require a locally installed Afxw instance.

## Commit & Pull Request Guidelines

Recent commits use concise prefixes such as `feat:`, `fix:`, `add:`, and `refactor:`, followed by a specific Japanese summary. Keep each commit focused. Pull requests should explain the user-visible change, identify affected commands, link relevant issues, and list verification performed. Include terminal output or screenshots when TUI behavior changes; do not commit generated `bin/`, `dist/`, or local configuration files.
