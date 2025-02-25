package model

// Chunk represents a single code chunk
type Chunk struct {
	ID            string   `json:"id"`
	FilePath      string   `json:"file_path"`
	StartLine     int      `json:"start_line"`
	EndLine       int      `json:"end_line"`
	Content       string   `json:"content"`
	Language      string   `json:"language"`
	Framework     string   `json:"framework,omitempty"`
	Symbols       []string `json:"symbols,omitempty"`
	Imports       []string `json:"imports,omitempty"`
	RelatedChunks []string `json:"related_chunks,omitempty"`
	TokenCount    int      `json:"token_count,omitempty"`
}

// ChunkResult contains all chunks from processing
type ChunkResult struct {
	Chunks []Chunk `json:"chunks"`
}
