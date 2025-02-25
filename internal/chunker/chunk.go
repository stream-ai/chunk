package chunker

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"strings"
)

// -----------------------------
// 1. Chunk struct
// -----------------------------

type Chunk struct {
	FilePath  string `json:"file"`
	ChunkID   string `json:"chunk_id"`
	ChunkType string `json:"chunk_type"` // e.g. "function", "method", "fallback"
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Code      string `json:"code"`
}

// -----------------------------
// 2. Shared Helpers
// -----------------------------

// IsLikelyBinaryByExtension checks a set of known binary extensions
func IsLikelyBinaryByExtension(filename string) bool {
	// This map can be extended
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

// AutoDetectLanguage returns "go", "ts", "react", or "fallback",
// based on file extension. This is simplistic; adapt as needed.
func AutoDetectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "go"
	case ".ts":
		return "ts"
	case ".tsx":
		// You might also treat .tsx as "ts" if you prefer,
		// or always treat .tsx as React
		return "react"
	default:
		return "fallback"
	}
}

// For convenience, a helper to generate a stable chunk ID
// from a chunk's file path + code content. If path/contents
// are unchanged, the ID remains the same. If either changes,
// the ID changes.
func makeChunkID(filePath, code string) string {
	combined := filePath + "\n" + code
	sum := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(sum[:])
}

// -----------------------------
// 3. Main Entry: ProcessFile
// -----------------------------

// ProcessFile is the primary entry point for chunking a single file.
//
// 1. Skips known-binary files
// 2. If forcedLang != "", uses that language
// 3. Otherwise calls AutoDetectLanguage
// 4. Dispatches to chunker for "go", "ts", "react", or line-based fallback
func ProcessFile(path string, forcedLang string, fallbackLines int) ([]Chunk, error) {
	// 1) Skip known-binary
	if IsLikelyBinaryByExtension(path) {
		// Return no chunks
		return nil, nil
	}

	// 2) Language
	lang := forcedLang
	if lang == "" {
		lang = AutoDetectLanguage(path)
	}

	// 3) Dispatch
	switch lang {
	case "go":
		return ChunkGoFile(path)
	case "ts", "react":
		return ChunkTypeScriptFile(path, lang)
	default:
		return ChunkFallback(path, fallbackLines)
	}
}
