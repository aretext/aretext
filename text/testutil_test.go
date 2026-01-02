package text

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

// Repeat creates a string with the same character repeated n times.
func Repeat(c rune, n int) string {
	runes := make([]rune, n)
	for i := 0; i < n; i++ {
		runes[i] = c
	}
	return string(runes)
}
