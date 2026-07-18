package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/tana9/afxw-tools/cmd/afxw-rg/search"
	"github.com/tana9/afxw-tools/internal/afxtest"
	"github.com/tana9/afxw-tools/internal/cliutil"
)

func TestPromptKeyword(t *testing.T) {
	var output strings.Builder
	got, err := promptKeyword(strings.NewReader("日本語 keyword\n"), &output)
	if err != nil {
		t.Fatal(err)
	}
	if got != "日本語 keyword" || output.String() != "検索キーワード: " {
		t.Errorf("promptKeyword() = %q, output = %q", got, output.String())
	}
}

func TestRunPromptedRetriesAfterNoMatches(t *testing.T) {
	a := &afxtest.MockAFX{}
	f := &afxtest.MockFinder{Idx: 0}
	var output strings.Builder
	var patterns []string

	err := runPrompted(
		context.Background(),
		a,
		f,
		search.Options{Root: `C:\work`},
		strings.NewReader("first\nsecond\n"),
		&output,
		func(_ context.Context, opts search.Options) ([]search.Match, error) {
			patterns = append(patterns, opts.Pattern)
			if opts.Pattern == "first" {
				return nil, nil
			}
			return []search.Match{{Path: `C:\work\found.txt`}}, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(patterns, ","), "first,second"; got != want {
		t.Fatalf("patterns = %q, want %q", got, want)
	}
	if !strings.Contains(output.String(), "検索結果が見つかりませんでした。\n検索キーワード: ") {
		t.Fatalf("output = %q", output.String())
	}
	if a.ExcdFilePath != `C:\work\found.txt` {
		t.Fatalf("EXCD file path = %q", a.ExcdFilePath)
	}
}

func TestRunPromptedEmptyKeywordCancels(t *testing.T) {
	called := false
	err := runPrompted(
		context.Background(),
		&afxtest.MockAFX{},
		&afxtest.MockFinder{},
		search.Options{},
		strings.NewReader("\n"),
		&strings.Builder{},
		func(context.Context, search.Options) ([]search.Match, error) {
			called = true
			return nil, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("empty keyword should not run a search")
	}
}

func TestRunSelectsMatchDirectory(t *testing.T) {
	a := &afxtest.MockAFX{}
	f := &afxtest.MockFinder{Idx: 1}
	opts := search.Options{Root: `C:\work`, Pattern: "TODO", Encoding: search.EncodingAuto}
	matches := []search.Match{
		{Path: `C:\work\a.go`, Line: 1, Text: "TODO a"},
		{Path: `C:\work\sub\b.go`, Line: 2, Text: "TODO b"},
	}
	err := run(context.Background(), a, f, opts, func(context.Context, search.Options) ([]search.Match, error) {
		return matches, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if a.ExcdFilePath != `C:\work\sub\b.go` {
		t.Errorf("EXCD file path = %q", a.ExcdFilePath)
	}
}

func TestRunNoMatches(t *testing.T) {
	err := run(context.Background(), &afxtest.MockAFX{}, &afxtest.MockFinder{}, search.Options{}, func(context.Context, search.Options) ([]search.Match, error) {
		return nil, nil
	})
	var notice *cliutil.Notice
	if !errors.As(err, &notice) {
		t.Fatalf("error = %v, want Notice", err)
	}
}

func TestRunCancel(t *testing.T) {
	f := &afxtest.MockFinder{Err: fuzzyfinder.ErrAbort}
	err := run(context.Background(), &afxtest.MockAFX{}, f, search.Options{}, func(context.Context, search.Options) ([]search.Match, error) {
		return []search.Match{{Path: `C:\work\a.go`}}, nil
	})
	if err != nil {
		t.Fatalf("cancel error = %v", err)
	}
}

func TestRunQueryError(t *testing.T) {
	err := run(context.Background(), &afxtest.MockAFX{}, &afxtest.MockFinder{}, search.Options{}, func(context.Context, search.Options) ([]search.Match, error) {
		return nil, errors.New("query failed")
	})
	if err == nil || !strings.Contains(err.Error(), "キーワード検索に失敗しました") {
		t.Fatalf("error = %v", err)
	}
}
