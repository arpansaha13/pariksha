package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"pariksha/server/internal/config/env"
)

var (
	encryptionKey []byte
)

func init() {
	encryptionKey = []byte(env.ID_ENCRYPTION_KEY)
}

// EncryptID encrypts an int64 ID to a URL-safe string using AES-GCM
func EncryptID(id int64) (string, error) {
	// Convert int64 to bytes
	plaintext := make([]byte, 8)
	binary.BigEndian.PutUint64(plaintext, uint64(id))

	// Create cipher block
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %v", err)
	}

	// Encrypt and authenticate
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Convert to URL-safe base64
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

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return 0, fmt.Errorf("failed to create GCM: %v", err)
	}

	// Extract nonce and decrypt
	if len(ciphertext) < gcm.NonceSize() {
		return 0, fmt.Errorf("invalid encrypted-id: too short")
	}
	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	// Decrypt and verify
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return 0, fmt.Errorf("invalid encrypted-id: authentication failed")
	}

	// Convert bytes back to int64
	if len(plaintext) != 8 {
		return 0, fmt.Errorf("invalid decrypted length")
	}

	return int64(binary.BigEndian.Uint64(plaintext)), nil
}
