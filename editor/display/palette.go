package display

import (
	"log"

	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/editor/config"
	"github.com/aretext/aretext/editor/state"
	"github.com/aretext/aretext/editor/syntax/parser"
)

// Palette controls the style of displayed text.
type Palette struct {
	lineNumStyle              tcell.Style
	selectionStyle            tcell.Style
	searchMatchStyle          tcell.Style
	searchCursorStyle         tcell.Style
	statusMsgSuccessStyle     tcell.Style
	statusMsgErrorStyle       tcell.Style
	statusInputModeStyle      tcell.Style
	statusInputBufferStyle    tcell.Style
	statusRecordingMacroStyle tcell.Style
	statusFilePathStyle       tcell.Style
	menuBorderStyle           tcell.Style
	menuIconStyle             tcell.Style
	menuPromptStyle           tcell.Style
	menuQueryStyle            tcell.Style
	menuCursorStyle           tcell.Style
	menuItemSelectedStyle     tcell.Style
	menuItemUnselectedStyle   tcell.Style
	textFieldPromptStyle      tcell.Style
	textFieldInputTextStyle   tcell.Style
	textFieldBorderStyle      tcell.Style
	searchPrefixStyle         tcell.Style
	searchQueryStyle          tcell.Style
	tokenRoleStyle            map[parser.TokenRole]tcell.Style
}

func NewPalette() *Palette {
	s := tcell.StyleDefault
	return &Palette{
		lineNumStyle:              s.Foreground(tcell.ColorOlive),
		selectionStyle:            s.Reverse(true).Dim(true),
		searchMatchStyle:          s.Reverse(true),
		searchCursorStyle:         s.Reverse(true).Dim(true),
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
			parser.TokenRoleCustom1:  s.Foreground(tcell.ColorTeal),
			parser.TokenRoleCustom2:  s.Foreground(tcell.ColorDarkBlue),
			parser.TokenRoleCustom3:  s.Foreground(tcell.ColorRed),
			parser.TokenRoleCustom4:  s.Foreground(tcell.ColorLime),
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
}

func NewPaletteFromConfigStyles(styles map[string]config.StyleConfig) *Palette {
	p := NewPalette()
	for k, v := range styles {
		s := styleFromConfig(v)
		switch k {
		case config.StyleLineNum:
			p.lineNumStyle = s
		case config.StyleTokenOperator:
			p.tokenRoleStyle[parser.TokenRoleOperator] = s
		case config.StyleTokenKeyword:
			p.tokenRoleStyle[parser.TokenRoleKeyword] = s
		case config.StyleTokenNumber:
			p.tokenRoleStyle[parser.TokenRoleNumber] = s
		case config.StyleTokenString:
			p.tokenRoleStyle[parser.TokenRoleString] = s
		case config.StyleTokenComment:
			p.tokenRoleStyle[parser.TokenRoleComment] = s
		case config.StyleTokenCustom1:
			p.tokenRoleStyle[parser.TokenRoleCustom1] = s
		case config.StyleTokenCustom2:
			p.tokenRoleStyle[parser.TokenRoleCustom2] = s
		case config.StyleTokenCustom3:
			p.tokenRoleStyle[parser.TokenRoleCustom3] = s
		case config.StyleTokenCustom4:
			p.tokenRoleStyle[parser.TokenRoleCustom4] = s
		case config.StyleTokenCustom5:
			p.tokenRoleStyle[parser.TokenRoleCustom5] = s
		case config.StyleTokenCustom6:
			p.tokenRoleStyle[parser.TokenRoleCustom6] = s
		case config.StyleTokenCustom7:
			p.tokenRoleStyle[parser.TokenRoleCustom7] = s
		case config.StyleTokenCustom8:
			p.tokenRoleStyle[parser.TokenRoleCustom8] = s
		case config.StyleTokenCustom9:
			p.tokenRoleStyle[parser.TokenRoleCustom9] = s
		case config.StyleTokenCustom10:
			p.tokenRoleStyle[parser.TokenRoleCustom10] = s
		case config.StyleTokenCustom11:
			p.tokenRoleStyle[parser.TokenRoleCustom11] = s
		case config.StyleTokenCustom12:
			p.tokenRoleStyle[parser.TokenRoleCustom12] = s
		case config.StyleTokenCustom13:
			p.tokenRoleStyle[parser.TokenRoleCustom13] = s
		case config.StyleTokenCustom14:
			p.tokenRoleStyle[parser.TokenRoleCustom14] = s
		case config.StyleTokenCustom15:
			p.tokenRoleStyle[parser.TokenRoleCustom15] = s
		case config.StyleTokenCustom16:
			p.tokenRoleStyle[parser.TokenRoleCustom16] = s
		default:
			log.Printf("Unrecognized style key: %s\n", k)
		}
	}
	return p
}

func (p *Palette) StyleForLineNum() tcell.Style {
	return p.lineNumStyle
}

func (p *Palette) StyleForSelection() tcell.Style {
	return p.selectionStyle
}

func (p *Palette) StyleForSearchMatch() tcell.Style {
	return p.searchMatchStyle
}

func (p *Palette) StyleForSearchCursor() tcell.Style {
	return p.searchCursorStyle
}

func (p *Palette) StyleForStatusInputMode() tcell.Style {
	return p.statusInputModeStyle
}

func (p *Palette) StyleForStatusInputBuffer() tcell.Style {
	return p.statusInputBufferStyle
}

func (p *Palette) StyleForStatusRecordingMacro() tcell.Style {
	return p.statusRecordingMacroStyle
}

func (p *Palette) StyleForStatusFilePath() tcell.Style {
	return p.statusFilePathStyle
}

func (p *Palette) StyleForStatusMsg(statusMsgStyle state.StatusMsgStyle) tcell.Style {
	switch statusMsgStyle {
	case state.StatusMsgStyleSuccess:
		return p.statusMsgSuccessStyle
	case state.StatusMsgStyleError:
		return p.statusMsgErrorStyle
	default:
		return tcell.StyleDefault
	}
}

func (p *Palette) StyleForMenuBorder() tcell.Style {
	return p.menuBorderStyle
}

func (p *Palette) StyleForMenuIcon() tcell.Style {
	return p.menuIconStyle
}

func (p *Palette) StyleForMenuPrompt() tcell.Style {
	return p.menuPromptStyle
}

func (p *Palette) StyleForMenuQuery() tcell.Style {
	return p.menuQueryStyle
}

func (p *Palette) StyleForMenuCursor() tcell.Style {
	return p.menuCursorStyle
}

func (p *Palette) StyleForMenuItem(selected bool) tcell.Style {
	if selected {
		return p.menuItemSelectedStyle
	} else {
		return p.menuItemUnselectedStyle
	}
}

func (p *Palette) StyleForTextFieldPrompt() tcell.Style {
	return p.textFieldPromptStyle
}

func (p *Palette) StyleForTextFieldInputText() tcell.Style {
	return p.textFieldInputTextStyle
}

func (p *Palette) StyleForTextFieldBorder() tcell.Style {
	return p.textFieldBorderStyle
}

func (p *Palette) StyleForSearchPrefix() tcell.Style {
	return p.searchPrefixStyle
}

func (p *Palette) StyleForSearchQuery() tcell.Style {
	return p.searchQueryStyle
}

func (p *Palette) StyleForTokenRole(tokenRole parser.TokenRole) tcell.Style {
	// If key is not set, returns tcell.StyleDefault (the zero value).
	return p.tokenRoleStyle[tokenRole]
}

func styleFromConfig(s config.StyleConfig) tcell.Style {
	c := tcell.GetColor(s.Color)
	style := tcell.StyleDefault.Foreground(c)

	if s.BackgroundColor != "" {
		bg := tcell.GetColor(s.BackgroundColor)
		style = style.Background(bg)
	}

	if s.Bold {
		style = style.Bold(true)
	}

	if s.Italic {
		style = style.Italic(true)
	}

	if s.Underline {
		style = style.Underline(true)
	}

	if s.StrikeThrough {
		style = style.StrikeThrough(true)
	}

	return style
}
