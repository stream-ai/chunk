// chunker/fallbackchunker.go
package chunker

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ChunkFallback(filepath string, linesPerChunk int) ([]Chunk, error) {
	f, err := os.Open(filepath)
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
			chunkID := fmt.Sprintf("fallback_%d_%d", currentStartLine, lineCount)
			chunks = append(chunks, Chunk{
				FilePath:  filepath,
				ChunkID:   chunkID,
				ChunkType: "fallback",
				StartLine: currentStartLine,
				EndLine:   lineCount,
				Code:      strings.Join(currentLines, "\n"),
			})
			currentStartLine = lineCount + 1
			currentLines = nil
		}
	}
	// remainder
	if len(currentLines) > 0 {
		chunkID := fmt.Sprintf("fallback_%d_%d", currentStartLine, lineCount)
		chunks = append(chunks, Chunk{
			FilePath:  filepath,
			ChunkID:   chunkID,
			ChunkType: "fallback",
			StartLine: currentStartLine,
			EndLine:   lineCount,
			Code:      strings.Join(currentLines, "\n"),
		})
	}

	return chunks, nil
}
