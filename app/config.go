package app

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aretext/aretext/config"
	"github.com/pkg/errors"
)

//go:embed default-config.json
var DefaultConfigJson []byte

// LoadOrCreateConfig loads the config file if it exists and creates a default config file otherwise.
func LoadOrCreateConfig(forceDefaultConfig bool) (config.RuleSet, error) {
	if forceDefaultConfig {
		log.Printf("Using default config\n")
		return unmarshalRuleSet(DefaultConfigJson)
	}

	path, err := defaultPath()
	if err != nil {
		return nil, err
	}

	log.Printf("Loading config from '%s'\n", path)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		log.Printf("Writing default config to '%s'\n", path)
		if err := saveDefaultConfig(path); err != nil {
			return nil, errors.Wrapf(err, fmt.Sprintf("Error writing default config to '%s'", path))
		}
		return unmarshalRuleSet(DefaultConfigJson)
	} else if err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("Error loading config from '%s'", path))
	}

	return unmarshalRuleSet(data)
}

// defaultPath returns the path to the user's configuration file.
func defaultPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrapf(err, "os.UserHomeDir")
	}
	path := filepath.Join(homeDir, ".config", "aretext", "config.json")
	return path, nil
}

func unmarshalRuleSet(data []byte) (config.RuleSet, error) {
	var rules []config.Rule
	if err := json.Unmarshal(data, &rules); err != nil {
		return nil, errors.Wrapf(err, "json.Unmarshal")
	}
	return config.RuleSet(rules), nil
}

func saveDefaultConfig(path string) error {
	dirPath := filepath.Dir(path)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return errors.Wrapf(err, "os.MkdirAll")
	}
	if err := os.WriteFile(path, DefaultConfigJson, 0644); err != nil {
		return errors.Wrapf(err, "os.WriteFile")
	}
	return nil
}
