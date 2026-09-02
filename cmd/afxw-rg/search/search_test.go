package search

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRunAutoMixedEncodings(t *testing.T) {
	rg, err := exec.LookPath("rg")
	if err != nil {
		t.Skip("rg is not installed")
	}
	dir := t.TempDir()
	files := map[string][]byte{
		"utf8.txt":    []byte("日本語 TODO\n"),
		"utf8bom.txt": append([]byte{0xef, 0xbb, 0xbf}, []byte("日本語 TODO\n")...),
		"sjis.txt":    {0x82, 0xa0, 0x20, 'T', 'O', 'D', 'O', '\n'},
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}

	matches, err := Run(context.Background(), rg, Options{
		Root:         dir,
		Pattern:      "TODO",
		Encoding:     EncodingAuto,
		FixedStrings: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 3 {
		t.Fatalf("matches = %d, want 3: %+v", len(matches), matches)
	}
}

func TestClassifyFiles(t *testing.T) {
	dir := t.TempDir()
	files := map[string][]byte{
		"utf8.txt":    []byte("日本語"),
		"utf8bom.txt": append([]byte{0xef, 0xbb, 0xbf}, []byte("日本語")...),
		"sjis.txt":    {0x82, 0xa0, 0x82, 0xa2},
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}

	utf8Files, sjisFiles, err := classifyFiles(dir, []string{"utf8.txt", "utf8bom.txt", "sjis.txt"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(utf8Files, []string{"utf8.txt", "utf8bom.txt"}) {
		t.Errorf("UTF-8 files = %v", utf8Files)
	}
	if !reflect.DeepEqual(sjisFiles, []string{"sjis.txt"}) {
		t.Errorf("Shift_JIS files = %v", sjisFiles)
	}
}

func TestClassifyFilesError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "ok.txt"), []byte("ok"), 0644); err != nil {
		t.Fatal(err)
	}

	_, _, err := classifyFiles(dir, []string{"ok.txt", "missing.txt"})
	if err == nil {
		t.Fatal("エラーを期待しましたが nil でした")
	}
}

// classifyConcurrencyを超える件数のファイルすべてが判定エラーとなるケースでも、
// デッドロックせず正しくエラーを返すことを確認する（並行起動の打ち切りロジックの回帰防止）。
func TestClassifyFilesErrorManyFiles(t *testing.T) {
	dir := t.TempDir()
	files := make([]string, 0, classifyConcurrency*3)
	for i := range classifyConcurrency * 3 {
		files = append(files, fmt.Sprintf("missing-%d.txt", i))
	}

	_, _, err := classifyFiles(dir, files)
	if err == nil {
		t.Fatal("エラーを期待しましたが nil でした")
	}
}

func TestParseJSON(t *testing.T) {
	root := t.TempDir()
	output := []byte("" +
		`{"type":"begin","data":{"path":{"text":"main.go"}}}` + "\n" +
		`{"type":"match","data":{"path":{"text":"src/main.go"},"lines":{"text":"func main() {}\r\n"},"line_number":12}}` + "\n" +
		`{"type":"end","data":{"path":{"text":"main.go"}}}` + "\n")

	matches, err := parseJSON(root, output)
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 {
		t.Fatalf("matches = %d", len(matches))
	}
	if matches[0].Path != filepath.Join(root, "src", "main.go") || matches[0].Line != 12 || matches[0].Text != "func main() {}" {
		t.Errorf("unexpected match: %+v", matches[0])
	}
}

func TestBatches(t *testing.T) {
	got := batches([]string{"aaaa", "bbbb", "cc"}, 13)
	want := [][]string{{"aaaa"}, {"bbbb", "cc"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("batches() = %v, want %v", got, want)
	}
}

func TestValidateOptions(t *testing.T) {
	for _, encoding := range []string{EncodingAuto, EncodingUTF8, EncodingUTF8BOM, EncodingSJIS, EncodingShiftJIS} {
		if err := validateOptions(Options{Root: ".", Pattern: "test", Encoding: encoding}); err != nil {
			t.Errorf("encoding %s: %v", encoding, err)
		}
	}
	if err := validateOptions(Options{Root: ".", Pattern: "test", Encoding: "utf-16"}); err == nil {
		t.Error("unsupported encoding should fail")
	}
}

func TestNormalizeEncoding(t *testing.T) {
	tests := map[string]string{
		"UTF-8":     EncodingUTF8,
		"utf-8bom":  EncodingUTF8,
		"sjis":      EncodingShiftJIS,
		"SHIFT_JIS": EncodingShiftJIS,
	}
	for input, want := range tests {
		got, err := normalizeEncoding(input)
		if err != nil || got != want {
			t.Errorf("normalizeEncoding(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
}
