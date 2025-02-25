package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/stream-ai/chunk/internal/chunker"
	"github.com/stream-ai/chunk/internal/detector"
	"github.com/stream-ai/chunk/internal/model"
	"github.com/stream-ai/chunk/internal/output"
	"github.com/stream-ai/chunk/pkg/gitignore"
	"github.com/stream-ai/chunk/pkg/util"
)

// rootOptions contains the command-line options
type rootOptions struct {
	Dir          string
	Output       string
	Format       string
	MinChunkSize int
	MaxChunkSize int
}

// NewRootCommand creates the root command for the application
func NewRootCommand() *cobra.Command {
	opts := &rootOptions{}

	// Create command
	cmd := &cobra.Command{
		Use:   "chunk",
		Short: "Chunk splits code files into discrete chunks for AI analysis",
		Long: `Chunk is a tool that recursively scans a source tree and splits code files into 
discrete chunks for AI code assistance and RAG workflows. It supports various languages
and frameworks, respects .gitignore rules, and provides rich metadata for vector databases.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChunk(opts)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&opts.Dir, "dir", "d", ".", "Directory to process")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "-", "Output file (- for stdout)")
	cmd.Flags().StringVarP(&opts.Format, "format", "f", "vector-ready", "Output format (vector-ready, json, or jsonl)")
	cmd.Flags().IntVarP(&opts.MinChunkSize, "min-chunk-size", "m", 10, "Minimum chunk size in lines")
	cmd.Flags().IntVarP(&opts.MaxChunkSize, "max-chunk-size", "M", 50, "Maximum chunk size in lines")

	// Bind flags to viper
	viper.BindPFlags(cmd.Flags())

	return cmd
}

// runChunk executes the main functionality
func runChunk(opts *rootOptions) error {
	// Initialize components
	langDetector := detector.NewDefaultLanguageDetector()
	frameworkDetector := detector.NewDefaultFrameworkDetector()

	// Initialize chunker registry
	chunkerRegistry := chunker.NewChunkerRegistry()
	chunkerRegistry.Register(chunker.NewGoChunker())
	chunkerRegistry.Register(chunker.NewShellChunker())
	chunkerRegistry.Register(chunker.NewDockerfileChunker())
	chunkerRegistry.Register(chunker.NewGenericChunker()) // Fallback chunker for unknown types

	// Initialize formatter registry
	formatterRegistry := output.NewFormatterRegistry()
	formatterRegistry.Register("json", output.NewJSONFormatter(true))
	formatterRegistry.Register("jsonl", output.NewJSONLinesFormatter())
	formatterRegistry.Register("vector-ready", output.NewJSONFormatter(true))

	// Get absolute path for the root directory
	rootDir, err := filepath.Abs(opts.Dir)
	if err != nil {
		return err
	}

	// Initialize GitIgnore manager
	ignoreManager := gitignore.NewManager(rootDir)
	if err := ignoreManager.LoadIgnores(); err != nil {
		return err
	}

	// Process the directory
	result, err := processDirectory(rootDir, opts.MinChunkSize, opts.MaxChunkSize,
		opts.Format, langDetector, frameworkDetector, chunkerRegistry, ignoreManager)
	if err != nil {
		return err
	}

	// Output the result
	return outputResult(result, opts.Output, opts.Format, formatterRegistry)
}

// outputResult writes the result to the specified output path
func outputResult(
	result model.ChunkResult,
	outputPath string,
	format string,
	formatterRegistry *output.FormatterRegistry,
) error {
	// Get the appropriate formatter
	formatter, exists := formatterRegistry.Get(format)
	if !exists {
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Open output file or use stdout
	var w io.Writer
	if outputPath == "-" {
		w = os.Stdout
	} else {
		file, err := os.Create(outputPath)
		if err != nil {
			return err
		}
		defer file.Close()
		w = file
	}

	// Format and write the output
	return formatter.Format(w, result)
}

// processDirectory processes all files in a directory recursively
func processDirectory(
	rootDir string,
	minChunkSize int,
	maxChunkSize int,
	format string,
	langDetector detector.LanguageDetector,
	frameworkDetector detector.FrameworkDetector,
	chunkerRegistry *chunker.ChunkerRegistry,
	ignoreManager *gitignore.Manager,
) (model.ChunkResult, error) {
	var result model.ChunkResult
	symbolTable := model.NewSymbolTable()

	// First pass: Process all files and build chunks
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories that should be excluded
		if info.IsDir() {
			dirName := info.Name()
			if util.ShouldSkipDirectory(dirName) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if the file should be ignored
		if ignoreManager.IsIgnored(path) {
			return nil
		}

		// Skip binary files
		if util.IsBinaryFile(path) {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading %s: %v", path, err)
		}

		// Detect language and framework
		language := langDetector.DetectLanguage(path, content)
		framework := frameworkDetector.DetectFramework(path, content, language)

		// Find appropriate chunker
		fileChunker := chunkerRegistry.FindChunker(path, language, framework)
		if fileChunker == nil {
			// Use generic chunker as fallback
			fileChunker = chunker.NewGenericChunker()
		}

		// Process the file
		options := chunker.ChunkingOptions{
			MinChunkSize: minChunkSize,
			MaxChunkSize: maxChunkSize,
		}

		fileChunks, err := fileChunker.Chunk(path, content, symbolTable, options)
		if err != nil {
			return fmt.Errorf("error processing %s: %v", path, err)
		}

		// Store chunks in symbol table and results
		for _, chunk := range fileChunks {
			symbolTable.AddChunk(chunk)
			result.Chunks = append(result.Chunks, chunk)
		}

		return nil
	})
	if err != nil {
		return result, err
	}

	// Second pass: Only for vector-ready format, add related chunks
	if format == "vector-ready" {
		for i, chunk := range result.Chunks {
			result.Chunks[i].RelatedChunks = symbolTable.FindRelatedChunks(chunk)
		}
	}

	return result, nil
}
