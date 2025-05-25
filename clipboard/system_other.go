//go:build !darwin && !linux
package clipboard

func setInSystemClipboard(text string) error {
	return systemClipboardNotSupportedErr
}

func getFromSystemClipboard() (string, error) {
	return "", systemClipboardNotSupportedErr
}
