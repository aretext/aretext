package display

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/config"
	"github.com/aretext/aretext/syntax/parser"
)

func TestPaletteFromConfigStyles(t *testing.T) {
	configStyles := map[string]config.StyleConfig{
		config.StyleEscapedUnicode: {
			Color: "blue",
			Bold:  true,
		},
		config.StyleTokenCustom1: {
			Color: "black",
			Bold:  true,
		},
		config.StyleTokenCustom2: {
			Color:     "red",
			Italic:    true,
			Underline: true,
		},
		config.StyleTokenCustom3: {
			Color:         "green",
			StrikeThrough: true,
		},
		config.StyleTokenCustom4: {
			BackgroundColor: "yellow",
		},
	}

	palette := NewPaletteFromConfigStyles(configStyles)

	s := tcell.StyleDefault
	expected := &Palette{
		lineNumStyle:              s.Foreground(color.Olive),
		selectionStyle:            s.Reverse(true).Dim(true),
		searchCursorStyle:         s.Reverse(true).Dim(true),
		searchMatchStyle:          s.Reverse(true),
		statusMsgSuccessStyle:     s.Foreground(color.Green).Bold(true),
		statusMsgErrorStyle:       s.Background(color.Maroon).Foreground(color.White).Bold(true),
		statusInputModeStyle:      s.Bold(true),
		statusInputBufferStyle:    s.Bold(true),
		statusRecordingMacroStyle: s.Bold(true),
		statusFilePathStyle:       s.Bold(true),
		menuBorderStyle:           s.Dim(true),
		menuIconStyle:             s,
		menuPromptStyle:           s.Dim(true),
		menuQueryStyle:            s,
		menuCursorStyle:           s.Bold(true),
		menuItemSelectedStyle:     s.Underline(true),
		menuItemUnselectedStyle:   s,
		textFieldPromptStyle:      s.Dim(true),
		textFieldInputTextStyle:   s,
		textFieldBorderStyle:      s,
		searchPrefixStyle:         s,
		searchQueryStyle:          s,
		escapedUnicodeStyle:       s.Foreground(color.Blue).Bold(true),
		tokenRoleStyle: map[parser.TokenRole]tcell.Style{
			parser.TokenRoleOperator: s.Foreground(color.Purple),
			parser.TokenRoleKeyword:  s.Foreground(color.Olive),
			parser.TokenRoleNumber:   s.Foreground(color.Green),
			parser.TokenRoleString:   s.Foreground(color.Maroon),
			parser.TokenRoleComment:  s.Foreground(color.Navy),
			parser.TokenRoleCustom1:  s.Foreground(color.Black).Bold(true),
			parser.TokenRoleCustom2:  s.Foreground(color.Red).Italic(true).Underline(true),
			parser.TokenRoleCustom3:  s.Foreground(color.Green).StrikeThrough(true),
			parser.TokenRoleCustom4:  s.Background(color.Yellow),
			parser.TokenRoleCustom5:  s.Foreground(color.Fuchsia),
			parser.TokenRoleCustom6:  s.Foreground(color.Aqua),
			parser.TokenRoleCustom7:  s.Foreground(color.DarkGreen),
			parser.TokenRoleCustom8:  s.Foreground(color.DarkCyan),
			parser.TokenRoleCustom9:  s.Foreground(color.Teal),
			parser.TokenRoleCustom10: s.Foreground(color.DarkBlue),
			parser.TokenRoleCustom11: s.Foreground(color.Red),
			parser.TokenRoleCustom12: s.Foreground(color.Lime),
			parser.TokenRoleCustom13: s.Foreground(color.Fuchsia),
			parser.TokenRoleCustom14: s.Foreground(color.Aqua),
			parser.TokenRoleCustom15: s.Foreground(color.DarkGreen),
			parser.TokenRoleCustom16: s.Foreground(color.DarkCyan),
		},
	}

	assert.Equal(t, expected, palette)
}
