package output

import (
	"encoding/json"
	"io"

	"github.com/stream-ai/chunk/internal/model"
)

// JSONFormatter implements the Formatter interface for JSON output
type JSONFormatter struct {
	// Pretty determines if the output should be pretty-printed
	Pretty bool
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter(pretty bool) *JSONFormatter {
	return &JSONFormatter{
		Pretty: pretty,
	}
}

// Format writes the chunks to the given writer in JSON format
func (f *JSONFormatter) Format(w io.Writer, result model.ChunkResult) error {
	encoder := json.NewEncoder(w)
	if f.Pretty {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(result)
}
