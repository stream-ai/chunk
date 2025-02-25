package gitignore

import (
	"os"
	"path/filepath"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
)

// Manager handles .gitignore rules across the directory tree
type Manager struct {
	ignores map[string]*ignore.GitIgnore
	rootDir string
}

// NewManager creates a new GitIgnoreManager
func NewManager(rootDir string) *Manager {
	return &Manager{
		ignores: make(map[string]*ignore.GitIgnore),
		rootDir: rootDir,
	}
}

// LoadIgnores loads all .gitignore files in the directory tree
func (m *Manager) LoadIgnores() error {
	return filepath.Walk(m.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Base(path) == ".gitignore" {
			dir := filepath.Dir(path)
			gitignore, err := ignore.CompileIgnoreFile(path)
			if err != nil {
				return err
			}
			m.ignores[dir] = gitignore
		}

		return nil
	})
}

// IsIgnored checks if a file should be ignored based on .gitignore rules
func (m *Manager) IsIgnored(path string) bool {
	// Start from the file's directory and walk up to the root
	dir := filepath.Dir(path)
	for {
		if gitignore, exists := m.ignores[dir]; exists {
			relPath, err := filepath.Rel(dir, path)
			if err == nil && gitignore.MatchesPath(relPath) {
				return true
			}
		}

		// Stop if we've reached the root or beyond
		if dir == m.rootDir || !strings.HasPrefix(dir, m.rootDir) {
			break
		}

		// Move up one directory
		dir = filepath.Dir(dir)
	}

	return false
}
