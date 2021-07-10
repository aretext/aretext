package parser

// TODO
type ShiftDirection int

const (
	ShiftDirectionForward = ShiftDirection(iota)
	ShiftDirectionBackward
)

func (d ShiftDirection) Reverse() ShiftDirection {
	switch d {
	case ShiftDirectionForward:
		return ShiftDirectionBackward
	case ShiftDirectionBackward:
		return ShiftDirectionForward
	default:
		panic("Unrecognized shift direction")
	}
}

// TODO
type Shift struct {
	Direction ShiftDirection
	Offset    uint64
}

// TODO
func (s Shift) Add(other Shift) Shift {
	if other.Direction == s.Direction {
		return Shift{
			Direction: s.Direction,
			Offset:    s.Offset + other.Offset,
		}
	} else {
		if s.Offset >= other.Offset {
			return Shift{
				Direction: s.Direction,
				Offset:    s.Offset - other.Offset,
			}
		} else {
			return Shift{
				Direction: other.Direction,
				Offset:    other.Offset - s.Offset,
			}
		}
	}
}

// TODO
func (s Shift) Negate() Shift {
	return Shift{
		Direction: s.Direction.Reverse(),
		Offset:    s.Offset,
	}
}

// TODO
func (s Shift) Resolve(pos uint64) uint64 {
	if s.Direction == ShiftDirectionForward {
		return pos + s.Offset
	} else {
		if pos >= s.Offset {
			return pos - s.Offset
		} else {
			return 0
		}
	}
}
