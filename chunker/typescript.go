package chunker

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

func ChunkTypeScriptFile(filepath, lang string) ([]Chunk, error) {
	// If you want to detect React, pass additional flags to the script, etc.
	// We'll do a simple example that just calls "tsparser.ts" with the file path.
	cmd := exec.Command("npx", "ts-node", "tsparser/tsparser.ts", filepath)
	// ^ "npx ts-node tsparser.ts <filepath>" requires that tsparser.ts is in the same directory,
	// or you can do an absolute path to tsparser.ts if needed.

	// If tsparser is in a separate folder, e.g. chunk/tsparser/tsparser.ts
	// you'd do something like:
	// cmd.Dir = "/workspace/chunk/tsparser"

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ts1 parsing failed: %v\n%s", err, string(out))
	}

	// 'out' should be JSON array of chunk objects
	var chunks []Chunk
	if err := json.Unmarshal(out, &chunks); err != nil {
		return nil, fmt.Errorf("invalid JSON from tsparser: %v\n%s", err, string(out))
	}

	return chunks, nil
}

// func parseJSONChunks(data []byte) ([]Chunk, error) {
// 	// Unmarshal to []Chunk
// 	// This is just a placeholder for the actual parsing logic
// 	var chunks []Chunk
// 	err := json.Unmarshal(data, &chunks)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to parse JSON: %v", err)
// 	}

// 	return chunks, nil
// }
