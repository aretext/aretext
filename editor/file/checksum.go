package file

import (
	"crypto/md5"
	"fmt"
	"hash"
)

// Checksummer calculates an MD5 checksum.
type Checksummer struct {
	h hash.Hash
}

func NewChecksummer() *Checksummer {
	return &Checksummer{
		h: md5.New(),
	}
}

// Write implements io.Writer#Write()
func (c *Checksummer) Write(p []byte) (n int, err error) {
	return c.h.Write(p)
}

// Checksum returns the MD5 checksum as a base16 string.
func (c *Checksummer) Checksum() string {
	return fmt.Sprintf("%x", c.h.Sum(nil))
}
