//go:build !windows

package configutil

import "os"

// replaceFile は source を destination へ os.Rename で移動して置き換えます。
func replaceFile(source, destination string) error {
	return os.Rename(source, destination)
}
