package config

// Config is a configuration for the editor.
type Config struct {
	SyntaxLanguage string `json:"syntaxLanguage"`
}

// DefaultConfig constructs a configuration with default values.
func DefaultConfig() Config {
	return Config{
		SyntaxLanguage: "undefined",
	}
}

// Apply overrides the base config values with values from another configuration.
func (c *Config) Apply(overlay Config) {
	if overlay.SyntaxLanguage != "" {
		c.SyntaxLanguage = overlay.SyntaxLanguage
	}
}
