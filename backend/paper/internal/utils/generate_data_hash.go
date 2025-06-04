package utils

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"pariksha/common/pkg/models"
)

// generateDataHash creates a SHA256 hash of test case content
func GenerateDataHash(content models.TestCaseContent, hidden bool) string {
	data, _ := json.Marshal(struct {
		Content models.TestCaseContent `json:"content"`
		Hidden  bool                   `json:"hidden"`
	}{
		Content: content,
		Hidden:  hidden,
	})
	return fmt.Sprintf("%x", sha256.Sum256(data))
}
