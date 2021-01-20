//go:generate go run gen.go -inputPath=spec.json -outputPath=config.go

package config

import (
	"fmt"
	"log"

	"github.com/pkg/errors"
)

// Rule is a configuration rule.
// Each rule contains a glob pattern matching the path of the current file.
// If the rule matches the current file, its configuration will be applied.
type Rule struct {
	Name    string        `json:"name"`
	Pattern string        `json:"pattern"`
	Config  PartialConfig `json:"config"`
}

// RuleSet is a set of configuration rules.
// If multiple rules match a file path, they are applied in order.
type RuleSet struct {
	Rules []Rule
}

func (rs *RuleSet) Validate() error {
	for _, rule := range rs.Rules {
		err := rule.Config.Validate()
		if err != nil {
			msg := fmt.Sprintf("Validation error in config rule %s", rule.Name)
			return errors.Wrapf(err, msg)
		}
	}
	return nil
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
