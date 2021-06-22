package display

import (
	"github.com/gdamore/tcell/v2"

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
	tokenKeyStyle           tcell.Style
	tokenCommentStyle       tcell.Style
}

func NewPalette() *Palette {
	s := tcell.StyleDefault
	return &Palette{
		lineNumStyle:            s.Foreground(tcell.ColorOrange),
		selectionStyle:          s.Reverse(true).Dim(true),
		searchMatchStyle:        s.Reverse(true),
		statusMsgSuccessStyle:   s.Foreground(tcell.ColorGreen).Bold(true),
		statusMsgErrorStyle:     s.Background(tcell.ColorRed).Foreground(tcell.ColorWhite).Bold(true),
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
		tokenOperatorStyle:      s.Foreground(tcell.ColorFuchsia),
		tokenKeywordStyle:       s.Foreground(tcell.ColorOrange),
		tokenNumberStyle:        s.Foreground(tcell.ColorGreen),
		tokenStringStyle:        s.Foreground(tcell.ColorRed),
		tokenKeyStyle:           s.Foreground(tcell.ColorTeal),
		tokenCommentStyle:       s.Foreground(tcell.ColorBlue),
	}
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
	case parser.TokenRoleKey:
		return p.tokenKeyStyle
	case parser.TokenRoleComment, parser.TokenRoleCommentDelimiter:
		return p.tokenCommentStyle
	default:
		return tcell.StyleDefault
	}
}
