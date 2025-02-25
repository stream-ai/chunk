package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Chunk represents a single code chunk
type Chunk struct {
	ID        string `json:"id"`
	FilePath  string `json:"file_path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Content   string `json:"content"`
	Language  string `json:"language"`
}

// ChunkResult contains all chunks from processing
type ChunkResult struct {
	Chunks []Chunk `json:"chunks"`
}

// GitIgnoreManager manages .gitignore rules across the directory tree
type GitIgnoreManager struct {
	ignores map[string]*ignore.GitIgnore
	rootDir string
}

// NewGitIgnoreManager creates a new GitIgnoreManager
func NewGitIgnoreManager(rootDir string) *GitIgnoreManager {
	return &GitIgnoreManager{
		ignores: make(map[string]*ignore.GitIgnore),
		rootDir: rootDir,
	}
}

// LoadGitIgnores loads all .gitignore files in the directory tree
func (gim *GitIgnoreManager) LoadGitIgnores() error {
	return filepath.Walk(gim.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Base(path) == ".gitignore" {
			dir := filepath.Dir(path)
			gitignore, err := ignore.CompileIgnoreFile(path)
			if err != nil {
				return err
			}
			gim.ignores[dir] = gitignore
		}

		return nil
	})
}

// IsIgnored checks if a file should be ignored based on .gitignore rules
func (gim *GitIgnoreManager) IsIgnored(path string) bool {
	// Start from the file's directory and walk up to the root
	dir := filepath.Dir(path)
	for {
		if gitignore, exists := gim.ignores[dir]; exists {
			relPath, err := filepath.Rel(dir, path)
			if err == nil && gitignore.MatchesPath(relPath) {
				return true
			}
		}

		// Stop if we've reached the root or beyond
		if dir == gim.rootDir || !strings.HasPrefix(dir, gim.rootDir) {
			break
		}

		// Move up one directory
		dir = filepath.Dir(dir)
	}

	return false
}

// isBinaryFile determines if a file is likely binary based on extension
func isBinaryFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	binaryExts := map[string]bool{
		".exe":   true,
		".dll":   true,
		".so":    true,
		".dylib": true,
		".bin":   true,
		".obj":   true,
		".o":     true,
		".a":     true,
		".lib":   true,
		".png":   true,
		".jpg":   true,
		".jpeg":  true,
		".gif":   true,
		".bmp":   true,
		".tiff":  true,
		".ico":   true,
		".zip":   true,
		".tar":   true,
		".gz":    true,
		".bz2":   true,
		".7z":    true,
		".rar":   true,
		".pdf":   true,
		".doc":   true,
		".docx":  true,
		".xls":   true,
		".xlsx":  true,
		".ppt":   true,
		".pptx":  true,
	}

	return binaryExts[ext]
}

// detectLanguage determines the language based on file extension
func detectLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	baseName := strings.ToLower(filepath.Base(filename))

	switch ext {
	case ".go":
		return "go"
	case ".ts":
		return "typescript"
	case ".js":
		return "javascript"
	case ".jsx":
		return "jsx"
	case ".tsx":
		return "tsx"
	case ".css":
		return "css"
	case ".scss":
		return "scss"
	case ".less":
		return "less"
	case ".html":
		return "html"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	case ".md", ".markdown":
		return "markdown"
	}

	// Check for special filenames
	if baseName == "dockerfile" || strings.HasPrefix(baseName, "dockerfile.") {
		return "dockerfile"
	}

	return "unknown"
}

// processFile processes a single file and returns its chunks
func processFile(filePath string) ([]Chunk, error) {
	language := detectLanguage(filePath)

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var chunks []Chunk

	// Choose chunking strategy based on language
	switch language {
	case "go":
		chunks = chunkGoFile(filePath, lines)
	case "typescript", "javascript", "jsx", "tsx":
		chunks = chunkJSFile(filePath, lines, language)
	case "json", "yaml", "toml", "markdown", "dockerfile":
		chunks = chunkMarkupFile(filePath, lines, language)
	default:
		// Use line-based chunking for unknown file types
		chunks = chunkLineBasedFile(filePath, lines, language)
	}

	return chunks, nil
}

// chunkGoFile chunks a Go file with awareness of function boundaries
func chunkGoFile(filePath string, lines []string) []Chunk {
	var chunks []Chunk
	var currentChunk []string
	startLine := 1
	insideBlock := false
	braceCount := 0

	funcRegex := regexp.MustCompile(`^func\s+\w+.*{$`)
	typeRegex := regexp.MustCompile(`^type\s+\w+.*{$`)

	for i, line := range lines {
		// Check for new function or type definition
		if !insideBlock && (funcRegex.MatchString(line) || typeRegex.MatchString(line)) {
			// If we have a current chunk, add it
			if len(currentChunk) > 0 {
				content := strings.Join(currentChunk, "\n")
				chunks = append(chunks, Chunk{
					ID:        generateChunkID(filePath, content),
					FilePath:  filePath,
					StartLine: startLine,
					EndLine:   i,
					Content:   content,
					Language:  "go",
				})

				currentChunk = []string{}
			}

			insideBlock = true
			startLine = i + 1
			braceCount = 1 // Opening brace on the function/type line
		}

		// Add the line to the current chunk
		currentChunk = append(currentChunk, line)

		// Count braces to track block boundaries
		if insideBlock {
			braceCount += strings.Count(line, "{") - strings.Count(line, "}")
			if braceCount == 0 {
				// Block has ended
				insideBlock = false
				content := strings.Join(currentChunk, "\n")
				chunks = append(chunks, Chunk{
					ID:        generateChunkID(filePath, content),
					FilePath:  filePath,
					StartLine: startLine,
					EndLine:   i + 1,
					Content:   content,
					Language:  "go",
				})

				currentChunk = []string{}
				startLine = i + 2
			}
		}

		// If we have accumulated many lines outside a block, create a chunk
		if !insideBlock && len(currentChunk) >= 50 {
			content := strings.Join(currentChunk, "\n")
			chunks = append(chunks, Chunk{
				ID:        generateChunkID(filePath, content),
				FilePath:  filePath,
				StartLine: startLine,
				EndLine:   i + 1,
				Content:   content,
				Language:  "go",
			})

			currentChunk = []string{}
			startLine = i + 2
		}
	}

	// Add any remaining lines as a chunk
	if len(currentChunk) > 0 {
		content := strings.Join(currentChunk, "\n")
		chunks = append(chunks, Chunk{
			ID:        generateChunkID(filePath, content),
			FilePath:  filePath,
			StartLine: startLine,
			EndLine:   len(lines),
			Content:   content,
			Language:  "go",
		})
	}

	return chunks
}

// chunkJSFile chunks a JS/TS file with awareness of function/class boundaries
func chunkJSFile(filePath string, lines []string, language string) []Chunk {
	var chunks []Chunk
	var currentChunk []string
	startLine := 1
	insideBlock := false
	braceCount := 0

	funcRegex := regexp.MustCompile(`(function\s+\w+|const\s+\w+\s*=\s*function|class\s+\w+|const\s+\w+\s*=\s*class).*{`)

	for i, line := range lines {
		// Check for new function or class definition
		if !insideBlock && funcRegex.MatchString(line) {
			// If we have a current chunk, add it
			if len(currentChunk) > 0 {
				content := strings.Join(currentChunk, "\n")
				chunks = append(chunks, Chunk{
					ID:        generateChunkID(filePath, content),
					FilePath:  filePath,
					StartLine: startLine,
					EndLine:   i,
					Content:   content,
					Language:  language,
				})

				currentChunk = []string{}
			}

			insideBlock = true
			startLine = i + 1
			braceCount = strings.Count(line, "{")
		}

		// Add the line to the current chunk
		currentChunk = append(currentChunk, line)

		// Count braces to track block boundaries
		if insideBlock {
			braceCount += strings.Count(line, "{") - strings.Count(line, "}")
			if braceCount == 0 {
				// Block has ended
				insideBlock = false
				content := strings.Join(currentChunk, "\n")
				chunks = append(chunks, Chunk{
					ID:        generateChunkID(filePath, content),
					FilePath:  filePath,
					StartLine: startLine,
					EndLine:   i + 1,
					Content:   content,
					Language:  language,
				})

				currentChunk = []string{}
				startLine = i + 2
			}
		}

		// If we have accumulated many lines outside a block, create a chunk
		if !insideBlock && len(currentChunk) >= 30 {
			content := strings.Join(currentChunk, "\n")
			chunks = append(chunks, Chunk{
				ID:        generateChunkID(filePath, content),
				FilePath:  filePath,
				StartLine: startLine,
				EndLine:   i + 1,
				Content:   content,
				Language:  language,
			})

			currentChunk = []string{}
			startLine = i + 2
		}
	}

	// Add any remaining lines as a chunk
	if len(currentChunk) > 0 {
		content := strings.Join(currentChunk, "\n")
		chunks = append(chunks, Chunk{
			ID:        generateChunkID(filePath, content),
			FilePath:  filePath,
			StartLine: startLine,
			EndLine:   len(lines),
			Content:   content,
			Language:  language,
		})
	}

	return chunks
}

// chunkMarkupFile chunks a markup file (JSON, YAML, etc.)
func chunkMarkupFile(filePath string, lines []string, language string) []Chunk {
	// For markup files, use a larger chunk size
	chunkSize := 100
	return chunkByLineCount(filePath, lines, language, chunkSize)
}

// chunkLineBasedFile chunks a file based on line count (fallback method)
func chunkLineBasedFile(filePath string, lines []string, language string) []Chunk {
	chunkSize := 50
	return chunkByLineCount(filePath, lines, language, chunkSize)
}

// chunkByLineCount chunks a file into fixed-size chunks based on line count
func chunkByLineCount(filePath string, lines []string, language string, chunkSize int) []Chunk {
	var chunks []Chunk

	for i := 0; i < len(lines); i += chunkSize {
		endIdx := i + chunkSize
		if endIdx > len(lines) {
			endIdx = len(lines)
		}

		chunkLines := lines[i:endIdx]
		content := strings.Join(chunkLines, "\n")

		chunks = append(chunks, Chunk{
			ID:        generateChunkID(filePath, content),
			FilePath:  filePath,
			StartLine: i + 1,
			EndLine:   endIdx,
			Content:   content,
			Language:  language,
		})
	}

	return chunks
}

// generateChunkID creates a stable ID for a chunk based on file path and content
func generateChunkID(filePath, content string) string {
	hasher := sha256.New()
	hasher.Write([]byte(filePath))
	hasher.Write([]byte(content))
	return hex.EncodeToString(hasher.Sum(nil))
}

// processDirectory processes all files in a directory recursively
func processDirectory(rootDir string) (ChunkResult, error) {
	var result ChunkResult

	// Initialize and load .gitignore files
	ignoreManager := NewGitIgnoreManager(rootDir)
	if err := ignoreManager.LoadGitIgnores(); err != nil {
		return result, err
	}

	// Walk the directory recursively
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories that should be excluded
		if info.IsDir() {
			dirname := info.Name()

			// Skip .git directory
			if dirname == ".git" {
				return filepath.SkipDir
			}

			// Skip other common directories to ignore
			dirsToSkip := map[string]bool{
				"node_modules": true,
				"vendor":       true,
				"dist":         true,
				"build":        true,
				"target":       true,
			}

			if dirsToSkip[dirname] {
				return filepath.SkipDir
			}

			return nil
		}

		// Check if the file should be ignored
		if ignoreManager.IsIgnored(path) {
			return nil
		}

		// Skip binary files
		if isBinaryFile(path) {
			return nil
		}

		// Process the file
		fileChunks, err := processFile(path)
		if err != nil {
			return fmt.Errorf("error processing %s: %v", path, err)
		}

		result.Chunks = append(result.Chunks, fileChunks...)
		return nil
	})

	return result, err
}

// outputResult writes the result to the specified output path
func outputResult(result ChunkResult, outputPath string) error {
	var output io.Writer

	if outputPath == "-" {
		output = os.Stdout
	} else {
		file, err := os.Create(outputPath)
		if err != nil {
			return err
		}
		defer file.Close()
		output = file
	}

	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func main() {
	// Create the root command
	rootCmd := &cobra.Command{
		Use:   "chunk",
		Short: "Chunk splits code files into discrete chunks for AI analysis",
		Long: `Chunk is a tool that recursively scans a source tree and splits code files into 
discrete chunks for AI code assistance and RAG workflows. It supports various languages
and respects .gitignore rules.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Get the options
			dir := viper.GetString("dir")
			output := viper.GetString("output")

			// Run the chunking process
			result, err := processDirectory(dir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Output the result
			err = outputResult(result, output)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	// Add flags
	rootCmd.Flags().StringP("dir", "d", ".", "Directory to process")
	rootCmd.Flags().StringP("output", "o", "-", "Output file (- for stdout)")

	// Bind flags to viper
	viper.BindPFlags(rootCmd.Flags())

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
