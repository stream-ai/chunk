package chunker

import (
	"bytes"
	"path/filepath"
	"regexp"

	"github.com/stream-ai/chunk/internal/model"
	"github.com/stream-ai/chunk/pkg/util"
)

// GenericChunker implements a fallback line-based chunker
type GenericChunker struct{}

// NewGenericChunker creates a new generic chunker
func NewGenericChunker() *GenericChunker {
	return &GenericChunker{}
}

// Language returns the language this chunker supports
func (c *GenericChunker) Language() string {
	return "generic"
}

// CanHandle checks if this chunker can handle the given file
func (c *GenericChunker) CanHandle(filePath string, language string, framework string) bool {
	// Generic chunker is a fallback for all files
	return true
}

// Chunk splits the file content into chunks based on line count
func (c *GenericChunker) Chunk(filePath string, content []byte, symbolTable *model.SymbolTable, options ChunkingOptions) ([]model.Chunk, error) {
	lines := bytes.Split(content, []byte("\n"))
	var chunks []model.Chunk

	// Determine language from the file path (best effort)
	language := filepath.Ext(filePath)
	if language != "" && language[0] == '.' {
		language = language[1:] // Remove the leading dot
	}

	// Simple line-based chunking
	for i := 0; i < len(lines); i += options.MaxChunkSize {
		end := i + options.MaxChunkSize
		if end > len(lines) {
			end = len(lines)
		}

		chunkLines := lines[i:end]
		chunkContent := string(bytes.Join(chunkLines, []byte("\n")))

		// Generate ID
		chunkID := util.GenerateID(filePath, chunkContent)

		// Extract symbols (basic approach for generic files)
		symbols := extractGenericSymbols(chunkContent)

		chunk := model.Chunk{
			ID:         chunkID,
			FilePath:   filePath,
			StartLine:  i + 1,
			EndLine:    end,
			Content:    chunkContent,
			Language:   language,
			Symbols:    symbols,
			TokenCount: util.EstimateTokenCount(chunkContent),
		}

		// Add symbols to symbol table
		for _, symbol := range symbols {
			symbolTable.AddDefinition(symbol, model.SymbolDefinition{
				Name:      symbol,
				ChunkID:   chunkID,
				FilePath:  filePath,
				StartLine: i + 1,
				EndLine:   end,
				Type:      "generic",
			})
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// extractGenericSymbols attempts to extract symbols from generic content
// This is a very simplified approach for fallback chunking
func extractGenericSymbols(content string) []string {
	// Look for common patterns that might indicate symbol definitions
	// This is a very basic implementation

	symbolsMap := make(map[string]bool)

	// Look for function-like patterns across languages
	funcRegex := regexp.MustCompile(`(function|func|def|method|void|int|string|bool|class|struct|interface)\s+(\w+)`)
	matches := funcRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 2 {
			symbolsMap[match[2]] = true
		}
	}

	// Convert map to slice
	symbols := make([]string, 0, len(symbolsMap))
	for symbol := range symbolsMap {
		symbols = append(symbols, symbol)
	}

	return symbols
}
