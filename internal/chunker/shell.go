package chunker

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/stream-ai/chunk/internal/model"
	"github.com/stream-ai/chunk/pkg/util"
)

// ShellChunker implements the Chunker interface for shell scripts
type ShellChunker struct{}

// NewShellChunker creates a new shell script chunker
func NewShellChunker() *ShellChunker {
	return &ShellChunker{}
}

// Language returns the language this chunker supports
func (c *ShellChunker) Language() string {
	return "shell"
}

// CanHandle checks if this chunker can handle the given file
func (c *ShellChunker) CanHandle(filePath string, language string, framework string) bool {
	return language == "shell" || language == "bash" || language == "zsh"
}

// Chunk splits shell script content into chunks
func (c *ShellChunker) Chunk(filePath string, content []byte, symbolTable *model.SymbolTable, options ChunkingOptions) ([]model.Chunk, error) {
	lines := bytes.Split(content, []byte("\n"))
	var chunks []model.Chunk

	// Find function definitions and block boundaries
	var currentChunk [][]byte
	startLine := 1
	inFunction := false

	for i, line := range lines {
		lineStr := string(line)
		trimmedLine := strings.TrimSpace(lineStr)

		// Check for function definition
		if isFunctionStart(trimmedLine) && !inFunction {
			// If we have accumulated lines, create a chunk
			if len(currentChunk) > 0 {
				chunk := c.createChunk(filePath, currentChunk, startLine, i, symbolTable)
				chunks = append(chunks, chunk)
				currentChunk = [][]byte{}
			}

			inFunction = true
			startLine = i + 1
		}

		// Check for end of function
		if inFunction && (trimmedLine == "}" || strings.HasPrefix(trimmedLine, "} #")) {
			currentChunk = append(currentChunk, line)
			chunk := c.createChunk(filePath, currentChunk, startLine, i+1, symbolTable)
			chunks = append(chunks, chunk)
			currentChunk = [][]byte{}
			inFunction = false
			startLine = i + 2
			continue
		}

		// Add line to current chunk
		currentChunk = append(currentChunk, line)

		// If we have accumulated many lines outside a function, create a chunk
		if !inFunction && len(currentChunk) >= options.MaxChunkSize {
			chunk := c.createChunk(filePath, currentChunk, startLine, i+1, symbolTable)
			chunks = append(chunks, chunk)
			currentChunk = [][]byte{}
			startLine = i + 2
		}
	}

	// Add any remaining lines
	if len(currentChunk) > 0 {
		chunk := c.createChunk(filePath, currentChunk, startLine, len(lines), symbolTable)
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// createChunk creates a chunk from lines
func (c *ShellChunker) createChunk(filePath string, lines [][]byte, startLine, endLine int, symbolTable *model.SymbolTable) model.Chunk {
	content := string(bytes.Join(lines, []byte("\n")))
	chunkID := util.GenerateID(filePath, content)

	// Extract function names as symbols
	symbols := c.extractFunctions(content)

	chunk := model.Chunk{
		ID:         chunkID,
		FilePath:   filePath,
		StartLine:  startLine,
		EndLine:    endLine,
		Content:    content,
		Language:   "shell",
		Symbols:    symbols,
		TokenCount: util.EstimateTokenCount(content),
	}

	// Add symbols to symbol table
	for _, symbol := range symbols {
		symbolTable.AddDefinition(symbol, model.SymbolDefinition{
			Name:      symbol,
			ChunkID:   chunkID,
			FilePath:  filePath,
			StartLine: startLine,
			EndLine:   endLine,
			Type:      "function",
		})
	}

	return chunk
}

// isFunctionStart checks if a line is a function definition
func isFunctionStart(line string) bool {
	// Match patterns like:
	// function name() {
	// function name {
	// name() {
	if strings.HasPrefix(line, "function ") && (strings.Contains(line, "()") || strings.HasSuffix(line, "{")) {
		return true
	}

	// Check for name() { pattern
	funcPattern := regexp.MustCompile(`^\s*\w+\s*\(\s*\)\s*\{`)
	return funcPattern.MatchString(line)
}

// extractFunctions extracts function names from shell script content
func (c *ShellChunker) extractFunctions(content string) []string {
	var symbols []string

	// Look for function definitions: function name() or name()
	funcPattern1 := regexp.MustCompile(`function\s+(\w+)\s*\(\s*\)`)
	funcPattern2 := regexp.MustCompile(`function\s+(\w+)\s*\{`)
	funcPattern3 := regexp.MustCompile(`\b(\w+)\s*\(\s*\)\s*\{`)

	matches1 := funcPattern1.FindAllStringSubmatch(content, -1)
	for _, match := range matches1 {
		symbols = append(symbols, match[1])
	}

	matches2 := funcPattern2.FindAllStringSubmatch(content, -1)
	for _, match := range matches2 {
		symbols = append(symbols, match[1])
	}

	matches3 := funcPattern3.FindAllStringSubmatch(content, -1)
	for _, match := range matches3 {
		symbols = append(symbols, match[1])
	}

	// Deduplicate
	symbolMap := make(map[string]bool)
	for _, symbol := range symbols {
		symbolMap[symbol] = true
	}

	var uniqueSymbols []string
	for symbol := range symbolMap {
		uniqueSymbols = append(uniqueSymbols, symbol)
	}

	return uniqueSymbols
}
