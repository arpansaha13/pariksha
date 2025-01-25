package utils

import (
	"crypto/rand"
	"encoding/base64"
)

// Generates a random base64 string of the given length
func GenerateBase64(length int) (string, error) {
	// Calculate the number of bytes needed
	byteLength := (length * 3) / 4

	// Create a byte slice of the calculated length
	randomBytes := make([]byte, byteLength)

	// Read random bytes into the slice
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Encode the bytes to a base64 string
	base64String := base64.StdEncoding.EncodeToString(randomBytes)

	// Return the string truncated to the requested length
	return base64String[:length], nil
}
