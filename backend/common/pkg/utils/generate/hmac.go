package generate

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
)

// HMACHash creates a HMAC SHA-256 hash for the given ID
func HMACHash(id int64) string {
	key := []byte(os.Getenv("HMAC_SECRET"))
	h := hmac.New(sha256.New, key)
	h.Write([]byte(fmt.Sprintf("%d", id)))
	return hex.EncodeToString(h.Sum(nil))
}
