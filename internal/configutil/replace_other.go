//go:build !windows

package configutil

import "os"

func replaceFile(source, destination string) error {
	return os.Rename(source, destination)
}
