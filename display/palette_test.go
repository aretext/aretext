package display

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/config"
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
		searchPrefixStyle:         s,
		searchQueryStyle:          s,
		tokenOperatorStyle:        s.Foreground(tcell.ColorPurple),
		tokenKeywordStyle:         s.Foreground(tcell.ColorOlive),
		tokenNumberStyle:          s.Foreground(tcell.ColorGreen),
		tokenStringStyle:          s.Foreground(tcell.ColorMaroon),
		tokenCommentStyle:         s.Foreground(tcell.ColorNavy),
		tokenCustom1Style:         s.Foreground(tcell.ColorBlack).Bold(true),
		tokenCustom2Style:         s.Foreground(tcell.ColorRed).Italic(true).Underline(true),
		tokenCustom3Style:         s.Foreground(tcell.ColorGreen).StrikeThrough(true),
		tokenCustom4Style:         s.Background(tcell.ColorYellow),
		tokenCustom5Style:         s.Foreground(tcell.ColorFuchsia),
		tokenCustom6Style:         s.Foreground(tcell.ColorAqua),
		tokenCustom7Style:         s.Foreground(tcell.ColorDarkGreen),
		tokenCustom8Style:         s.Foreground(tcell.ColorDarkCyan),
	}

	assert.Equal(t, expected, palette)
}
