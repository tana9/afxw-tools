package configutil

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// atomicWrite は一時ファイルに書き込んでから元のファイルに置換することで、書き込みエラー時にファイル破損を防ぎます。
func atomicWrite(path string, write func(io.Writer) error) error {
	tmp, tmpPath, err := createTempFile(path)
	if err != nil {
		return err
	}
	defer cleanupTempFile(tmp, tmpPath)

	if err := write(tmp); err != nil {
		return err
	}
	if err := syncAndClose(tmp); err != nil {
		return err
	}
	if err := replaceFile(tmpPath, path); err != nil {
		return fmt.Errorf("設定ファイルの置換に失敗しました: %w", err)
	}
	return nil
}

// createTempFile はpath と同じディレクトリに一時ファイルを作成し、既存ファイルが存在する場合はその権限をコピーします。
func createTempFile(path string) (*os.File, string, error) {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return nil, "", fmt.Errorf("一時ファイルの作成に失敗しました: %w", err)
	}
	tmpPath := tmp.Name()

	if info, err := os.Stat(path); err == nil {
		if err := tmp.Chmod(info.Mode().Perm()); err != nil {
			cleanupTempFile(tmp, tmpPath)
			return nil, "", fmt.Errorf("一時ファイルの権限設定に失敗しました: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		cleanupTempFile(tmp, tmpPath)
		return nil, "", fmt.Errorf("既存ファイルの確認に失敗しました: %w", err)
	} else if err := tmp.Chmod(0644); err != nil {
		cleanupTempFile(tmp, tmpPath)
		return nil, "", fmt.Errorf("一時ファイルの権限設定に失敗しました: %w", err)
	}
	return tmp, tmpPath, nil
}

// syncAndClose はファイルの内容をディスクに同期してからファイルをクローズします。
func syncAndClose(file *os.File) error {
	if err := file.Sync(); err != nil {
		return fmt.Errorf("一時ファイルの同期に失敗しました: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("一時ファイルのクローズに失敗しました: %w", err)
	}
	return nil
}

// cleanupTempFile は一時ファイルをクローズして削除します。エラーは無視されます。
func cleanupTempFile(file *os.File, path string) {
	_ = file.Close()
	_ = os.Remove(path)
}
