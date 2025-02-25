// internal/chunker/chunker.go (updated with exported ChunkingOptions)

package chunker

import (
	"github.com/stream-ai/chunk/internal/model"
)

// ChunkingOptions contains configuration for the chunking process
type ChunkingOptions struct {
	MinChunkSize int
	MaxChunkSize int
}

// Chunker interface defines the contract for code chunkers
type Chunker interface {
	// Language returns the language this chunker supports
	Language() string

	// CanHandle checks if this chunker can handle the given file
	CanHandle(filePath string, language string, framework string) bool

	// Chunk splits the file content into chunks
	Chunk(filePath string, content []byte, symbolTable *model.SymbolTable, options ChunkingOptions) ([]model.Chunk, error)
}

// ChunkerRegistry stores all available chunkers
type ChunkerRegistry struct {
	chunkers []Chunker
}

// NewChunkerRegistry creates a new chunker registry
func NewChunkerRegistry() *ChunkerRegistry {
	return &ChunkerRegistry{
		chunkers: make([]Chunker, 0),
	}
}

// Register adds a chunker to the registry
func (cr *ChunkerRegistry) Register(chunker Chunker) {
	cr.chunkers = append(cr.chunkers, chunker)
}

// FindChunker finds the appropriate chunker for the given file
func (cr *ChunkerRegistry) FindChunker(filePath string, language string, framework string) Chunker {
	for _, chunker := range cr.chunkers {
		if chunker.CanHandle(filePath, language, framework) {
			return chunker
		}
	}
	return nil // No suitable chunker found
}
