package chunker

import (
	"bytes"
	"regexp"
	"strings"
	
	"github.com/stream-ai/chunk/internal/model"
	"github.com/stream-ai/chunk/pkg/util"
)

// DockerfileChunker implements the Chunker interface for Dockerfiles
type DockerfileChunker struct{}

// NewDockerfileChunker creates a new Dockerfile chunker
func NewDockerfileChunker() *DockerfileChunker {
	return &DockerfileChunker{}
}

// Language returns the language this chunker supports
func (c *DockerfileChunker) Language() string {
	return "dockerfile"
}

// CanHandle checks if this chunker can handle the given file
func (c *DockerfileChunker) CanHandle(filePath string, language string, framework string) bool {
	return language == "dockerfile"
}

// Chunk splits Dockerfile content into chunks based on stages and major sections
func (c *DockerfileChunker) Chunk(filePath string, content []byte, symbolTable *model.SymbolTable, options ChunkingOptions) ([]model.Chunk, error) {
	lines := bytes.Split(content, []byte("\n"))
	var chunks []model.Chunk
	
	// Group by stages and major instructions
	var currentChunk [][]byte
	startLine := 1
	
	for i, line := range lines {
		lineStr := string(line)
		trimmedLine := strings.TrimSpace(lineStr)
		
		// Skip empty lines and comments at the beginning of potential chunks
		if len(currentChunk) == 0 && (trimmedLine == "" || strings.HasPrefix(trimmedLine, "#")) {
			currentChunk = append(currentChunk, line)
			continue
		}
		
		// Check for stage boundaries and major instructions
		// FROM creates a new stage, so it should start a new chunk
		// Major instructions like RUN, COPY, ADD usually represent significant blocks
		if isChunkBoundary(trimmedLine) && len(currentChunk) > 0 {
			chunk := c.createChunk(filePath, currentChunk, startLine, i, symbolTable)
			chunks = append(chunks, chunk)
			currentChunk = [][]byte{}
			startLine = i + 1
		}
		
		// Add the line to the current chunk
		currentChunk = append(currentChunk, line)
		
		// If we have accumulated many lines in a single instruction (like a big RUN),
		// consider breaking it into smaller chunks
		if len(currentChunk) >= options.MaxChunkSize {
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
func (c *DockerfileChunker) createChunk(filePath string, lines [][]byte, startLine, endLine int, symbolTable *model.SymbolTable) model.Chunk {
	content := string(bytes.Join(lines, []byte("\n")))
	chunkID := util.GenerateID(filePath, content)
	
	// Extract instructions/stages as symbols
	symbols := c.extractInstructions(content)
	
	chunk := model.Chunk{
		ID:         chunkID,
		FilePath:   filePath,
		StartLine:  startLine,
		EndLine:    endLine,
		Content:    content,
		Language:   "dockerfile",
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
			Type:      "instruction",
		})
	}
	
	return chunk
}

// isChunkBoundary checks if a line should be considered a boundary between chunks
func isChunkBoundary(line string) bool {
	if line == "" {
		return false
	}
	
	// Standard Dockerfile instructions that typically mark logical sections
	majorInstructions := []string{
		"FROM", "MAINTAINER", "RUN", "CMD", "LABEL", 
		"EXPOSE", "ENV", "ADD", "COPY", "ENTRYPOINT", 
		"VOLUME", "USER", "WORKDIR", "ARG", "ONBUILD", 
		"STOPSIGNAL", "HEALTHCHECK", "SHELL",
	}
	
	for _, instr := range majorInstructions {
		if strings.HasPrefix(strings.ToUpper(line), instr+" ") {
			return true
		}
	}
	
	// Also check for stage names like "FROM xyz AS stagename"
	if strings.HasPrefix(strings.ToUpper(line), "FROM ") && strings.Contains(strings.ToUpper(line), " AS ") {
		return true
	}
	
	// Multi-line instructions like RUN that use backslash continuation
	// should not break at continuation lines
	return false
}

// extractInstructions extracts Dockerfile instructions and stage names
func (c *DockerfileChunker) extractInstructions(content string) []string {
	var symbols []string
	
	// Extract instructions (FROM, RUN, etc.)
	instructionPattern := regexp.MustCompile(`(?m)^(FROM|RUN|CMD|LABEL|EXPOSE|ENV|ADD|COPY|ENTRYPOINT|VOLUME|USER|WORKDIR|ARG|ONBUILD|STOPSIGNAL|HEALTHCHECK|SHELL)\s+`)
	
	matches := instructionPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		symbols = append(symbols, match[1])
	}
	
	// Extract stage names if present (FROM xyz AS stagename)
	stagePattern := regexp.MustCompile(`(?i)FROM\s+\S+\s+AS\s+(\S+)`)
	stageMatches := stagePattern.FindAllStringSubmatch(content, -1)
	for _, match := range stageMatches {
		symbols = append(symbols, "stage:"+match[1])
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