package utf8

// Lookup table for UTF-8 character byte counts.  Set to the byte count of the character for start bytes, zero otherwise.
var CharWidth [256]byte

func init() {
	for b := 0; b < 256; b++ {
		if b>>7 == 0 {
			CharWidth[b] = 1
		} else if b>>5 == 0b110 {
			CharWidth[b] = 2
		} else if b>>4 == 0b1110 {
			CharWidth[b] = 3
		} else if b>>3 == 0b11110 {
			CharWidth[b] = 4
		}
	}
}

// Lookup table for UTF-8 start bytes. Set to 1 for start bytes, zero otherwise.
var StartByteIndicator [256]byte

func init() {
	for b := 0; b < 256; b++ {
		if b>>7 == 0 ||
			b>>5 == 0b110 ||
			b>>4 == 0b1110 ||
			b>>3 == 0b11110 {
			StartByteIndicator[b] = 1
		}
	}
}
