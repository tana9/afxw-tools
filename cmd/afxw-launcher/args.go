package main

import (
	"fmt"
	"strings"

	"github.com/tana9/afxw-tools/internal/afx"
)

// resolveArgs はargs内のプレースホルダーを実際の値に置換します。
// {file}  : カーソル位置のファイルのフルパスに展開されます。
// {files} : マーク済みファイルのフルパス一覧に展開されます（マークなしの場合はカーソルファイル）。
//
//	1つの引数が複数の引数に展開されます。
func resolveArgs(args []string) ([]string, error) {
	hasFile, hasFiles := false, false
	for _, arg := range args {
		if strings.Contains(arg, "{files}") {
			hasFiles = true
		} else if strings.Contains(arg, "{file}") {
			hasFile = true
		}
	}
	if !hasFile && !hasFiles {
		return args, nil
	}

	a, err := afx.NewOleAFX()
	if err != nil {
		return nil, fmt.Errorf("afxw.objへの接続に失敗しました: %w", err)
	}
	defer a.Close()

	var currentFile string
	if hasFile {
		if currentFile, err = a.GetCurrentFile(); err != nil {
			return nil, fmt.Errorf("カレントファイルの取得に失敗しました: %w", err)
		}
	}

	var markedFiles []string
	if hasFiles {
		if markedFiles, err = a.GetMarkedFiles(); err != nil {
			return nil, fmt.Errorf("マーク済みファイルの取得に失敗しました: %w", err)
		}
	}

	return expandArgs(args, currentFile, markedFiles), nil
}

// expandArgs はプレースホルダーを展開する純粋関数です。
// {files} が引数単体の場合は markedFiles の各要素に展開されます。
func expandArgs(args []string, currentFile string, markedFiles []string) []string {
	resolved := make([]string, 0, len(args))
	for _, arg := range args {
		if strings.Contains(arg, "{files}") {
			if arg == "{files}" {
				resolved = append(resolved, markedFiles...)
			} else {
				for _, f := range markedFiles {
					resolved = append(resolved, strings.ReplaceAll(arg, "{files}", f))
				}
			}
		} else {
			resolved = append(resolved, strings.ReplaceAll(arg, "{file}", currentFile))
		}
	}
	return resolved
}
