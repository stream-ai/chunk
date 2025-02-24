package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/stream-ai/chunk/chunker"
)

var (
	outputPath    string
	forcedLang    string
	fallbackLines int
)

var rootCmd = &cobra.Command{
	Use:   "chunk [files or globs]",
	Short: "Chunk code files (Go, TypeScript React, or fallback) and output JSON chunks",
	Long: `chunk is a CLI tool that parses code in Go or TypeScript (React) 
and splits them into structured "chunks." For unknown or overridden file types, 
it uses a line-based fallback approach. It prints the chunks as JSON by default.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runChunk,
}

func init() {
	// Define CLI flags
	rootCmd.Flags().StringVarP(
		&outputPath, "output", "o", "-",
		"Output destination ('-' for STDOUT, default)",
	)
	rootCmd.Flags().StringVarP(
		&forcedLang, "lang", "l", "",
		"Force a specific language parser (go|ts|react|fallback)",
	)
	rootCmd.Flags().IntVar(
		&fallbackLines, "fallback-lines", 200,
		"Number of lines per fallback chunk",
	)

	// Optionally integrate with Viper for config if desired
	// viper.BindPFlag("fallback_lines", rootCmd.Flags().Lookup("fallback-lines"))
}

// Execute is called by main.go to run the CLI.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runChunk(cmd *cobra.Command, args []string) error {
	var allChunks []chunker.Chunk

	// Expand any file globs
	for _, path := range args {
		matches, err := filepath.Glob(path)
		if err != nil || matches == nil {
			matches = []string{path}
		}

		for _, file := range matches {
			fileChunks, err := processFile(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error chunking file %s: %v\n", file, err)
				continue
			}
			allChunks = append(allChunks, fileChunks...)
		}
	}

	// Output the results
	var out *os.File
	if outputPath == "-" {
		out = os.Stdout
	} else {
		f, err := os.Create(outputPath)
		if err != nil {
			return err
		}
		defer f.Close()
		out = f
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", " ") // pretty print
	if err := enc.Encode(allChunks); err != nil {
		return err
	}

	return nil
}

func processFile(path string) ([]chunker.Chunk, error) {
	// 0. Skip likely-binary files
	if chunker.IsLikelyBinaryByExtension(path) {
		log.Printf("Skipping binary file: %s\n", path)
		return nil, nil
	}

	// 1. Determine language if user hasn't forced one
	lang := forcedLang
	if lang == "" {
		lang = chunker.AutoDetectLanguage(path)
		log.Printf("Detected language for %s: %s\n", path, lang)
	}

	// 2. Dispatch to chunkers
	switch lang {
	case "go":
		return chunker.ChunkGoFile(path)
	case "ts", "react":
		return chunker.ChunkTypeScriptFile(path, lang)
	default:
		return chunker.ChunkFallback(path, fallbackLines)
	}
}
