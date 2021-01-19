package config

const DefaultTabSize = 4

// Config is a configuration for the editor.
type Config struct {
	SyntaxLanguage string `json:"syntaxLanguage"`
	TabSize        int    `json:"tabSize"`
}

// DefaultConfig constructs a configuration with default values.
func DefaultConfig() Config {
	return Config{
		SyntaxLanguage: "undefined",
		TabSize:        DefaultTabSize,
	}
}

// Apply overrides the base config values with values from another configuration.
func (c *Config) Apply(overlay Config) {
	if overlay.SyntaxLanguage != "" {
		c.SyntaxLanguage = overlay.SyntaxLanguage
	}

	if overlay.TabSize > 0 {
		c.TabSize = overlay.TabSize
	}
}
