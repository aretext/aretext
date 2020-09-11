package repl

import (
	"fmt"
	"net"
)

// ApiConfig describes how the REPL can call the Aretext API (RPC server).
type ApiConfig struct {
	addr net.Addr
	key  string
}

func NewApiConfig(addr net.Addr, key string) *ApiConfig {
	return &ApiConfig{
		addr: addr,
		key:  key,
	}
}

func (c *ApiConfig) EnvVars() []string {
	return []string{
		fmt.Sprintf("API_ADDRESS=%s", c.addr.String()),
		fmt.Sprintf("API_KEY=%s", c.key),
	}
}
