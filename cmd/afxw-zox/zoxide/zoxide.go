package zoxide

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type Entry struct {
	Path  string
	Score float64
}

// candidateExecutablePaths はPATHに無い場合にzoxideを探す候補パスを優先度順に返します。
func candidateExecutablePaths() []string {
	var candidates []string
	if scoop := os.Getenv("SCOOP"); scoop != "" {
		candidates = append(candidates, filepath.Join(scoop, "shims", "zoxide.exe"))
	} else if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, "scoop", "shims", "zoxide.exe"))
	}
	if scoopGlobal := os.Getenv("SCOOP_GLOBAL"); scoopGlobal != "" {
		candidates = append(candidates, filepath.Join(scoopGlobal, "shims", "zoxide.exe"))
	} else if programData := os.Getenv("ProgramData"); programData != "" {
		candidates = append(candidates, filepath.Join(programData, "scoop", "shims", "zoxide.exe"))
	}
	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		// wingetはportable/zip形式のパッケージ用にこのフォルダへ実体のシムを作成する。
		candidates = append(candidates, filepath.Join(localAppData, "Microsoft", "WinGet", "Links", "zoxide.exe"))

		// Linksが無い環境向けに、Packages配下のインストール先を探索する。
		// フォルダ名にバージョンやハッシュを含むためglobで柔軟に探す。
		packagesGlob := filepath.Join(localAppData, "Microsoft", "WinGet", "Packages", "ajeetdsouza.zoxide_*", "zoxide.exe")
		if matches, err := filepath.Glob(packagesGlob); err == nil {
			candidates = append(candidates, matches...)
		}
		nestedGlob := filepath.Join(localAppData, "Microsoft", "WinGet", "Packages", "ajeetdsouza.zoxide_*", "*", "zoxide.exe")
		if matches, err := filepath.Glob(nestedGlob); err == nil {
			candidates = append(candidates, matches...)
		}
	}
	return candidates
}

var (
	resolveOnce      sync.Once
	resolvedExecPath string
)

// ResolveExecutable はPATH上のzoxideを探し、見つからない場合はscoopやwingetの既定インストール先を確認します。
// どれでも見つからない場合は "zoxide" をそのまま返し、実行時にexec.Commandがエラーを報告します。
// 探索結果はプロセス内でキャッシュされ、2回目以降の呼び出しはファイルシステムを再走査しません。
func ResolveExecutable() string {
	resolveOnce.Do(func() {
		resolvedExecPath = resolveExecutable()
	})
	return resolvedExecPath
}

func resolveExecutable() string {
	if path, err := exec.LookPath("zoxide"); err == nil {
		return path
	}
	for _, candidate := range candidateExecutablePaths() {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}
	return "zoxide"
}

// Query はzoxideのクエリを実行してスコア降順のエントリを返します。
func Query() ([]Entry, error) {
	cmd := exec.Command(ResolveExecutable(), "query", "--list", "--score")
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("zoxideコマンドの実行に失敗しました: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("zoxideコマンドの実行に失敗しました: %w", err)
	}

	return parseQueryOutput(string(output))
}

// parseQueryOutput はzoxide query --list --scoreの出力（"スコア パス"形式）をパースします。
func parseQueryOutput(output string) ([]Entry, error) {
	var raw []Entry
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}
		score, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			continue
		}
		raw = append(raw, Entry{Path: parts[1], Score: score})
	}

	// zoxide は --all が指定されない限り利用不能なディレクトリを出力しない。
	return raw, nil
}

func Paths(entries []Entry) []string {
	paths := make([]string, len(entries))
	for i, entry := range entries {
		paths[i] = entry.Path
	}
	return paths
}
