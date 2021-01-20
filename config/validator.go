package config

func StringLenGreaterThan(s string, minLen int) bool {
	return len(s) > minLen
}

func IntGreaterThan(i int, minVal int) bool {
	return i > minVal
}
