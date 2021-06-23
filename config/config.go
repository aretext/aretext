package config

import (
	"errors"
	"fmt"
	"log"
)

const DefaultSyntaxLanguage = "undefined"
const DefaultTabSize = 4
const DefaultTabExpand = false
const DefaultShowTabs = false
const DefaultAutoIndent = false
const DefaultShowLineNumbers = false

// Config is a configuration for the editor.
type Config struct {
	// Language used for syntax highlighting.
	SyntaxLanguage string

	// Size of a tab character in columns.
	TabSize int

	// If enabled, the tab key inserts spaces.
	TabExpand bool

	// If enabled, display tab characters in the document.
	ShowTabs bool

	// If enabled, indent a new line to match indentation of the previous line.
	AutoIndent bool

	// If enabled, show line numbers in the left margin.
	ShowLineNumbers bool

	// User-defined commands to include in the menu.
	MenuCommands []MenuCommandConfig

	// Glob patterns for directories to exclude from file search.
	HideDirectories []string

	// Style overrides.
	Styles map[string]StyleConfig
}

const (
	CmdModeSilent        = "silent"        // accepts no input and any output is discarded.
	CmdModeTerminal      = "terminal"      // takes control of the terminal.
	CmdModeInsert        = "insert"        // output is inserted into the document at the cursor position, replacing any selection.
	CmdModeFileLocations = "fileLocations" // output is interpreted as a list of file locations that can be opened in the editor.
)

// MenuCommandConfig is a configuration for a user-defined menu item.
type MenuCommandConfig struct {
	// Name is the displayed name of the menu.
	Name string

	// ShellCmd is the shell command to execute when the menu item is selected.
	ShellCmd string

	// Mode controls how the command's input and output are handled.
	Mode string
}

// Names of styles that can be overridden by configuration.
const (
	StyleLineNum       = "lineNum"
	StyleTokenOperator = "tokenOperator"
	StyleTokenKeyword  = "tokenKeyword"
	StyleTokenNumber   = "tokenNumber"
	StyleTokenString   = "tokenString"
	StyleTokenComment  = "tokenComment"
)

// StyleConfig is a configuration for how text should be displayed.
type StyleConfig struct {
	// Color is either a W3C color name ("green", "red", etc.)
	// or a hexadecimal RGB code (formatted like "#ffffff").
	// The named colors respect the palette set by the terminal.
	// The hexadecimal colors represent the exact 24-bit RGB value
	// for the color; if the terminal does not support true-color,
	// we fallback to a similar 8-bit color.
	Color string
}

// ConfigFromUntypedMap constructs a configuration from an untyped map.
// The map is usually loaded from a JSON document.
func ConfigFromUntypedMap(m map[string]interface{}) Config {
	return Config{
		SyntaxLanguage:  stringOrDefault(m, "syntaxLanguage", DefaultSyntaxLanguage),
		TabSize:         intOrDefault(m, "tabSize", DefaultTabSize),
		TabExpand:       boolOrDefault(m, "tabExpand", DefaultTabExpand),
		ShowTabs:        boolOrDefault(m, "showTabs", DefaultShowTabs),
		AutoIndent:      boolOrDefault(m, "autoIndent", DefaultAutoIndent),
		ShowLineNumbers: boolOrDefault(m, "showLineNumbers", DefaultShowLineNumbers),
		MenuCommands:    menuCommandsFromSlice(sliceOrNil(m, "menuCommands")),
		HideDirectories: stringSliceOrNil(m, "hideDirectories"),
		Styles:          stylesFromMap(mapOrNil(m, "styles")),
	}
}

// Validate checks that the values in the configuration are valid.
func (c Config) Validate() error {
	if c.TabSize < 1 {
		return errors.New("TabSize must be greater than zero")
	}

	for _, cmd := range c.MenuCommands {
		if cmd.Mode != CmdModeSilent && cmd.Mode != CmdModeTerminal && cmd.Mode != CmdModeInsert && cmd.Mode != CmdModeFileLocations {
			msg := fmt.Sprintf(
				"Menu command '%s' must have mode set to either '%s', '%s', '%s', or '%s'",
				cmd.Name,
				CmdModeSilent,
				CmdModeTerminal,
				CmdModeInsert,
				CmdModeFileLocations,
			)
			return errors.New(msg)
		}
	}

	return nil
}

func stringOrDefault(m map[string]interface{}, key string, defaultVal string) string {
	v, ok := m[key]
	if !ok {
		return defaultVal
	}

	s, ok := v.(string)
	if !ok {
		log.Printf("Could not decode string for config key '%s'\n", key)
		return defaultVal
	}

	return s
}

func intOrDefault(m map[string]interface{}, key string, defaultVal int) int {
	v, ok := m[key]
	if !ok {
		return defaultVal
	}

	switch v.(type) {
	case int:
		return v.(int)
	case float64:
		return int(v.(float64))
	default:
		log.Printf("Could not decode int for config key '%s'\n", key)
		return defaultVal
	}
}

func boolOrDefault(m map[string]interface{}, key string, defaultVal bool) bool {
	v, ok := m[key]
	if !ok {
		return defaultVal
	}

	b, ok := v.(bool)
	if !ok {
		log.Printf("Could not decode bool for config key '%s'\n", key)
		return defaultVal
	}

	return b
}

func sliceOrNil(m map[string]interface{}, key string) []interface{} {
	v, ok := m[key]
	if !ok {
		return nil
	}

	s, ok := v.([]interface{})
	if !ok {
		log.Printf("Could not decode slice for config key '%s'\n", key)
		return nil
	}

	return s
}

func stringSliceOrNil(m map[string]interface{}, key string) []string {
	slice := sliceOrNil(m, key)
	if slice == nil {
		return nil
	}

	stringSlice := make([]string, 0, len(slice))
	for i := 0; i < len(slice); i++ {
		s, ok := (slice[i]).(string)
		if !ok {
			log.Printf("Could not decode string in slice for config key '%s'\n", key)
			continue
		}
		stringSlice = append(stringSlice, s)
	}
	return stringSlice
}

func mapOrNil(m map[string]interface{}, key string) map[string]interface{} {
	v, ok := m[key]
	if !ok {
		return nil
	}

	subMap, ok := v.(map[string]interface{})
	if !ok {
		log.Printf("Could not decode map for config key '%s'\n", key)
		return nil
	}

	return subMap
}

func menuCommandsFromSlice(s []interface{}) []MenuCommandConfig {
	result := make([]MenuCommandConfig, 0, len(s))
	for _, m := range s {
		menuMap, ok := m.(map[string]interface{})
		if !ok {
			log.Printf("Could not decode menu command map from %v\n", m)
			continue
		}

		result = append(result, MenuCommandConfig{
			Name:     stringOrDefault(menuMap, "name", "[nil]"),
			ShellCmd: stringOrDefault(menuMap, "shellCmd", ""),
			Mode:     stringOrDefault(menuMap, "mode", CmdModeTerminal),
		})
	}
	return result
}

func stylesFromMap(m map[string]interface{}) map[string]StyleConfig {
	result := make(map[string]StyleConfig, len(m))
	for k, v := range m {
		styleMap, ok := v.(map[string]interface{})
		if !ok {
			log.Printf("Could not decode style map from %v\n", v)
			continue
		}

		color := stringOrDefault(styleMap, "color", "")
		result[k] = StyleConfig{Color: color}
	}
	return result
}
