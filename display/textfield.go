package display

import (
	"github.com/gdamore/tcell/v2"

	"github.com/aretext/aretext/state"
)

// DrawTextField draws the text field for user input at the top of the screen.
func DrawTextField(screen tcell.Screen, palette *Palette, textfield *state.TextFieldState) {
	screenWidth, screenHeight := screen.Size()
	if screenHeight == 0 || screenWidth == 0 {
		return
	}

	// Textfield prompt and input drawn in the first two rows.
	height := screenHeight
	if height > 2 {
		height = 2
	}
	sr := NewScreenRegion(screen, 0, 0, screenWidth, height)
	sr.Clear()

	// Draw the prompt in the first row.
	promptText := textfield.PromptText()
	sr.PutStrStyled(0, 0, promptText, palette.StyleForTextFieldPrompt())

	// Draw the user input on the second row, with the cursor at the end.
	col := sr.PutStrStyled(0, 1, textfield.InputText(), palette.StyleForTextFieldInputText())

	// Append autocomplete suffix (could be empty).
	col = sr.PutStrStyled(col, 1, textfield.AutocompleteSuffix(), palette.StyleForTextFieldInputText())

	// Cursor the end of user input + autocomplete suffix.
	sr.ShowCursor(col, 1)

	// Draw bottom border, unless it would overlap the status bar in last row.
	if screenHeight > 2 {
		borderRegion := NewScreenRegion(screen, 0, 2, screenWidth, 1)
		borderRegion.Fill(tcell.RuneHLine, palette.StyleForTextFieldBorder())
	}
}
