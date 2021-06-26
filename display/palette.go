package display

import (
	"log"

	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/config"
	"github.com/aretext/aretext/state"
	"github.com/aretext/aretext/syntax/parser"
)

// Palette controls the style of displayed text.
type Palette struct {
	lineNumStyle            tcell.Style
	selectionStyle          tcell.Style
	searchMatchStyle        tcell.Style
	statusMsgSuccessStyle   tcell.Style
	statusMsgErrorStyle     tcell.Style
	statusInputModeStyle    tcell.Style
	statusInputBufferStyle  tcell.Style
	statusFilePathStyle     tcell.Style
	menuBorderStyle         tcell.Style
	menuIconStyle           tcell.Style
	menuPromptStyle         tcell.Style
	menuQueryStyle          tcell.Style
	menuCursorStyle         tcell.Style
	menuItemSelectedStyle   tcell.Style
	menuItemUnselectedStyle tcell.Style
	searchPrefixStyle       tcell.Style
	searchQueryStyle        tcell.Style
	tokenOperatorStyle      tcell.Style
	tokenKeywordStyle       tcell.Style
	tokenNumberStyle        tcell.Style
	tokenStringStyle        tcell.Style
	tokenCommentStyle       tcell.Style
	tokenCustom1Style       tcell.Style
	tokenCustom2Style       tcell.Style
	tokenCustom3Style       tcell.Style
	tokenCustom4Style       tcell.Style
	tokenCustom5Style       tcell.Style
	tokenCustom6Style       tcell.Style
	tokenCustom7Style       tcell.Style
	tokenCustom8Style       tcell.Style
}

func NewPalette() *Palette {
	s := tcell.StyleDefault
	return &Palette{
		lineNumStyle:            s.Foreground(tcell.ColorOlive),
		selectionStyle:          s.Reverse(true).Dim(true),
		searchMatchStyle:        s.Reverse(true),
		statusMsgSuccessStyle:   s.Foreground(tcell.ColorGreen).Bold(true),
		statusMsgErrorStyle:     s.Background(tcell.ColorMaroon).Foreground(tcell.ColorWhite).Bold(true),
		statusInputModeStyle:    s.Bold(true),
		statusInputBufferStyle:  s.Bold(true),
		statusFilePathStyle:     s,
		menuBorderStyle:         s.Dim(true),
		menuIconStyle:           s,
		menuPromptStyle:         s.Dim(true),
		menuQueryStyle:          s,
		menuCursorStyle:         s.Bold(true),
		menuItemSelectedStyle:   s.Underline(true),
		menuItemUnselectedStyle: s,
		searchPrefixStyle:       s,
		searchQueryStyle:        s,
		tokenOperatorStyle:      s.Foreground(tcell.ColorPurple),
		tokenKeywordStyle:       s.Foreground(tcell.ColorOlive),
		tokenNumberStyle:        s.Foreground(tcell.ColorGreen),
		tokenStringStyle:        s.Foreground(tcell.ColorMaroon),
		tokenCommentStyle:       s.Foreground(tcell.ColorNavy),
		tokenCustom1Style:       s,
		tokenCustom2Style:       s,
		tokenCustom3Style:       s,
		tokenCustom4Style:       s,
		tokenCustom5Style:       s,
		tokenCustom6Style:       s,
		tokenCustom7Style:       s,
		tokenCustom8Style:       s,
	}
}

func NewPaletteFromConfigStyles(styles map[string]config.StyleConfig) *Palette {
	p := NewPalette()
	for k, v := range styles {
		switch k {
		case config.StyleLineNum:
			p.lineNumStyle = styleFromConfig(v)
		case config.StyleTokenOperator:
			p.tokenOperatorStyle = styleFromConfig(v)
		case config.StyleTokenKeyword:
			p.tokenKeywordStyle = styleFromConfig(v)
		case config.StyleTokenNumber:
			p.tokenNumberStyle = styleFromConfig(v)
		case config.StyleTokenString:
			p.tokenStringStyle = styleFromConfig(v)
		case config.StyleTokenComment:
			p.tokenCommentStyle = styleFromConfig(v)
		case config.StyleTokenCustom1:
			p.tokenCustom1Style = styleFromConfig(v)
		case config.StyleTokenCustom2:
			p.tokenCustom2Style = styleFromConfig(v)
		case config.StyleTokenCustom3:
			p.tokenCustom3Style = styleFromConfig(v)
		case config.StyleTokenCustom4:
			p.tokenCustom4Style = styleFromConfig(v)
		case config.StyleTokenCustom5:
			p.tokenCustom5Style = styleFromConfig(v)
		case config.StyleTokenCustom6:
			p.tokenCustom6Style = styleFromConfig(v)
		case config.StyleTokenCustom7:
			p.tokenCustom7Style = styleFromConfig(v)
		case config.StyleTokenCustom8:
			p.tokenCustom8Style = styleFromConfig(v)
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

func (p *Palette) StyleForStatusInputMode() tcell.Style {
	return p.statusInputModeStyle
}

func (p *Palette) StyleForStatusInputBuffer() tcell.Style {
	return p.statusInputBufferStyle
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

func (p *Palette) StyleForSearchPrefix() tcell.Style {
	return p.searchPrefixStyle
}

func (p *Palette) StyleForSearchQuery() tcell.Style {
	return p.searchQueryStyle
}

func (p *Palette) StyleForTokenRole(tokenRole parser.TokenRole) tcell.Style {
	switch tokenRole {
	case parser.TokenRoleOperator:
		return p.tokenOperatorStyle
	case parser.TokenRoleKeyword:
		return p.tokenKeywordStyle
	case parser.TokenRoleNumber:
		return p.tokenNumberStyle
	case parser.TokenRoleString, parser.TokenRoleStringQuote:
		return p.tokenStringStyle
	case parser.TokenRoleComment, parser.TokenRoleCommentDelimiter:
		return p.tokenCommentStyle
	case parser.TokenRoleCustom1:
		return p.tokenCustom1Style
	case parser.TokenRoleCustom2:
		return p.tokenCustom2Style
	case parser.TokenRoleCustom3:
		return p.tokenCustom3Style
	case parser.TokenRoleCustom4:
		return p.tokenCustom4Style
	case parser.TokenRoleCustom5:
		return p.tokenCustom5Style
	case parser.TokenRoleCustom6:
		return p.tokenCustom6Style
	case parser.TokenRoleCustom7:
		return p.tokenCustom7Style
	case parser.TokenRoleCustom8:
		return p.tokenCustom8Style
	default:
		return tcell.StyleDefault
	}
}

func styleFromConfig(s config.StyleConfig) tcell.Style {
	c := tcell.GetColor(s.Color)
	return tcell.StyleDefault.Foreground(c)
}
