package output

import (
	"encoding/json"
	"io"

	"github.com/stream-ai/chunk/internal/model"
)

// JSONLinesFormatter implements the Formatter interface for JSON Lines output
type JSONLinesFormatter struct{}

// NewJSONLinesFormatter creates a new JSON Lines formatter
func NewJSONLinesFormatter() *JSONLinesFormatter {
	return &JSONLinesFormatter{}
}

// Format writes the chunks to the given writer in JSON Lines format
func (f *JSONLinesFormatter) Format(w io.Writer, result model.ChunkResult) error {
	for _, chunk := range result.Chunks {
		chunkBytes, err := json.Marshal(chunk)
		if err != nil {
			return err
		}
		if _, err := w.Write(chunkBytes); err != nil {
			return err
		}
		if _, err := w.Write([]byte("\n")); err != nil {
			return err
		}
	}
	return nil
}
