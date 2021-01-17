package parser

import (
	"io"
	"math"
)

// InputReader provides input text for the parser.
type InputReader interface {
	io.Reader

	// SeekBackward moves the reader position backward by offset bytes.
	SeekBackward(offset uint64) error
}

// ReadSeekerInput wraps an io.ReadSeeker.
type ReadSeekerInput struct {
	R io.ReadSeeker
}

func (r *ReadSeekerInput) Read(b []byte) (int, error) {
	return r.R.Read(b)
}

func (r *ReadSeekerInput) SeekBackward(offset uint64) error {
	for offset > 0 {
		i := int64(offset)
		if offset > uint64(math.MaxInt64) {
			i = math.MaxInt64
		}

		if _, err := r.R.Seek(-1*i, io.SeekCurrent); err != nil {
			return err
		}

		offset -= uint64(i)
	}

	return nil
}
