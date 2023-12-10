package text

// RuneStack represents a string with efficient operations to push/pop runes.
// The zero value is equivalent to an empty string.
type RuneStack struct {
	runes  []rune
	dirty  bool
	cached string
}

func (rs *RuneStack) Push(r rune) {
	rs.runes = append(rs.runes, r)
	rs.dirty = true
}

func (rs *RuneStack) Pop() (bool, rune) {
	if len(rs.runes) == 0 {
		return false, '\x00'
	}

	lastRune := rs.runes[len(rs.runes)-1]
	rs.runes = rs.runes[0 : len(rs.runes)-1]
	rs.dirty = true
	return true, lastRune
}

func (rs *RuneStack) Len() int {
	return len(rs.runes)
}

func (rs *RuneStack) String() string {
	if rs.dirty {
		rs.cached = string(rs.runes)
	}
	return rs.cached
}
