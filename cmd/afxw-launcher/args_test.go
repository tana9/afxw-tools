package main

import (
	"errors"
	"reflect"
	"testing"

	"github.com/tana9/afxw-tools/internal/afx"
	"github.com/tana9/afxw-tools/internal/afxtest"
)

func TestExpandArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		currentFile string
		markedFiles []string
		want        []string
	}{
		{
			name: "no placeholders",
			args: []string{"--flag", "value"},
			want: []string{"--flag", "value"},
		},
		{
			name:        "file standalone",
			args:        []string{"{file}"},
			currentFile: `C:\dir\file.txt`,
			want:        []string{`C:\dir\file.txt`},
		},
		{
			name:        "file embedded",
			args:        []string{"--input={file}"},
			currentFile: `C:\dir\file.txt`,
			want:        []string{`--input=C:\dir\file.txt`},
		},
		{
			name:        "files standalone",
			args:        []string{"{files}"},
			markedFiles: []string{`C:\a.txt`, `C:\b.txt`},
			want:        []string{`C:\a.txt`, `C:\b.txt`},
		},
		{
			name:        "files embedded",
			args:        []string{"--file={files}"},
			markedFiles: []string{`C:\a.txt`, `C:\b.txt`},
			want:        []string{`--file=C:\a.txt`, `--file=C:\b.txt`},
		},
		{
			name:        "file and files in same argument",
			args:        []string{"{file}:{files}"},
			currentFile: `C:\current.txt`,
			markedFiles: []string{`C:\a.txt`, `C:\b.txt`},
			want:        []string{`C:\current.txt:C:\a.txt`, `C:\current.txt:C:\b.txt`},
		},
		{
			name:        "mixed arguments",
			args:        []string{"--flag", "{files}", "--current", "{file}"},
			currentFile: `C:\current.txt`,
			markedFiles: []string{`C:\a.txt`, `C:\b.txt`},
			want:        []string{"--flag", `C:\a.txt`, `C:\b.txt`, "--current", `C:\current.txt`},
		},
		{
			name: "empty files",
			args: []string{"before", "{files}", "after"},
			want: []string{"before", "after"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandArgs(tt.args, tt.currentFile, tt.markedFiles)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expandArgs(%v, %q, %v) = %v, want %v", tt.args, tt.currentFile, tt.markedFiles, got, tt.want)
			}
		})
	}
}

func TestResolveArgs(t *testing.T) {
	t.Run("no placeholders skips AFX connection", func(t *testing.T) {
		called := false
		newClient := func() (afx.Client, error) {
			called = true
			return nil, errors.New("must not be called")
		}
		args := []string{"--flag"}
		got, err := resolveArgsWith(args, newClient)
		if err != nil {
			t.Fatalf("resolveArgsWith() error = %v", err)
		}
		if called || !reflect.DeepEqual(got, args) {
			t.Errorf("called = %v, args = %v", called, got)
		}
	})

	tests := []struct {
		name    string
		args    []string
		client  *afxtest.MockClient
		newErr  error
		want    []string
		wantErr string
	}{
		{
			name:   "file",
			args:   []string{"{file}"},
			client: &afxtest.MockClient{CurrentFileResult: `C:\current.txt`},
			want:   []string{`C:\current.txt`},
		},
		{
			name:   "files",
			args:   []string{"{files}"},
			client: &afxtest.MockClient{MarkedFilesResult: []string{`C:\a.txt`, `C:\b.txt`}},
			want:   []string{`C:\a.txt`, `C:\b.txt`},
		},
		{
			name: "file and files",
			args: []string{"{file}", "{files}"},
			client: &afxtest.MockClient{
				CurrentFileResult: `C:\current.txt`,
				MarkedFilesResult: []string{`C:\a.txt`},
			},
			want: []string{`C:\current.txt`, `C:\a.txt`},
		},
		{name: "connection error", args: []string{"{file}"}, newErr: errors.New("connect failed"), wantErr: "afxw.objへの接続に失敗しました: connect failed"},
		{name: "current file error", args: []string{"{file}"}, client: &afxtest.MockClient{CurrentFileErr: errors.New("current failed")}, wantErr: "カレントファイルの取得に失敗しました: current failed"},
		{name: "marked files error", args: []string{"{files}"}, client: &afxtest.MockClient{MarkedFilesErr: errors.New("marked failed")}, wantErr: "マーク済みファイルの取得に失敗しました: marked failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newClient := func() (afx.Client, error) { return tt.client, tt.newErr }
			got, err := resolveArgsWith(tt.args, newClient)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("resolveArgsWith() error = %v", err)
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("resolveArgsWith() = %v, want %v", got, tt.want)
				}
			} else if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("resolveArgsWith() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}
