package output

import (
	"io"

	"github.com/stream-ai/chunk/internal/model"
)

// Formatter defines the contract for output formatting
type Formatter interface {
	// Format writes the chunks to the given writer
	Format(w io.Writer, result model.ChunkResult) error
}

// FormatterRegistry stores all available formatters
type FormatterRegistry struct {
	formatters map[string]Formatter
}

// NewFormatterRegistry creates a new formatter registry
func NewFormatterRegistry() *FormatterRegistry {
	return &FormatterRegistry{
		formatters: make(map[string]Formatter),
	}
}

// Register adds a formatter to the registry
func (fr *FormatterRegistry) Register(name string, formatter Formatter) {
	fr.formatters[name] = formatter
}

// Get retrieves a formatter by name
func (fr *FormatterRegistry) Get(name string) (Formatter, bool) {
	formatter, exists := fr.formatters[name]
	return formatter, exists
}
