//go:build e2e

package e2e_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const testVersion = "e2e-test"

var commands = []string{
	"afxw-his",
	"afxw-bm",
	"afxw-zox",
	"afxw-launcher",
	"afxw-open",
	"afxw-wt",
	"afxw-rg",
	"afxw-diff",
}

func TestCLI(t *testing.T) {
	binDir := buildCommands(t)
	t.Run("help", func(t *testing.T) { testCommandHelp(t, binDir) })
	t.Run("version", func(t *testing.T) { testCommandVersion(t, binDir) })
	t.Run("bookmark add", func(t *testing.T) { testBookmarkAdd(t, binDir) })
	t.Run("open requires file path", func(t *testing.T) { testOpenRequiresFilePath(t, binDir) })
}

func testCommandHelp(t *testing.T, binDir string) {
	for _, name := range commands {
		t.Run(name, func(t *testing.T) {
			stdout, stderr, err := runCommand(binDir, name, "--help")
			if err != nil {
				t.Fatalf("--help failed: %v\nstderr:\n%s", err, stderr)
			}
			if !strings.Contains(stdout, "USAGE:") {
				t.Errorf("help output does not contain USAGE:\n%s", stdout)
			}
			if !strings.Contains(stdout, name) {
				t.Errorf("help output does not contain command name %q:\n%s", name, stdout)
			}
		})
	}
}

func testCommandVersion(t *testing.T, binDir string) {
	for _, name := range commands {
		t.Run(name, func(t *testing.T) {
			stdout, stderr, err := runCommand(binDir, name, "--version")
			if err != nil {
				t.Fatalf("--version failed: %v\nstderr:\n%s", err, stderr)
			}
			if !strings.Contains(stdout, testVersion) {
				t.Errorf("version output does not contain %q:\n%s", testVersion, stdout)
			}
		})
	}
}

func testBookmarkAdd(t *testing.T, binDir string) {
	target := filepath.Join(t.TempDir(), "target directory")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, err := runCommand(binDir, "afxw-bm", "--add", target)
	if err != nil {
		t.Fatalf("bookmark add failed: %v\nstderr:\n%s", err, stderr)
	}
	if !strings.Contains(stdout, "ブックマークに追加しました") {
		t.Errorf("unexpected stdout:\n%s", stdout)
	}

	data, err := os.ReadFile(filepath.Join(binDir, "bookmarks.txt"))
	if err != nil {
		t.Fatalf("read bookmarks: %v", err)
	}
	want, err := filepath.Abs(target)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(data)); got != want {
		t.Errorf("saved bookmark = %q, want %q", got, want)
	}
}

func testOpenRequiresFilePath(t *testing.T, binDir string) {
	_, stderr, err := runCommand(binDir, "afxw-open")
	if err == nil {
		t.Fatal("afxw-open without arguments succeeded, want failure")
	}
	if !strings.Contains(stderr, "ファイルパスを指定してください") {
		t.Errorf("unexpected stderr:\n%s", stderr)
	}
}

func buildCommands(t *testing.T) string {
	t.Helper()
	root := repositoryRoot(t)
	binDir := t.TempDir()

	for _, name := range commands {
		output := filepath.Join(binDir, name+exeSuffix())
		cmd := exec.Command("go", "build", "-ldflags=-X main.version="+testVersion, "-o", output, "./cmd/"+name)
		cmd.Dir = root
		if result, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build %s: %v\n%s", name, err, result)
		}
	}
	return binDir
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file path")
	}
	return filepath.Dir(filepath.Dir(filename))
}

func runCommand(binDir, name string, args ...string) (string, string, error) {
	path := filepath.Join(binDir, name+exeSuffix())
	cmd := exec.Command(path, args...)
	cmd.Dir = binDir
	cmd.Env = append(os.Environ(), "NO_COLOR=1")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func TestMain(m *testing.M) {
	if runtime.GOOS != "windows" {
		fmt.Fprintln(os.Stderr, "E2E tests require Windows")
		os.Exit(0)
	}
	os.Exit(m.Run())
}
