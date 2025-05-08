package utils

import "unicode"

// IsAlpha checks if a string contains only alphabetic characters
func IsAlpha(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}
