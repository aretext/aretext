package app

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/aretext/aretext/config"
)

//go:embed default-config.yaml
var DefaultConfigYaml []byte

// LoadOrCreateConfig loads the config file if it exists and creates a default config file otherwise.
func LoadOrCreateConfig(forceDefaultConfig bool) (config.RuleSet, error) {
	if forceDefaultConfig {
		log.Printf("Using default config\n")
		return unmarshalRuleSet(DefaultConfigYaml)
	}

	path, err := xdg.ConfigFile("aretext/config.yaml")
	if err != nil {
		return nil, err
	}

	log.Printf("Loading config from '%s'\n", path)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		log.Printf("Writing default config to '%s'\n", path)
		if err := saveDefaultConfig(path); err != nil {
			return nil, errors.Wrapf(err, "Error writing default config to '%s'", path)
		}
		return unmarshalRuleSet(DefaultConfigYaml)
	} else if err != nil {
		return nil, errors.Wrapf(err, "Error loading config from '%s'", path)
	}

	ruleSet, err := unmarshalRuleSet(data)
	if err != nil {
		return nil, err
	}

	if err := ruleSet.Validate(); err != nil {
		errMsg := err.Error()
		helpMsg := fmt.Sprintf("To edit the config, try\n\taretext -noconfig %s", path)
		newErrMsg := fmt.Sprintf("Invalid configuration: %s\n%s", errMsg, helpMsg)
		return nil, errors.New(newErrMsg)
	}

	return ruleSet, nil
}

func unmarshalRuleSet(data []byte) (config.RuleSet, error) {
	var rules []config.Rule
	if err := yaml.Unmarshal(data, &rules); err != nil {
		return nil, errors.Wrap(err, "yaml")
	}
	return config.RuleSet(rules), nil
}

func saveDefaultConfig(path string) error {
	dirPath := filepath.Dir(path)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return errors.Wrap(err, "os.MkdirAll")
	}
	if err := os.WriteFile(path, DefaultConfigYaml, 0644); err != nil {
		return errors.Wrap(err, "os.WriteFile")
	}
	return nil
}
