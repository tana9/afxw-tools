package bookmark

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		createFile bool
		want       []string
	}{
		{name: "missing file", want: []string{}},
		{name: "empty file", createFile: true, want: []string{}},
		{
			name:       "content",
			createFile: true,
			content:    "C:\\Users\\Test\\Dir1\nC:\\Users\\Test\\Dir2\nC:\\Users\\Test\\Dir3\n",
			want:       []string{`C:\Users\Test\Dir1`, `C:\Users\Test\Dir2`, `C:\Users\Test\Dir3`},
		},
		{
			name:       "duplicates",
			createFile: true,
			content:    "C:\\Users\\Test\\Dir1\nC:\\Users\\Test\\Dir2\nC:\\Users\\Test\\Dir1\nC:\\Users\\Test\\Dir3\nC:\\Users\\Test\\Dir2\n",
			want:       []string{`C:\Users\Test\Dir1`, `C:\Users\Test\Dir2`, `C:\Users\Test\Dir3`},
		},
		{
			name:       "blank lines",
			createFile: true,
			content:    "C:\\Users\\Test\\Dir1\n\nC:\\Users\\Test\\Dir2\n\nC:\\Users\\Test\\Dir3\n",
			want:       []string{`C:\Users\Test\Dir1`, `C:\Users\Test\Dir2`, `C:\Users\Test\Dir3`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "bookmarks.txt")
			if tt.createFile {
				if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
					t.Fatal(err)
				}
			}
			got, err := Load(path)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Load() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		name  string
		items []string
		want  []string
	}{
		{
			name:  "new item",
			items: []string{`C:\Users\Test\NewDir`},
			want:  []string{`C:\Users\Test\NewDir`},
		},
		{
			name:  "duplicate item",
			items: []string{`C:\Users\Test\Dir1`, `C:\Users\Test\Dir1`},
			want:  []string{`C:\Users\Test\Dir1`},
		},
		{
			name:  "case insensitive duplicate",
			items: []string{`C:\Users\Test\Dir1`, `c:\users\test\dir1`},
			want:  []string{`C:\Users\Test\Dir1`},
		},
		{
			name:  "multiple items",
			items: []string{`C:\Users\Test\Dir1`, `C:\Users\Test\Dir2`, `C:\Users\Test\Dir3`},
			want:  []string{`C:\Users\Test\Dir1`, `C:\Users\Test\Dir2`, `C:\Users\Test\Dir3`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "bookmarks.txt")
			for _, item := range tt.items {
				if _, err := Add(path, item); err != nil {
					t.Fatalf("Add(%q) error = %v", item, err)
				}
			}
			got, err := Load(path)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Load() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddConcurrentDuplicate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bookmarks.txt")
	item := `C:\Users\Test\Concurrent`

	const workers = 20
	var wg sync.WaitGroup
	errCh := make(chan error, workers)
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := Add(path, item)
			errCh <- err
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("Add() error = %v", err)
		}
	}

	dirs, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(dirs) != 1 || dirs[0] != filepath.Clean(item) {
		t.Fatalf("got %v, want one %q", dirs, filepath.Clean(item))
	}
}

func TestMutexNameForPath(t *testing.T) {
	base := filepath.Join(t.TempDir(), "bookmarks.txt")

	name1 := mutexNameForPath(base)
	name2 := mutexNameForPath(base)
	if name1 != name2 {
		t.Errorf("同一パスに対する名前が一致しません: %q != %q", name1, name2)
	}

	// Windowsパスは大文字小文字を区別しないため、大文字化しても同じ名前になるべき
	if got := mutexNameForPath(strings.ToUpper(base)); got != name1 {
		t.Errorf("大文字小文字違いのパスで名前が一致しません: %q != %q", got, name1)
	}

	other := filepath.Join(t.TempDir(), "bookmarks.txt")
	if got := mutexNameForPath(other); got == name1 {
		t.Errorf("異なるパスなのに同じ名前になりました: %q", got)
	}
}

func TestAdd_UnicodeCaseFold(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bookmarks.txt")

	// U+0130(İ)はstrings.ToLowerでは"i"になるがstrings.EqualFoldでは"i"と一致しない。
	// LoadのdedupとAddの重複チェックが同じ正規化キー(normKey)を使うことを保証する。
	if _, err := Add(path, `C:\Users\Test\Dİr1`); err != nil {
		t.Fatal(err)
	}
	if _, err := Add(path, `C:\Users\Test\Dir1`); err != nil {
		t.Fatal(err)
	}

	dirs, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dirs) != 1 {
		t.Errorf("expected 1 item, got %d: %v", len(dirs), dirs)
	}
}
