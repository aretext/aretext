package config

import (
	"errors"
	"fmt"
	"log"
)

const DefaultSyntaxLanguage = "plaintext"
const DefaultTabSize = 4
const DefaultTabExpand = false
const DefaultShowTabs = false
const DefaultShowSpaces = false
const DefaultShowUnicode = false
const DefaultAutoIndent = false
const DefaultShowLineNumbers = false
const DefaultLineWrap = LineWrapCharacter
const DefaultLineNumberMode = LineNumberModeAbsolute

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

	// If enabled, display space characters in the document.
	ShowSpaces bool

	// If enabled, show codepoints for non-ascii unicode.
	ShowUnicode bool

	// If enabled, indent a new line to match indentation of the previous line.
	AutoIndent bool

	// If enabled, show line numbers in the left margin.
	ShowLineNumbers bool

	// Display mode for line numbers (relative or absolute)
	LineNumberMode string

	// LineWrap controls how lines are soft-wrapped.
	LineWrap string

	// User-defined commands to include in the menu.
	MenuCommands []MenuCommandConfig

	// Glob patterns for files or directories to exclude from file search.
	HidePatterns []string

	// (DEPRECATED) Glob patterns for directories to exclude from file search.
	HideDirectories []string

	// Style overrides.
	Styles map[string]StyleConfig
}

const (
	LineWrapCharacter = "character" // Break lines between any two characters.
	LineWrapWord      = "word"      // Break lines only between words.
)

const (
	CmdModeSilent        = "silent"        // accepts no input and any output is discarded.
	CmdModeTerminal      = "terminal"      // takes control of the terminal.
	CmdModeInsert        = "insert"        // output is inserted into the document at the cursor position, replacing any selection.
	CmdModeInsertChoice  = "insertChoice"  // user can select one line from the output to insert into the document.
	CmdModeFileLocations = "fileLocations" // output is interpreted as a list of file locations that can be opened in the editor.
	CmdModeWorkingDir    = "workingDir"    // output is interpreted as a list of directories to set as the current working directory.
)

type LineNumberMode string

const (
	LineNumberModeAbsolute LineNumberMode = "absolute" // shows the line number from the beginning of the file
	LineNumberModeRelative LineNumberMode = "relative" // shows the line number relative to the cursor
)

// MenuCommandConfig is a configuration for a user-defined menu item.
type MenuCommandConfig struct {
	// Name is the displayed name of the menu.
	Name string

	// ShellCmd is the shell command to execute when the menu item is selected.
	ShellCmd string

	// Mode controls how the command's input and output are handled.
	Mode string

	// Save controls whether the document will be saved before running the command.
	Save bool
}

// Names of styles that can be overridden by configuration.
const (
	StyleLineNum        = "lineNum"
	StyleEscapedUnicode = "escapedUnicode"
	StyleTokenOperator  = "tokenOperator"
	StyleTokenKeyword   = "tokenKeyword"
	StyleTokenNumber    = "tokenNumber"
	StyleTokenString    = "tokenString"
	StyleTokenComment   = "tokenComment"
	StyleTokenCustom1   = "tokenCustom1"
	StyleTokenCustom2   = "tokenCustom2"
	StyleTokenCustom3   = "tokenCustom3"
	StyleTokenCustom4   = "tokenCustom4"
	StyleTokenCustom5   = "tokenCustom5"
	StyleTokenCustom6   = "tokenCustom6"
	StyleTokenCustom7   = "tokenCustom7"
	StyleTokenCustom8   = "tokenCustom8"
	StyleTokenCustom9   = "tokenCustom9"
	StyleTokenCustom10  = "tokenCustom10"
	StyleTokenCustom11  = "tokenCustom11"
	StyleTokenCustom12  = "tokenCustom12"
	StyleTokenCustom13  = "tokenCustom13"
	StyleTokenCustom14  = "tokenCustom14"
	StyleTokenCustom15  = "tokenCustom15"
	StyleTokenCustom16  = "tokenCustom16"
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

	// BackgroundColor is the color of the background.
	// It has the same format as Color.
	BackgroundColor string

	// Bold sets the bold attribute.
	Bold bool

	// Italic sets the italic attribute.
	Italic bool

	// Underline sets the underline attribute.
	Underline bool

	// Strikethrough sets the strikethrough attribute.
	StrikeThrough bool
}

// ConfigFromUntypedMap constructs a configuration from an untyped map.
func ConfigFromUntypedMap(m map[string]any) Config {
	return Config{
		SyntaxLanguage:  stringOrDefault(m, "syntaxLanguage", DefaultSyntaxLanguage),
		TabSize:         intOrDefault(m, "tabSize", DefaultTabSize),
		TabExpand:       boolOrDefault(m, "tabExpand", DefaultTabExpand),
		ShowTabs:        boolOrDefault(m, "showTabs", DefaultShowTabs),
		ShowSpaces:      boolOrDefault(m, "showSpaces", DefaultShowSpaces),
		ShowUnicode:     boolOrDefault(m, "showUnicode", DefaultShowUnicode),
		AutoIndent:      boolOrDefault(m, "autoIndent", DefaultAutoIndent),
		ShowLineNumbers: boolOrDefault(m, "showLineNumbers", DefaultShowLineNumbers),
		LineNumberMode:  stringOrDefault(m, "lineNumberMode", string(DefaultLineNumberMode)),
		LineWrap:        stringOrDefault(m, "lineWrap", DefaultLineWrap),
		MenuCommands:    menuCommandsFromSlice(sliceOrNil(m, "menuCommands")),
		HidePatterns:    stringSliceOrNil(m, "hidePatterns"),
		HideDirectories: stringSliceOrNil(m, "hideDirectories"), // Deprecated by HidePatterns
		Styles:          stylesFromMap(mapOrNil(m, "styles")),
	}
}

// Validate checks that the values in the configuration are valid.
func (c Config) Validate() error {
	if c.TabSize < 1 {
		return errors.New("TabSize must be greater than zero")
	}

	if c.LineWrap != LineWrapCharacter && c.LineWrap != LineWrapWord {
		return fmt.Errorf("LineWrap must be either %q or %q", LineWrapCharacter, LineWrapWord)
	}

	lnm := LineNumberMode(c.LineNumberMode)
	if lnm != LineNumberModeAbsolute && lnm != LineNumberModeRelative {
		return fmt.Errorf("LineNumberMode must be either %q or %q", LineNumberModeAbsolute, LineNumberModeRelative)
	}

	for _, cmd := range c.MenuCommands {
		if cmd.Name == "" {
			return fmt.Errorf("Menu name cannot be empty")
		}

		if cmd.ShellCmd == "" {
			return fmt.Errorf("Menu command %q shellCmd cannot be empty", cmd.Name)
		}

		if cmd.Mode != CmdModeSilent && cmd.Mode != CmdModeTerminal && cmd.Mode != CmdModeInsert && cmd.Mode != CmdModeInsertChoice && cmd.Mode != CmdModeFileLocations && cmd.Mode != CmdModeWorkingDir {
			return fmt.Errorf(
				"Menu command %q must have mode set to either %q, %q, %q, %q, %q, or %q",
				cmd.Name,
				CmdModeSilent,
				CmdModeTerminal,
				CmdModeInsert,
				CmdModeInsertChoice,
				CmdModeFileLocations,
				CmdModeWorkingDir,
			)
		}
	}

	return nil
}

func (c Config) HidePatternsAndHideDirectories() []string {
	result := make([]string, 0, len(c.HidePatterns)+len(c.HideDirectories))
	result = append(result, c.HidePatterns...)
	result = append(result, c.HideDirectories...)
	return result
}

func stringOrDefault(m map[string]any, key string, defaultVal string) string {
	v, ok := m[key]
	if !ok {
		return defaultVal
	}

	s, ok := v.(string)
	if !ok {
		log.Printf("Could not decode string for config key %q\n", key)
		return defaultVal
	}

	return s
}

func intOrDefault(m map[string]any, key string, defaultVal int) int {
	v, ok := m[key]
	if !ok {
		return defaultVal
	}

	switch v := v.(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		log.Printf("Could not decode int for config key %q\n", key)
		return defaultVal
	}
}

func boolOrDefault(m map[string]any, key string, defaultVal bool) bool {
	v, ok := m[key]
	if !ok {
		return defaultVal
	}

	b, ok := v.(bool)
	if !ok {
		log.Printf("Could not decode bool for config key %q\n", key)
		return defaultVal
	}

	return b
}

func sliceOrNil(m map[string]any, key string) []any {
	v, ok := m[key]
	if !ok {
		return nil
	}

	s, ok := v.([]any)
	if !ok {
		log.Printf("Could not decode slice for config key %q\n", key)
		return nil
	}

	return s
}

func stringSliceOrNil(m map[string]any, key string) []string {
	slice := sliceOrNil(m, key)
	if slice == nil {
		return nil
	}

	stringSlice := make([]string, 0, len(slice))
	for i := 0; i < len(slice); i++ {
		s, ok := (slice[i]).(string)
		if !ok {
			log.Printf("Could not decode string in slice for config key %q\n", key)
			continue
		}
		stringSlice = append(stringSlice, s)
	}
	return stringSlice
}

func mapOrNil(m map[string]any, key string) map[string]any {
	v, ok := m[key]
	if !ok {
		return nil
	}

	subMap, ok := v.(map[string]any)
	if !ok {
		log.Printf("Could not decode map for config key %q\n", key)
		return nil
	}

	return subMap
}

func menuCommandsFromSlice(s []any) []MenuCommandConfig {
	result := make([]MenuCommandConfig, 0, len(s))
	for _, m := range s {
		menuMap, ok := m.(map[string]any)
		if !ok {
			log.Printf("Could not decode menu command map from %v\n", m)
			continue
		}

		result = append(result, MenuCommandConfig{
			Name:     stringOrDefault(menuMap, "name", "[nil]"),
			ShellCmd: stringOrDefault(menuMap, "shellCmd", ""),
			Mode:     stringOrDefault(menuMap, "mode", CmdModeTerminal),
			Save:     boolOrDefault(menuMap, "save", false),
		})
	}
	return result
}

func stylesFromMap(m map[string]any) map[string]StyleConfig {
	result := make(map[string]StyleConfig, len(m))
	for k, v := range m {
		styleMap, ok := v.(map[string]any)
		if !ok {
			log.Printf("Could not decode style map from %v\n", v)
			continue
		}

		result[k] = StyleConfig{
			Color:           stringOrDefault(styleMap, "color", ""),
			BackgroundColor: stringOrDefault(styleMap, "backgroundColor", ""),
			Bold:            boolOrDefault(styleMap, "bold", false),
			Italic:          boolOrDefault(styleMap, "italic", false),
			Underline:       boolOrDefault(styleMap, "underline", false),
			StrikeThrough:   boolOrDefault(styleMap, "strikethrough", false),
		}
	}
	return result
}
