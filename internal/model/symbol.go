package model

import (
	"path/filepath"
	"sort"
	"strings"
)

// RelationStrength defines the strength of a relationship between chunks
type RelationStrength int

const (
	// RelationWeak represents a weak relationship (same package, etc.)
	RelationWeak RelationStrength = 1

	// RelationMedium represents a medium-strength relationship (import relation, etc.)
	RelationMedium RelationStrength = 5

	// RelationStrong represents a strong relationship (direct call, type-method, etc.)
	RelationStrong RelationStrength = 10
)

// SymbolDefinition represents a symbol (function, class, etc) defined in the code
type SymbolDefinition struct {
	Name      string
	ChunkID   string
	FilePath  string
	StartLine int
	EndLine   int
	Type      string // "function", "interface", "type", "const", "var", etc.
}

// SymbolReference represents a usage of a symbol
type SymbolReference struct {
	Name     string
	ChunkID  string
	FilePath string
	Line     int
}

// SymbolTable holds all the detected symbols, their references, and chunks
// We'll simplify by using string as the key type since that's what we're using for symbols
type SymbolTable struct {
	Definitions map[string][]SymbolDefinition
	References  map[string][]SymbolReference
	Chunks      map[string]Chunk // Map of chunk ID to chunk
}

// NewSymbolTable creates a new empty symbol table
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		Definitions: make(map[string][]SymbolDefinition),
		References:  make(map[string][]SymbolReference),
		Chunks:      make(map[string]Chunk),
	}
}

// AddDefinition adds a symbol definition
func (st *SymbolTable) AddDefinition(name string, def SymbolDefinition) {
	st.Definitions[name] = append(st.Definitions[name], def)
}

// AddReference adds a symbol reference
func (st *SymbolTable) AddReference(name string, ref SymbolReference) {
	st.References[name] = append(st.References[name], ref)
}

// AddChunk adds a chunk to the symbol table
func (st *SymbolTable) AddChunk(chunk Chunk) {
	st.Chunks[chunk.ID] = chunk
}

// GetChunk retrieves a chunk by ID
func (st *SymbolTable) GetChunk(id string) (Chunk, bool) {
	chunk, exists := st.Chunks[id]
	return chunk, exists
}

// AllChunks returns all chunks in the symbol table
func (st *SymbolTable) AllChunks() []Chunk {
	chunks := make([]Chunk, 0, len(st.Chunks))
	for _, chunk := range st.Chunks {
		chunks = append(chunks, chunk)
	}
	return chunks
}

// FindDefinitionsByType finds definitions of a specific type
func (st *SymbolTable) FindDefinitionsByType(defType string) []SymbolDefinition {
	var result []SymbolDefinition

	for _, defs := range st.Definitions {
		for _, def := range defs {
			if def.Type == defType {
				result = append(result, def)
			}
		}
	}

	return result
}

// FindRelatedChunks finds chunks related to the given chunk with proper semantic understanding
func (st *SymbolTable) FindRelatedChunks(chunk Chunk) []string {
	// Map of chunk IDs to their relation strength
	relatedChunks := make(map[string]RelationStrength)

	// Package name for context
	packageName := extractPackageName(chunk.FilePath)

	// 1. Find direct symbol reference relationships
	// -----------------------------------------------

	// Symbols defined in this chunk and referenced elsewhere
	for _, symbol := range chunk.Symbols {
		for _, ref := range st.References[symbol] {
			if ref.ChunkID != chunk.ID {
				relatedChunks[ref.ChunkID] = max(relatedChunks[ref.ChunkID], RelationStrong)
			}
		}
	}

	// Symbols referenced in this chunk but defined elsewhere
	for symbol := range st.Definitions {
		// Check if this symbol is referenced in the chunk
		isReferenced := false
		for _, ref := range st.References[symbol] {
			if ref.ChunkID == chunk.ID {
				isReferenced = true
				break
			}
		}

		if isReferenced {
			// Find all definitions of this symbol
			for _, def := range st.Definitions[symbol] {
				if def.ChunkID != chunk.ID {
					relatedChunks[def.ChunkID] = max(relatedChunks[def.ChunkID], RelationStrong)
				}
			}
		}
	}

	// 2. Find method-type relationships (specific to Go)
	// --------------------------------------------------

	// If this chunk defines methods, relate to the type definition
	for _, symbol := range chunk.Symbols {
		if strings.Contains(symbol, ".") {
			// This looks like a method, e.g. "Type.Method"
			parts := strings.Split(symbol, ".")
			if len(parts) == 2 {
				typeName := parts[0]
				// Find the type definition
				for _, def := range st.Definitions[typeName] {
					if def.ChunkID != chunk.ID {
						relatedChunks[def.ChunkID] = max(relatedChunks[def.ChunkID], RelationStrong)
					}
				}
			}
		}
	}

	// If this chunk defines a type, relate to its methods
	for _, symbol := range chunk.Symbols {
		methodPattern := symbol + "."
		for methodSymbol := range st.Definitions {
			if strings.HasPrefix(methodSymbol, methodPattern) {
				for _, methodDef := range st.Definitions[methodSymbol] {
					if methodDef.ChunkID != chunk.ID {
						relatedChunks[methodDef.ChunkID] = max(relatedChunks[methodDef.ChunkID], RelationStrong)
					}
				}
			}
		}
	}

	// 3. Find interface-implementation relationships
	// ---------------------------------------------
	// This is a simplified approach without full type checking

	// If this chunk defines an interface
	for _, symbol := range chunk.Symbols {
		// Look for definitions that seem to be interfaces (type name followed by "interface")
		for _, def := range st.Definitions[symbol] {
			if def.Type == "interface" && def.ChunkID == chunk.ID {
				// Find potential implementers by looking for method name patterns
				// For each chunk that contains method definitions...
				for methodSymbol := range st.Definitions {
					if strings.Contains(methodSymbol, ".") {
						methodName := strings.Split(methodSymbol, ".")[1]

						// Check if this method is mentioned in the interface
						if strings.Contains(chunk.Content, methodName) {
							for _, methodDef := range st.Definitions[methodSymbol] {
								if methodDef.ChunkID != chunk.ID {
									relatedChunks[methodDef.ChunkID] = max(relatedChunks[methodDef.ChunkID], RelationMedium)
								}
							}
						}
					}
				}
			}
		}
	}

	// 4. Find import relationships
	// ----------------------------

	for _, imp := range chunk.Imports {
		// Get the package name from the import path
		pkgName := extractPackageNameFromImport(imp)

		// Find all chunks that belong to this package
		for _, otherChunk := range st.AllChunks() {
			otherPkgName := extractPackageName(otherChunk.FilePath)
			if pkgName == otherPkgName && otherChunk.ID != chunk.ID {
				relatedChunks[otherChunk.ID] = max(relatedChunks[otherChunk.ID], RelationMedium)
			}
		}
	}

	// 5. Same-package relationship (weakest)
	// -------------------------------------

	if packageName != "" {
		for _, otherChunk := range st.AllChunks() {
			otherPkgName := extractPackageName(otherChunk.FilePath)
			if packageName == otherPkgName && otherChunk.ID != chunk.ID {
				// Only add if we don't already have a stronger relationship
				if _, exists := relatedChunks[otherChunk.ID]; !exists {
					relatedChunks[otherChunk.ID] = RelationWeak
				}
			}
		}
	}

	// 6. Sort by relationship strength and return top N
	// ------------------------------------------------

	// Convert to a slice of (ChunkID, Strength) pairs
	type relationPair struct {
		chunkID  string
		strength RelationStrength
	}

	pairs := make([]relationPair, 0, len(relatedChunks))
	for id, strength := range relatedChunks {
		pairs = append(pairs, relationPair{id, strength})
	}

	// Sort by strength (descending)
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].strength > pairs[j].strength
	})

	// Take top N (limit to 10 to avoid overwhelming the vector database)
	maxRelations := 10
	if len(pairs) > maxRelations {
		pairs = pairs[:maxRelations]
	}

	// Extract chunk IDs
	result := make([]string, len(pairs))
	for i, pair := range pairs {
		result[i] = pair.chunkID
	}

	return result
}

// Helper functions

// extractPackageName extracts the package name from a file path
func extractPackageName(filePath string) string {
	dir := filepath.Dir(filePath)
	return filepath.Base(dir)
}

// extractPackageNameFromImport extracts the package name from an import path
func extractPackageNameFromImport(importPath string) string {
	// For import paths like "github.com/user/repo/pkg/subpkg",
	// we want to extract "subpkg"
	return filepath.Base(importPath)
}

// max returns the maximum of two RelationStrength values
func max(a, b RelationStrength) RelationStrength {
	if a > b {
		return a
	}
	return b
}
