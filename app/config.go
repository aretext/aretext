package app

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aretext/aretext/config"
	"github.com/aretext/aretext/syntax"
	"github.com/pkg/errors"
)

// LoadOrCreateConfig loads the config file if it exists and creates a default config file otherwise.
func LoadOrCreateConfig(forceDefaultConfig bool) (config.RuleSet, error) {
	if forceDefaultConfig {
		log.Printf("Using default config\n")
		return defaultConfigRuleSet(), nil
	}

	path, err := defaultPath()
	if err != nil {
		return config.RuleSet{}, err
	}

	log.Printf("Loading config from '%s'\n", path)

	rs, err := config.LoadRuleSet(path)
	if os.IsNotExist(err) {
		log.Printf("Writing default config to '%s'\n", path)
		rs = defaultConfigRuleSet()
		err = config.SaveRuleSet(path, rs)
		if err != nil {
			log.Printf("Could not save config to '%s': %v\n", path, errors.Wrapf(err, "config.SaveRuleSet"))
		}
	} else if err != nil {
		return config.RuleSet{}, errors.Wrapf(err, fmt.Sprintf("Error loading config from '%s'", path))
	}

	err = rs.Validate()
	if err != nil {
		return config.RuleSet{}, err
	}

	return rs, nil
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

func defaultConfigRuleSet() config.RuleSet {
	return config.RuleSet{
		Rules: []config.Rule{
			{
				Name:    "default",
				Pattern: "**",
				Config:  config.DefaultConfig().ToPartial(),
			},
			{
				Name:    "json",
				Pattern: "**/*.json",
				Config: config.Config{
					SyntaxLanguage: syntax.LanguageJson.String(),
				}.ToPartial(),
			},
		},
	}
}
