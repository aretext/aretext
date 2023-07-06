//go:build !windows

package app

import _ "embed"

//go:embed default-config-unix.yaml
var DefaultConfigYaml []byte
