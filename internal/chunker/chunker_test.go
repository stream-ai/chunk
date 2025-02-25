// internal/chunker/chunker_test.go

package chunker_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stream-ai/chunk/internal/chunker"
	"github.com/stream-ai/chunk/internal/detector"
	"github.com/stream-ai/chunk/internal/model"
)

func TestChunkerRegistry(t *testing.T) {
	registry := chunker.NewChunkerRegistry()

	// Register test chunkers
	goChunker := chunker.NewGoChunker()
	shellChunker := chunker.NewShellChunker()
	dockerfileChunker := chunker.NewDockerfileChunker()

	registry.Register(goChunker)
	registry.Register(shellChunker)
	registry.Register(dockerfileChunker)

	tests := []struct {
		name        string
		filePath    string
		language    string
		framework   string
		wantChunker chunker.Chunker
	}{
		{"Go file", "test.go", "go", "", goChunker},
		{"Shell file", "test.sh", "shell", "", shellChunker},
		{"Bash file", "test.bash", "shell", "", shellChunker},
		{"Dockerfile", "Dockerfile", "dockerfile", "", dockerfileChunker},
		{"Unknown file", "test.xyz", "unknown", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.FindChunker(tt.filePath, tt.language, tt.framework)
			if got != tt.wantChunker {
				t.Errorf("ChunkerRegistry.FindChunker() = %v, want %v", got, tt.wantChunker)
			}
		})
	}
}

// TestGoChunker tests the Go chunker functionality
func TestGoChunker(t *testing.T) {
	chunkerImpl := chunker.NewGoChunker()
	symbolTable := model.NewSymbolTable()

	// Simple Go file content
	content := []byte(`package test

import (
	"fmt"
)

// Simple function
func HelloWorld() string {
	return "Hello, World!"
}

// Struct definition
type Person struct {
	Name string
	Age  int
}

// Method definition
func (p *Person) Greet() string {
	return fmt.Sprintf("Hello, my name is %s", p.Name)
}
`)

	chunks, err := chunkerImpl.Chunk("test.go", content, symbolTable, chunker.ChunkingOptions{
		MinChunkSize: 5,
		MaxChunkSize: 50,
	})
	if err != nil {
		t.Fatalf("Chunker.Chunk() error = %v", err)
	}

	// We should have multiple chunks: imports, function, struct, method
	if len(chunks) < 3 {
		t.Errorf("Expected at least 3 chunks, got %d", len(chunks))
	}

	// Check extracted symbols
	foundFunction := false
	foundStruct := false
	foundMethod := false

	for _, chunk := range chunks {
		for _, symbol := range chunk.Symbols {
			switch symbol {
			case "HelloWorld":
				foundFunction = true
			case "Person":
				foundStruct = true
			case "Person.Greet":
				foundMethod = true
			}
		}
	}

	if !foundFunction {
		t.Error("Failed to extract HelloWorld function symbol")
	}
	if !foundStruct {
		t.Error("Failed to extract Person struct symbol")
	}
	if !foundMethod {
		t.Error("Failed to extract Greet method symbol")
	}
}

// TestShellChunker tests the Shell script chunker
func TestShellChunker(t *testing.T) {
	chunkerImpl := chunker.NewShellChunker()
	symbolTable := model.NewSymbolTable()

	// Shell script content
	content := []byte(`#!/bin/bash

# A simple shell script

# First function
function say_hello() {
    echo "Hello, $1!"
}

# Second function
goodbye() {
    echo "Goodbye, $1!"
}

# Main code
name="World"
say_hello "$name"
goodbye "$name"
`)

	chunks, err := chunkerImpl.Chunk("test.sh", content, symbolTable, chunker.ChunkingOptions{
		MinChunkSize: 5,
		MaxChunkSize: 50,
	})
	if err != nil {
		t.Fatalf("Chunker.Chunk() error = %v", err)
	}

	// We should have at least 2 function chunks and 1 main code chunk
	if len(chunks) < 3 {
		t.Errorf("Expected at least 3 chunks, got %d", len(chunks))
	}

	// Check for function symbols
	foundSayHello := false
	foundGoodbye := false

	for _, chunk := range chunks {
		for _, symbol := range chunk.Symbols {
			switch symbol {
			case "say_hello":
				foundSayHello = true
			case "goodbye":
				foundGoodbye = true
			}
		}
	}

	if !foundSayHello {
		t.Error("Failed to extract say_hello function symbol")
	}
	if !foundGoodbye {
		t.Error("Failed to extract goodbye function symbol")
	}
}

// TestDockerfileChunker tests the Dockerfile chunker
func TestDockerfileChunker(t *testing.T) {
	chunkerImpl := chunker.NewDockerfileChunker()
	symbolTable := model.NewSymbolTable()

	// Dockerfile content
	content := []byte(`# Stage 1: Build
FROM golang:1.18 AS build

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/myapp

# Stage 2: Run
FROM alpine:latest

WORKDIR /app
COPY --from=build /app/myapp .

EXPOSE 8080
CMD ["/app/myapp"]
`)

	chunks, err := chunkerImpl.Chunk("Dockerfile", content, symbolTable, chunker.ChunkingOptions{
		MinChunkSize: 5,
		MaxChunkSize: 50,
	})
	if err != nil {
		t.Fatalf("Chunker.Chunk() error = %v", err)
	}

	// We should have multiple chunks for each major instruction
	if len(chunks) < 2 {
		t.Errorf("Expected at least 2 chunks, got %d", len(chunks))
	}

	// Check for stage and instruction symbols
	foundFromBuild := false
	foundFromAlpine := false
	foundCopy := false
	foundRun := false

	for _, chunk := range chunks {
		for _, symbol := range chunk.Symbols {
			switch symbol {
			case "FROM":
				if !foundFromBuild {
					foundFromBuild = true
				} else {
					foundFromAlpine = true
				}
			case "COPY":
				foundCopy = true
			case "RUN":
				foundRun = true
			}
		}
	}

	if !foundFromBuild || !foundFromAlpine {
		t.Error("Failed to extract FROM instruction symbols")
	}
	if !foundCopy {
		t.Error("Failed to extract COPY instruction symbol")
	}
	if !foundRun {
		t.Error("Failed to extract RUN instruction symbol")
	}
}

// TestEndToEndChunking tests the complete chunking process with different file types
func TestEndToEndChunking(t *testing.T) {
	// Setup a temporary test directory with sample files
	tmpDir, err := os.MkdirTemp("", "chunker-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a Go file
	goFile := filepath.Join(tmpDir, "main.go")
	err = os.WriteFile(goFile, []byte(`package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`), 0o644)
	if err != nil {
		t.Fatalf("Failed to write Go file: %v", err)
	}

	// Create a shell script
	shellFile := filepath.Join(tmpDir, "script.sh")
	err = os.WriteFile(shellFile, []byte(`#!/bin/bash
echo "Hello from shell!"
`), 0o644)
	if err != nil {
		t.Fatalf("Failed to write shell file: %v", err)
	}

	// Create a Dockerfile
	dockerFile := filepath.Join(tmpDir, "Dockerfile")
	err = os.WriteFile(dockerFile, []byte(`FROM alpine
CMD ["echo", "Hello from Docker!"]
`), 0o644)
	if err != nil {
		t.Fatalf("Failed to write Dockerfile: %v", err)
	}

	// Set up the chunker registry
	registry := chunker.NewChunkerRegistry()
	registry.Register(chunker.NewGoChunker())
	registry.Register(chunker.NewShellChunker())
	registry.Register(chunker.NewDockerfileChunker())
	registry.Register(chunker.NewGenericChunker())

	langDetector := detector.NewDefaultLanguageDetector()
	symbolTable := model.NewSymbolTable()

	// Process each file
	files := []string{goFile, shellFile, dockerFile}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("Failed to read file %s: %v", file, err)
		}

		lang := langDetector.DetectLanguage(file, content)
		chunkerImpl := registry.FindChunker(file, lang, "")

		if chunkerImpl == nil {
			t.Errorf("No chunker found for %s with language %s", file, lang)
			continue
		}

		chunks, err := chunkerImpl.Chunk(file, content, symbolTable, chunker.ChunkingOptions{
			MinChunkSize: 5,
			MaxChunkSize: 50,
		})
		if err != nil {
			t.Errorf("Failed to chunk file %s: %v", file, err)
			continue
		}

		if len(chunks) == 0 {
			t.Errorf("No chunks produced for file %s", file)
		}

		// Verify language field is properly set
		for i, chunk := range chunks {
			if chunk.Language == "" {
				t.Errorf("Chunk %d from file %s has empty language", i, file)
			}
		}
	}
}
