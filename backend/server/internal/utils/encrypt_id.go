package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"pariksha/server/internal/config/env"
	"strings"
)

var (
	encryptionKey []byte
	initVector    []byte
)

func init() {
	encryptionKey = []byte(env.ID_ENCRYPTION_KEY)

	// Use a fixed IV for consistent encryption/decryption
	initVector = make([]byte, aes.BlockSize)
	copy(initVector, []byte("ParikshaPlatform")) // 16 bytes IV
}

// EncryptID encrypts an int64 ID to a URL-safe string
func EncryptID(id int64) (string, error) {
	// Convert int64 to bytes
	plaintext := make([]byte, 8)
	binary.BigEndian.PutUint64(plaintext, uint64(id))

	// Create cipher block
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	// Pad the plaintext to match block size
	padding := aes.BlockSize - (len(plaintext) % aes.BlockSize)
	padtext := make([]byte, len(plaintext)+padding)
	copy(padtext, plaintext)
	for i := len(plaintext); i < len(padtext); i++ {
		padtext[i] = byte(padding)
	}

	// Encrypt
	ciphertext := make([]byte, len(padtext))
	mode := cipher.NewCBCEncrypter(block, initVector)
	mode.CryptBlocks(ciphertext, padtext)

	// Convert to base64 and remove padding
	encoded := base64.URLEncoding.EncodeToString(ciphertext)
	return strings.TrimRight(encoded, "="), nil
}

// DecryptID decrypts an encrypted string back to int64 ID
func DecryptID(encrypted string) (int64, error) {
	// Add back base64 padding
	padding := len(encrypted) % 4
	if padding > 0 {
		encrypted += strings.Repeat("=", 4-padding)
	}

	// Decode base64
	ciphertext, err := base64.URLEncoding.DecodeString(encrypted)
	if err != nil {
		return 0, fmt.Errorf("failed to decode base64: %v", err)
	}

	// Create cipher block
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return 0, fmt.Errorf("failed to create cipher: %v", err)
	}

	// Decrypt
	mode := cipher.NewCBCDecrypter(block, initVector)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove padding
	padding = int(plaintext[len(plaintext)-1])
	plaintext = plaintext[:len(plaintext)-padding]

	// Convert bytes back to int64
	if len(plaintext) != 8 {
		return 0, fmt.Errorf("invalid decrypted length")
	}

	return int64(binary.BigEndian.Uint64(plaintext)), nil
}
