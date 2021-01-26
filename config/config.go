package config

import (
	"errors"
	"log"
)

const DefaultSyntaxLanguage = "undefined"
const DefaultTabSize = 4
const DefaultTabExpand = false
const DefaultAutoIndent = false

// Config is a configuration for the editor.
type Config struct {
	// Language used for syntax highlighting.
	SyntaxLanguage string

	// Size of a tab character in columns.
	TabSize int

	// If enabled, the tab key inserts spaces.
	TabExpand bool

	// If enabled, indent a new line to match indentation of the previous line.
	AutoIndent bool
}

// ConfigFromUntypedMap constructs a configuration from an untyped map.
// The map is usually loaded from a JSON document.
func ConfigFromUntypedMap(m map[string]interface{}) Config {
	return Config{
		SyntaxLanguage: stringOrDefault(m, "syntaxLanguage", DefaultSyntaxLanguage),
		TabSize:        intOrDefault(m, "tabSize", DefaultTabSize),
		TabExpand:      boolOrDefault(m, "tabExpand", DefaultTabExpand),
		AutoIndent:     boolOrDefault(m, "autoIndent", DefaultAutoIndent),
	}
}

// DefaultConfig is a configuration with all keys set to default values.
func DefaultConfig() Config {
	return ConfigFromUntypedMap(nil)
}

// Validate checks that the values in the configuration are valid.
func (c Config) Validate() error {
	if c.TabSize < 1 {
		return errors.New("TabSize must be greater than zero")
	}

	return nil
}

func stringOrDefault(m map[string]interface{}, key string, defaultVal string) string {
	v, ok := m[key]
	if !ok {
		return defaultVal
	}

	s, ok := v.(string)
	if !ok {
		log.Printf("Could not decode string for config key '%s'\n", key)
		return defaultVal
	}

	return s
}

func intOrDefault(m map[string]interface{}, key string, defaultVal int) int {
	v, ok := m[key]
	if !ok {
		return defaultVal
	}

	switch v.(type) {
	case int:
		return v.(int)
	case float64:
		return int(v.(float64))
	default:
		log.Printf("Could not decode int for config key '%s'\n", key)
		return defaultVal
	}
}

func boolOrDefault(m map[string]interface{}, key string, defaultVal bool) bool {
	v, ok := m[key]
	if !ok {
		return defaultVal
	}

	b, ok := v.(bool)
	if !ok {
		log.Printf("Could not decode bool for config key '%s'\n", key)
		return defaultVal
	}

	return b
}
