package clipboard

import "errors"

var systemClipboardNotSupportedErr = errors.New("System clipboard not supported")

type systemClipboardTool struct {
	copyCmd  []string
	pasteCmd []string
}

func setInSystemClipboard(text string) error {
	// TODO: lookup, then exec
	return nil
}

func getFromSystemClipboard() (string, error) {
	// TODO: lookup, then exec
	return "", nil
}
