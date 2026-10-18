package configutil

import "golang.org/x/sys/windows"

// replaceFile は source を destination へ置換して移動します。書き込みを確実に反映させるため MoveFileEx を使います。
func replaceFile(source, destination string) error {
	sourcePtr, err := windows.UTF16PtrFromString(source)
	if err != nil {
		return err
	}
	destinationPtr, err := windows.UTF16PtrFromString(destination)
	if err != nil {
		return err
	}
	return windows.MoveFileEx(
		sourcePtr,
		destinationPtr,
		windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH,
	)
}
