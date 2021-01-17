package text

import "io"

// SingleByteReader is a CloneableReader that produces a single byte at a time.
type SingleByteReader struct {
	s string
	i int
}

func NewSingleByteReader(s string) CloneableReader {
	return &SingleByteReader{s, 0}
}

func (r *SingleByteReader) Read(p []byte) (n int, err error) {
	n = copy(p, r.s[r.i:r.i+1])
	r.i++
	if r.i >= len(r.s) {
		err = io.EOF
	}
	return
}

func (r *SingleByteReader) Clone() CloneableReader {
	return &SingleByteReader{
		s: r.s,
		i: r.i,
	}
}

// Reverse reverses the bytes of a string.
// The result may not be a valid UTF-8.
func Reverse(s string) string {
	bytes := []byte(s)
	reversedBytes := make([]byte, len(bytes))
	for i := 0; i < len(reversedBytes); i++ {
		reversedBytes[i] = bytes[len(bytes)-1-i]
	}
	return string(reversedBytes)
}
