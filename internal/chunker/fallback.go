package chunker

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ChunkFallback(filePath string, linesPerChunk int) ([]Chunk, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var chunks []Chunk
	var currentLines []string
	currentStartLine := 1

	scanner := bufio.NewScanner(f)
	lineCount := 0

	for scanner.Scan() {
		lineCount++
		currentLines = append(currentLines, scanner.Text())

		if len(currentLines) >= linesPerChunk {
			code := strings.Join(currentLines, "\n")
			hashedID := makeChunkID(filePath, code)
			chunkID := fmt.Sprintf("fallback_%d_%d:%s", currentStartLine, lineCount, hashedID[:8])

			chunks = append(chunks, Chunk{
				FilePath:  filePath,
				ChunkID:   chunkID,
				ChunkType: "fallback",
				StartLine: currentStartLine,
				EndLine:   lineCount,
				Code:      code,
			})
			currentStartLine = lineCount + 1
			currentLines = nil
		}
	}
	// remainder
	if len(currentLines) > 0 {
		code := strings.Join(currentLines, "\n")
		hashedID := makeChunkID(filePath, code)
		chunkID := fmt.Sprintf("fallback_%d_%d:%s", currentStartLine, lineCount, hashedID[:8])

		chunks = append(chunks, Chunk{
			FilePath:  filePath,
			ChunkID:   chunkID,
			ChunkType: "fallback",
			StartLine: currentStartLine,
			EndLine:   lineCount,
			Code:      code,
		})
	}

	return chunks, nil
}
