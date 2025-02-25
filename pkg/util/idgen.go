// pkg/util/idgen.go
package util

import (
	"crypto/sha256"
	"encoding/hex"
)

// GenerateID creates a stable ID based on input strings
func GenerateID(inputs ...string) string {
	hasher := sha256.New()
	for _, input := range inputs {
		hasher.Write([]byte(input))
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

// EstimateTokenCount provides a rough estimate of tokens in the content
func EstimateTokenCount(content string) int {
	// A very rough estimate - about 1 token per 4 characters for code
	return len(content) / 4
}
