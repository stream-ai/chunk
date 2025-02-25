package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	// This library does the parsing of .gitignore patterns
	gitignore "github.com/sabhiram/go-gitignore"
	"github.com/stream-ai/chunk/internal/chunker"

	"github.com/spf13/cobra"
)

var (
	outputPath    string
	forcedLang    string
	fallbackLines int
	rootDir       string
)

var rootCmd = &cobra.Command{
	Use:   "chunk",
	Short: "Chunk code files in the current directory (and subdirs) using .gitignore logic",
	Long: `chunk automatically scans the specified directory (default: current dir)
and all subdirectories, merges local .gitignore files, and chunks code.
No file arguments are required or accepted.`,
	RunE: runChunk,
}

func init() {
	// Common flags
	rootCmd.Flags().StringVarP(&outputPath, "output", "o", "-", "Output destination ('-' for STDOUT)")
	rootCmd.Flags().StringVarP(&forcedLang, "lang", "l", "", "Force a specific language parser (go|ts|react|fallback)")
	rootCmd.Flags().IntVar(&fallbackLines, "fallback-lines", 200, "Number of lines per fallback chunk")

	// Directory to walk (default: current directory)
	rootCmd.Flags().StringVar(&rootDir, "dir", ".", "Root directory from which to chunk everything")
}

// Execute is called by main.go to run the CLI.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runChunk(cmd *cobra.Command, args []string) error {
	// We'll gather all the resulting chunks in memory
	var allChunks []chunker.Chunk

	err := walkWithGitIgnore(rootDir, func(path string, isDir bool) error {
		if isDir {
			// For directories, do nothing special
			return nil
		}
		// For each file: call your actual chunking logic
		fileChunks, cErr := chunker.ProcessFile(path, forcedLang, fallbackLines)
		if cErr != nil {
			// If there's an error chunking, log and skip
			log.Printf("Error chunking file %s: %v\n", path, cErr)
			return nil
		}
		allChunks = append(allChunks, fileChunks...)
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking directory: %w", err)
	}

	// Output the chunks
	return writeChunks(allChunks)
}

// writeChunks writes chunk data as JSON
func writeChunks(chunks []chunker.Chunk) error {
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
	enc.SetIndent("", "  ")
	return enc.Encode(chunks)
}

// ------------------
// .gitignore Stack Logic
// ------------------

// We'll store the compiled GitIgnore plus the directory it was loaded from
type ignoreItem struct {
	ig  *gitignore.GitIgnore
	dir string
}

type ignoreStack struct {
	items []ignoreItem
}

func (s *ignoreStack) push(item ignoreItem) {
	s.items = append(s.items, item)
}

func (s *ignoreStack) pop() {
	if len(s.items) > 0 {
		s.items = s.items[:len(s.items)-1]
	}
}

// Checks from top -> bottom. If any ignore rule matches, we skip the file/dir.
func (s *ignoreStack) isIgnored(path string) bool {
	// from top to bottom
	for i := len(s.items) - 1; i >= 0; i-- {
		baseDir := s.items[i].dir
		ig := s.items[i].ig
		rel, err := filepath.Rel(baseDir, path)
		if err == nil {
			if ig.MatchesPath(rel) {
				return true
			}
		}
	}
	return false
}

// walkWithGitIgnore recursively descends from "dir", merging .gitignore files.
func walkWithGitIgnore(dir string, fileCallback func(path string, isDir bool) error) error {
	var stack ignoreStack

	// We'll define a nested function for DFS so we can push/pop per directory
	var dfs func(string) error

	dfs = func(currentDir string) error {
		// Attempt to load a .gitignore in the currentDir
		gitignorePath := filepath.Join(currentDir, ".gitignore")
		if info, err := os.Stat(gitignorePath); err == nil && !info.IsDir() {
			// compile it
			ig, err := gitignore.CompileIgnoreFile(gitignorePath)
			if err != nil {
				return fmt.Errorf("error compiling.gitignore in %s: %w", currentDir, err)
			}
			// push onto stack
			stack.push(ignoreItem{
				ig:  ig,
				dir: currentDir,
			})
		}

		// read dir entries
		entries, err := os.ReadDir(currentDir)
		if err != nil {
			return err
		}

		for _, e := range entries {
			name := e.Name()
			path := filepath.Join(currentDir, name)
			isDir := e.IsDir()

			// Check stack to see if ignored
			if stack.isIgnored(path) {
				continue
			}

			// Callback
			if err := fileCallback(path, isDir); err != nil {
				return err
			}

			// Recurse if directory
			if isDir {
				if err := dfs(path); err != nil {
					return err
				}
			}
		}

		// If we pushed in this directory, pop it
		if len(stack.items) > 0 {
			top := stack.items[len(stack.items)-1]
			if top.dir == currentDir {
				stack.pop()
			}
		}
		return nil
	}

	return dfs(dir)
}
