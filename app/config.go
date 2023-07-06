package app

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"

	"github.com/aretext/aretext/config"
)

// ConfigPath returns the path to the configuration file.
func ConfigPath() (string, error) {
	path := filepath.Join("aretext", "config.yaml")
	return xdg.ConfigFile(path)
}

// LoadOrCreateConfig loads the config file if it exists and creates a default config file otherwise.
func LoadOrCreateConfig(forceDefaultConfig bool) (config.RuleSet, error) {
	if forceDefaultConfig {
		log.Printf("Using default config\n")
		return unmarshalRuleSet(DefaultConfigYaml)
	}

	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	log.Printf("Loading config from %q\n", path)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		log.Printf("Writing default config to %q\n", path)
		if err := saveDefaultConfig(path); err != nil {
			return nil, fmt.Errorf("Error writing default config to %q: %w", path, err)
		}
		return unmarshalRuleSet(DefaultConfigYaml)
	} else if err != nil {
		return nil, fmt.Errorf("Error loading config from %q: %w", path, err)
	}

	ruleSet, err := unmarshalRuleSet(data)
	if err != nil {
		return nil, err
	}

	if err := ruleSet.Validate(); err != nil {
		errMsg := err.Error()
		helpMsg := fmt.Sprintf("To edit the config, try\n\taretext -noconfig %s", path)
		return nil, fmt.Errorf("Invalid configuration: %s\n%s", errMsg, helpMsg)
	}

	return ruleSet, nil
}

func unmarshalRuleSet(data []byte) (config.RuleSet, error) {
	var rules []config.Rule
	if err := yaml.Unmarshal(data, &rules); err != nil {
		return nil, fmt.Errorf("yaml.Unmarshal: %w", err)
	}
	return config.RuleSet(rules), nil
}

func saveDefaultConfig(path string) error {
	dirPath := filepath.Dir(path)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("os.MkdirAll: %w", err)
	}
	if err := os.WriteFile(path, DefaultConfigYaml, 0644); err != nil {
		return fmt.Errorf("os.WriteFile: %w", err)
	}
	return nil
}
