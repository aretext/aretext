package state

// StatusMsgStyle controls how a status message will be displayed.
type StatusMsgStyle int

const (
	StatusMsgStyleSuccess = StatusMsgStyle(iota)
	StatusMsgStyleError
)

func (s StatusMsgStyle) String() string {
	switch s {
	case StatusMsgStyleSuccess:
		return "success"
	case StatusMsgStyleError:
		return "error"
	default:
		panic("invalid style")
	}
}

// StatusMsg is a message displayed in the status bar.
type StatusMsg struct {
	Style StatusMsgStyle
	Text  string
}

// SetStatusMsg sets the message displayed in the status bar.
func SetStatusMsg(state *EditorState, statusMsg StatusMsg) {
	state.statusMsg = statusMsg
}
