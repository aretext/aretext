package utf8

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateBytes(t *testing.T) {
	testCases := []struct {
		name        string
		bytes       []byte
		expectValid bool
	}{
		{"empty", nil, true},
		{"ascii", []byte("abcd1234"), true},
		{"multi-byte", []byte("丂丄丅丆丏 ¢ह€한"), true},
		{"invalid start byte", []byte{0xFF}, false},
		{"too many continuation chars", []byte{0b11100000, 0b10000000, 0b00000000}, false},
		{"missing continuation chars", []byte{0b11100000, 0b10000000, 0b00000000}, false},
		{"missing continuation chars at end", []byte{0b11110000, 0b10000000}, false},
		{"overlong sequence", []byte{0b11110111, 0b10111111, 0b10111111, 0b10111111}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := NewValidator()
			valid := v.ValidateBytes(tc.bytes) && v.ValidateEnd()
			assert.Equal(t, tc.expectValid, valid)
		})
	}
}

func TestValidateBytesIndividually(t *testing.T) {
	v := NewValidator()

	s := []byte("ვეპხის ტყაოსანი შოთა რუსთაველი ღმერთსი შემვედრე, ნუთუ კვლა დამხსნას სოფლისა შრომასა, ცეცხლს, წყალსა და მიწასა, ჰაერთა თანა მრომასა; მომცნეს ფრთენი და აღვფრინდე, მივჰხვდე მას ჩემსა ნდომასა, დღისით და ღამით ვჰხედვიდე მზისა ელვათა კრთომაასა.")

	for _, b := range s {
		assert.True(t, v.ValidateBytes([]byte{b}))
	}
	assert.True(t, v.ValidateEnd())
}

func TestValidateFile(t *testing.T) {
	v := NewValidator()
	data, err := os.ReadFile("testdata/utf8.txt")
	require.NoError(t, err)
	assert.True(t, v.ValidateBytes(data))
}

func BenchmarkValidateAscii(b *testing.B) {
	s := make([]byte, 4096)
	for i := 0; i < len(s); i++ {
		s[i] = 'a'
	}

	var valid bool
	for n := 0; n < b.N; n++ {
		v := NewValidator()
		valid = v.ValidateBytes(s) && v.ValidateEnd()
	}
	b.Log(valid)
}
