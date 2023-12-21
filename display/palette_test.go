package display

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/config"
	"github.com/aretext/aretext/syntax/parser"
)

func TestPaletteFromConfigStyles(t *testing.T) {
	configStyles := map[string]config.StyleConfig{
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
		lineNumStyle:              s.Foreground(tcell.ColorOlive),
		selectionStyle:            s.Reverse(true).Dim(true),
		searchCursorStyle:         s.Reverse(true).Dim(true),
		searchMatchStyle:          s.Reverse(true),
		statusMsgSuccessStyle:     s.Foreground(tcell.ColorGreen).Bold(true),
		statusMsgErrorStyle:       s.Background(tcell.ColorMaroon).Foreground(tcell.ColorWhite).Bold(true),
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
		tokenRoleStyle: map[parser.TokenRole]tcell.Style{
			parser.TokenRoleOperator: s.Foreground(tcell.ColorPurple),
			parser.TokenRoleKeyword:  s.Foreground(tcell.ColorOlive),
			parser.TokenRoleNumber:   s.Foreground(tcell.ColorGreen),
			parser.TokenRoleString:   s.Foreground(tcell.ColorMaroon),
			parser.TokenRoleComment:  s.Foreground(tcell.ColorNavy),
			parser.TokenRoleCustom1:  s.Foreground(tcell.ColorBlack).Bold(true),
			parser.TokenRoleCustom2:  s.Foreground(tcell.ColorRed).Italic(true).Underline(true),
			parser.TokenRoleCustom3:  s.Foreground(tcell.ColorGreen).StrikeThrough(true),
			parser.TokenRoleCustom4:  s.Background(tcell.ColorYellow),
			parser.TokenRoleCustom5:  s.Foreground(tcell.ColorFuchsia),
			parser.TokenRoleCustom6:  s.Foreground(tcell.ColorAqua),
			parser.TokenRoleCustom7:  s.Foreground(tcell.ColorDarkGreen),
			parser.TokenRoleCustom8:  s.Foreground(tcell.ColorDarkCyan),
			parser.TokenRoleCustom9:  s.Foreground(tcell.ColorTeal),
			parser.TokenRoleCustom10: s.Foreground(tcell.ColorDarkBlue),
			parser.TokenRoleCustom11: s.Foreground(tcell.ColorRed),
			parser.TokenRoleCustom12: s.Foreground(tcell.ColorLime),
			parser.TokenRoleCustom13: s.Foreground(tcell.ColorFuchsia),
			parser.TokenRoleCustom14: s.Foreground(tcell.ColorAqua),
			parser.TokenRoleCustom15: s.Foreground(tcell.ColorDarkGreen),
			parser.TokenRoleCustom16: s.Foreground(tcell.ColorDarkCyan),
		},
	}

	assert.Equal(t, expected, palette)
}
