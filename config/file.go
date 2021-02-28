package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// LoadRuleSet loads configuration rules from a file.
func LoadRuleSet(path string) (RuleSet, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		// Return the error directly so callers can use os.IsNotExist(err) to check if the file exists.
		return RuleSet{}, err
	}

	var rules []Rule
	err = json.Unmarshal(data, &rules)
	if err != nil {
		return RuleSet{}, errors.Wrapf(err, "json.Unmarshal")
	}

	return RuleSet{rules}, nil
}

// SaveRuleSet saves configuration rules to a file.
func SaveRuleSet(path string, rs RuleSet) error {
	data, err := json.MarshalIndent(rs.Rules, "", "  ")
	if err != nil {
		return errors.Wrapf(err, "json.Marshal")
	}

	dirPath := filepath.Dir(path)
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return errors.Wrapf(err, "os.MkdirAll")
	}

	err = ioutil.WriteFile(path, data, 0644)
	if err != nil {
		return errors.Wrapf(err, "ioutil.WriteFile")
	}

	return nil
}
