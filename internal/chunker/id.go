package chunker

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
)

func GenerateChunkID(projectRoot, absFilePath string, chunkContent string) (string, error) {
	// 1. relative path from project root
	relPath, err := filepath.Rel(projectRoot, absFilePath)
	if err != nil {
		return "", err
	}

	// 2. combine
	toHash := relPath + "\n" + chunkContent

	// 3. compute SHA-256
	sum := sha256.Sum256([]byte(toHash))
	// Return hex string (could return short prefix if desired)
	return hex.EncodeToString(sum[:]), nil
}
