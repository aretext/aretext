package config

import (
	"log"

	"github.com/aretext/aretext/file"
)

// Rule is a configuration rule.
// Each rule contains a glob pattern matching the path of the current file.
// If the rule matches the current file, its configuration will be applied.
type Rule struct {
	Name    string         `json:"name"`
	Pattern string         `json:"pattern"`
	Config  map[string]any `json:"config"`
}

// RuleSet is a set of configuration rules.
// If multiple rules match a file path, they are applied in order.
type RuleSet []Rule

// Rules that match the file path are applied in order to produce the configuration.
func (rs RuleSet) ConfigForPath(path string) Config {
	c := make(map[string]any, 0)
	for _, rule := range rs {
		if file.GlobMatch(rule.Pattern, path) {
			log.Printf("Applying config rule %q with pattern %q for path %q\n", rule.Name, rule.Pattern, path)
			c = MergeRecursive(c, rule.Config).(map[string]any)
		}
	}
	log.Printf("Resolved config for path %q: %#v\n", path, c)
	return ConfigFromUntypedMap(c)
}

// Validate checks whether every rule in the set has a valid configuration.
func (rs RuleSet) Validate() error {
	for _, r := range rs {
		c := ConfigFromUntypedMap(r.Config)
		err := c.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}
