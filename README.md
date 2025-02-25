# Chunk

Chunk is a command-line tool written in Go that recursively scans source code repositories and splits them into discrete "chunks." It's designed with AI code assistance in mind, where these chunks can be embedded and stored in a vector database for Retrieval-Augmented Generation (RAG) workflows.

## Features

- Uses Go's robust standard library for fast and efficient file processing
- Automatically walks directories, respecting `.gitignore` rules
- Skips binary files and common non-source directories (node_modules, vendor, etc.)
- Language detection based on file extensions with fallback strategies
- Smart chunking that preserves semantic boundaries:
  - Functions, types, and methods in Go (using AST parsing)
  - Functions and blocks in shell scripts
  - Stages and instructions in Dockerfiles
  - Generic chunking for other file types
- Cross-reference tracking to maintain relationships between code chunks
- Rich metadata including:
  - Stable chunk IDs that only change when content changes
  - Symbol extraction (functions, classes, methods)
  - Import/dependency tracking
  - Related chunk references
  - Token count estimates
- Multiple output formats:
  - `vector-ready`: Default format with related chunks (best for RAG)
  - `json`: Standard JSON output
  - `jsonl`: JSON Lines (one JSON object per line)

## Installation

```bash
go install github.com/stream-ai/chunk@latest
```

Or build from source:

```bash
git clone https://github.com/stream-ai/chunk.git
cd chunk
go build
```

## Usage

Basic usage:

```bash
chunk --dir /path/to/codebase --output chunks.json
```

All options:

```bash
chunk --dir /path/to/codebase \
      --output chunks.json \
      --format json \
      --min-chunk-size 10 \
      --max-chunk-size 100
```

### Options

- `--dir`, `-d`: Directory to process (default: current directory)
- `--output`, `-o`: Output file (use "-" for stdout, default)
- `--format`, `-f`: Output format (vector-ready, json, or jsonl, default: vector-ready)
- `--min-chunk-size`, `-m`: Minimum chunk size in lines (default: 10)
- `--max-chunk-size`, `-M`: Maximum chunk size in lines (default: 50)

## Output Format

The tool produces JSON output containing:

```json
{
  "chunks": [
    {
      "id": "e4ace233f253b08b0c8fb0b8a19f8b84a892e1c39fe46408634010fd81e17ec4",
      "file_path": "main.go",
      "start_line": 1,
      "end_line": 50,
      "content": "package main\n\nimport (...",
      "language": "go",
      "framework": "",
      "symbols": ["main", "processFile", "NewChunker"],
      "imports": ["fmt", "os", "path/filepath"],
      "related_chunks": ["a8c2750dbd4fcae981d7c84a3996d30b7d252ee3038da4db7d7ea3768edd6c2b"],
      "token_count": 320
    }
  ]
}
```

## Examples

1. Process a Go project with the default vector-ready format:

```bash
chunk -d ./go-project -o chunks.json
```

2. Process a mixed codebase with the JSONL format:

```bash
chunk -d ./mixed-repo -o chunks.jsonl -f jsonl
```

3. Extract and analyze chunks using JQ:

```bash
# List all unique files
jq -r '[.chunks[].file_path] | unique | sort[]' chunks.json

# Count chunks per language
jq -r '[.chunks[].language] | group_by(.) | map({language: .[0], count: length}) | sort_by(.count) | reverse[]' chunks.json
```

## Language Support

Chunk provides specialized chunking for:

- **Go**: Uses AST parsing for accurate function, method, and type boundaries
- **Shell Scripts**: Detects function definitions and logical blocks
- **Dockerfiles**: Chunks based on stages and instructions

Other supported languages use generic chunking:
- JavaScript/TypeScript
- Python, Ruby, PHP
- Java, Kotlin, C, C++, C#, Rust
- HTML, CSS, JSON, YAML, TOML, Markdown

## Architecture

Chunk uses a plugin-based architecture with interfaces for:

- Language detection
- Chunking strategies
- Output formatting

This makes it easy to extend with new languages or output formats.

## Extending Chunk

Chunk is designed to be extensible. To add support for new languages:

1. Implement the `chunker.Chunker` interface
2. Register your implementation in `cmd/root.go`

For example:

```go
// Create a new chunker
myChunker := NewMyLanguageChunker()

// Register it
chunkerRegistry.Register(myChunker)
```

## Running Tests

```bash
go test ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License