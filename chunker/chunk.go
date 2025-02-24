package chunker

import (
	"path/filepath"
	"strings"
)

type Chunk struct {
	FilePath  string `json:"file"`
	ChunkID   string `json:"chunk_id"`
	ChunkType string `json:"chunk_type"` // e.g. "function", "struct", "method"
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Code      string `json:"code"`
}

// AutoDetectLanguage attempts to guess the language from file extension
func AutoDetectLanguage(path string) string {
	// Simplistic approach:
	// .go => "go"
	// .ts or .tsx => "react" or "ts"
	// fallback => "fallback"
	// If you want to differentiate ".ts" vs. ".tsx" for React, do so here.
	ext := getExtension(path)

	switch ext {
	case ".go":
		return "go"
	case ".ts":
		return "ts"
	case ".tsx":
		return "react"
	default:
		return "fallback"
	}
}

func getExtension(path string) string {
	return filepath.Ext(path)
}

func isLikelyBinaryByExtension(filename string) bool {
	binaryExt := map[string]bool{
		"":       true,
		".dll":   true,
		".so":    true,
		".dylib": true,
		".a":     true,
		".o":     true,
		".exe":   true,
		".bin":   true,
		".png":   true,
		".jpg":   true,
		".gif":   true,
		".pdf":   true,
		".zip":   true,
		".tar":   true,
		".gz":    true,
		".bz2":   true,
		".xz":    true,
	}
	ext := strings.ToLower(filepath.Ext(filename))
	return binaryExt[ext]
}
