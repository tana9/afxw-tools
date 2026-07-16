// Package selectnav は「候補一覧からfuzzyfinderで選択してあふwで移動する」共通フローを提供します。
package selectnav

import (
	"errors"
	"fmt"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/tana9/afxw-tools/internal/afx"
	"github.com/tana9/afxw-tools/internal/finder"
)

// SelectAndMove は候補から1つをfuzzyfinderで選択し、あふwで選択先ディレクトリに移動します。
// 候補が空の場合とキャンセル(Esc/Ctrl+C)の場合は何もせず正常終了します。
func SelectAndMove(a afx.AFX, f finder.Finder, dirs []string) error {
	if len(dirs) == 0 {
		return nil
	}

	idx, err := f.Find(dirs)
	if err != nil {
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return nil // ESC / Ctrl+C でキャンセル
		}
		return err
	}

	if err := a.EXCD(dirs[idx]); err != nil {
		return fmt.Errorf("ディレクトリ移動に失敗しました: %w", err)
	}

	return nil
}
