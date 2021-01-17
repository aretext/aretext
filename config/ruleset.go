package config

import (
	"log"
)

// Rule is a configuration rule.
// Each rule contains a glob pattern matching the path of the current file.
// If the rule matches the current file, its configuration will be applied.
type Rule struct {
	Name    string `json:"name"`
	Pattern string `json:"pattern"`
	Config  Config `json:"config"`
}

// RuleSet is a set of configuration rules.
// If multiple rules match a file path, they are applied in order.
type RuleSet struct {
	Rules []Rule
}

// ConfigForPath returns a configuration for a specific file path.
// Rules that match the file path are applied in order to produce the configuration.
func (rs *RuleSet) ConfigForPath(path string) Config {
	config := DefaultConfig()
	for _, rule := range rs.Rules {
		if GlobMatch(rule.Pattern, path) {
			log.Printf("Applying config rule '%s' with pattern '%s' for path '%s'\n", rule.Name, rule.Pattern, path)
			config.Apply(rule.Config)
		}
	}
	return config
}
