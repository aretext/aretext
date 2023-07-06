//go:build windows

package app

import _ "embed"

//go:embed default-config-windows.yaml
var DefaultConfigYaml []byte
