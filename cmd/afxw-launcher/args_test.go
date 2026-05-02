package main

import (
	"reflect"
	"testing"
)

func TestExpandArgs_NoPlaceholders(t *testing.T) {
	args := []string{"--flag", "value"}
	got := expandArgs(args, `C:\file.txt`, []string{`C:\a.txt`, `C:\b.txt`})
	if !reflect.DeepEqual(got, args) {
		t.Errorf("got %v, want %v", got, args)
	}
}

func TestExpandArgs_FileReplaced(t *testing.T) {
	args := []string{"--input", "{file}"}
	got := expandArgs(args, `C:\foo.txt`, nil)
	want := []string{"--input", `C:\foo.txt`}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestExpandArgs_FileEmbedded(t *testing.T) {
	args := []string{"prefix_{file}_suffix"}
	got := expandArgs(args, `C:\foo.txt`, nil)
	want := []string{`prefix_C:\foo.txt_suffix`}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestExpandArgs_FilesStandalone_ExpandsToMultiple(t *testing.T) {
	marked := []string{`C:\a.txt`, `C:\b.txt`, `C:\c.txt`}
	got := expandArgs([]string{"{files}"}, "", marked)
	if !reflect.DeepEqual(got, marked) {
		t.Errorf("got %v, want %v", got, marked)
	}
}

func TestExpandArgs_FilesEmbedded_ExpandsPerFile(t *testing.T) {
	marked := []string{`C:\a.txt`, `C:\b.txt`}
	got := expandArgs([]string{"--file={files}"}, "", marked)
	want := []string{`--file=C:\a.txt`, `--file=C:\b.txt`}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestExpandArgs_MixedArgs(t *testing.T) {
	// {file} と {files} を含まない引数はそのまま保持される
	marked := []string{`C:\a.txt`, `C:\b.txt`}
	args := []string{"--output", "result.txt", "{files}"}
	got := expandArgs(args, "", marked)
	want := []string{"--output", "result.txt", `C:\a.txt`, `C:\b.txt`}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestExpandArgs_FilesEmpty(t *testing.T) {
	// マーク済みファイルが空のとき {files} は0要素に展開される
	got := expandArgs([]string{"{files}"}, "", []string{})
	if len(got) != 0 {
		t.Errorf("expected empty result, got %v", got)
	}
}
