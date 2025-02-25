// internal/chunker/go.go
package chunker

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/stream-ai/chunk/internal/model"
	"github.com/stream-ai/chunk/pkg/util"
)

// GoChunker implements the Chunker interface for Go code using Go's AST parser
type GoChunker struct{}

// NewGoChunker creates a new Go code chunker
func NewGoChunker() *GoChunker {
	return &GoChunker{}
}

// Language returns the language this chunker supports
func (c *GoChunker) Language() string {
	return "go"
}

// CanHandle checks if this chunker can handle the given file
func (c *GoChunker) CanHandle(filePath string, language string, framework string) bool {
	return language == "go"
}

// Chunk splits the Go file content into chunks using the Go AST parser
func (c *GoChunker) Chunk(filePath string, content []byte, symbolTable *model.SymbolTable, options ChunkingOptions) ([]model.Chunk, error) {
	// Setup file set and parse the file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		// If parsing fails, fall back to line-based chunking
		return c.fallbackChunking(filePath, content, symbolTable, options)
	}

	// Extract package name
	packageName := file.Name.Name

	var chunks []model.Chunk

	// Process imports
	imports := c.extractImports(file)

	// First pass: Create chunks for top-level declarations
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			// Function or method declaration
			chunk := c.processFuncDecl(fset, d, filePath, content, packageName, imports, symbolTable)
			chunks = append(chunks, chunk)

		case *ast.GenDecl:
			// Type, const, var declarations
			if d.Tok == token.TYPE || d.Tok == token.CONST || d.Tok == token.VAR {
				chunk := c.processGenDecl(fset, d, filePath, content, packageName, imports, symbolTable)
				if chunk.Content != "" {
					chunks = append(chunks, chunk)
				}
			}
		}
	}

	// Handle imports section if not already included in a chunk
	importsChunk := c.processImportsSection(fset, file, filePath, content, packageName, symbolTable)
	if importsChunk.Content != "" {
		chunks = append(chunks, importsChunk)
	}

	// Handle package declaration and comments at the top of the file
	packageChunk := c.processPackageSection(fset, file, filePath, content, packageName, symbolTable)
	if packageChunk.Content != "" {
		chunks = append(chunks, packageChunk)
	}

	// Second pass: Collect references
	c.collectReferences(fset, file, chunks, symbolTable)

	return chunks, nil
}

// processFuncDecl creates a chunk for a function declaration
func (c *GoChunker) processFuncDecl(fset *token.FileSet, decl *ast.FuncDecl, filePath string, content []byte, packageName string, imports []string, symbolTable *model.SymbolTable) model.Chunk {
	// Get position information
	startPos := fset.Position(decl.Pos())
	endPos := fset.Position(decl.End())

	// Extract function content
	funcContent := string(content[startPos.Offset:endPos.Offset])

	// Generate ID for the chunk
	chunkID := util.GenerateID(filePath, funcContent)

	// Determine symbol name
	var symbolName string
	if decl.Recv == nil {
		// Regular function
		symbolName = decl.Name.Name
	} else {
		// Method
		// Get the receiver type
		receiver := decl.Recv.List[0].Type
		var receiverType string

		// Handle pointer receivers like (*T)
		if starExpr, ok := receiver.(*ast.StarExpr); ok {
			if ident, ok := starExpr.X.(*ast.Ident); ok {
				receiverType = ident.Name
			}
		} else if ident, ok := receiver.(*ast.Ident); ok {
			receiverType = ident.Name
		}

		if receiverType != "" {
			symbolName = receiverType + "." + decl.Name.Name
		} else {
			symbolName = decl.Name.Name
		}
	}

	// Create chunk
	chunk := model.Chunk{
		ID:         chunkID,
		FilePath:   filePath,
		StartLine:  startPos.Line,
		EndLine:    endPos.Line,
		Content:    funcContent,
		Language:   "go",
		Symbols:    []string{symbolName},
		Imports:    imports,
		TokenCount: util.EstimateTokenCount(funcContent),
	}

	// Add symbol definition to symbol table
	symbolTable.AddDefinition(symbolName, model.SymbolDefinition{
		Name:      symbolName,
		ChunkID:   chunkID,
		FilePath:  filePath,
		StartLine: startPos.Line,
		EndLine:   endPos.Line,
		Type:      "function",
	})

	// For methods, also record the receiver type to establish relationships
	if decl.Recv != nil {
		receiver := decl.Recv.List[0].Type
		var receiverType string

		// Handle pointer receivers like (*T)
		if starExpr, ok := receiver.(*ast.StarExpr); ok {
			if ident, ok := starExpr.X.(*ast.Ident); ok {
				receiverType = ident.Name
			}
		} else if ident, ok := receiver.(*ast.Ident); ok {
			receiverType = ident.Name
		}

		if receiverType != "" {
			// Add a reference to the type
			symbolTable.AddReference(receiverType, model.SymbolReference{
				Name:     receiverType,
				ChunkID:  chunkID,
				FilePath: filePath,
				Line:     startPos.Line,
			})
		}
	}

	return chunk
}

// processGenDecl creates a chunk for a type, const, or var declaration
func (c *GoChunker) processGenDecl(fset *token.FileSet, decl *ast.GenDecl, filePath string, content []byte, packageName string, imports []string, symbolTable *model.SymbolTable) model.Chunk {
	// Get position information
	startPos := fset.Position(decl.Pos())
	endPos := fset.Position(decl.End())

	// Extract declaration content
	declContent := string(content[startPos.Offset:endPos.Offset])

	// Generate ID for the chunk
	chunkID := util.GenerateID(filePath, declContent)

	// Extract symbols based on declaration type
	var symbols []string
	var declType string

	switch decl.Tok {
	case token.TYPE:
		declType = "type"
		for _, spec := range decl.Specs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				symbols = append(symbols, typeSpec.Name.Name)
			}
		}
	case token.CONST:
		declType = "const"
		for _, spec := range decl.Specs {
			if valueSpec, ok := spec.(*ast.ValueSpec); ok {
				for _, name := range valueSpec.Names {
					symbols = append(symbols, name.Name)
				}
			}
		}
	case token.VAR:
		declType = "var"
		for _, spec := range decl.Specs {
			if valueSpec, ok := spec.(*ast.ValueSpec); ok {
				for _, name := range valueSpec.Names {
					symbols = append(symbols, name.Name)
				}
			}
		}
	}

	// Skip empty declarations
	if len(symbols) == 0 || declContent == "" {
		return model.Chunk{}
	}

	// Create chunk
	chunk := model.Chunk{
		ID:         chunkID,
		FilePath:   filePath,
		StartLine:  startPos.Line,
		EndLine:    endPos.Line,
		Content:    declContent,
		Language:   "go",
		Symbols:    symbols,
		Imports:    imports,
		TokenCount: util.EstimateTokenCount(declContent),
	}

	// Add symbol definitions to symbol table
	for _, symbol := range symbols {
		symbolTable.AddDefinition(symbol, model.SymbolDefinition{
			Name:      symbol,
			ChunkID:   chunkID,
			FilePath:  filePath,
			StartLine: startPos.Line,
			EndLine:   endPos.Line,
			Type:      declType,
		})
	}

	return chunk
}

// processImportsSection creates a chunk for the imports section
func (c *GoChunker) processImportsSection(fset *token.FileSet, file *ast.File, filePath string, content []byte, packageName string, symbolTable *model.SymbolTable) model.Chunk {
	// Find import declarations
	var importDecl *ast.GenDecl
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			importDecl = genDecl
			break
		}
	}

	if importDecl == nil {
		return model.Chunk{}
	}

	// Get position information
	startPos := fset.Position(importDecl.Pos())
	endPos := fset.Position(importDecl.End())

	// Extract imports content
	importsContent := string(content[startPos.Offset:endPos.Offset])

	// Generate ID for the chunk
	chunkID := util.GenerateID(filePath, importsContent)

	// Create chunk
	chunk := model.Chunk{
		ID:         chunkID,
		FilePath:   filePath,
		StartLine:  startPos.Line,
		EndLine:    endPos.Line,
		Content:    importsContent,
		Language:   "go",
		Symbols:    []string{"imports"},
		TokenCount: util.EstimateTokenCount(importsContent),
	}

	return chunk
}

// processPackageSection creates a chunk for the package declaration and top comments
func (c *GoChunker) processPackageSection(fset *token.FileSet, file *ast.File, filePath string, content []byte, packageName string, symbolTable *model.SymbolTable) model.Chunk {
	// Get position information for package declaration
	startPos := fset.Position(file.Package)
	endPos := fset.Position(file.Name.End())

	// Check for comments before package declaration
	var commentStartPos token.Position
	if file.Doc != nil && len(file.Doc.List) > 0 {
		commentStartPos = fset.Position(file.Doc.Pos())
		startPos = commentStartPos
	}

	// Extract package content
	packageContent := string(content[startPos.Offset:endPos.Offset])

	// Generate ID for the chunk
	chunkID := util.GenerateID(filePath, packageContent)

	// Create chunk
	chunk := model.Chunk{
		ID:         chunkID,
		FilePath:   filePath,
		StartLine:  startPos.Line,
		EndLine:    endPos.Line,
		Content:    packageContent,
		Language:   "go",
		Symbols:    []string{packageName},
		TokenCount: util.EstimateTokenCount(packageContent),
	}

	return chunk
}

// extractImports extracts imports from the AST
func (c *GoChunker) extractImports(file *ast.File) []string {
	var imports []string

	for _, imp := range file.Imports {
		// Remove quotes from import path
		path := strings.Trim(imp.Path.Value, "\"")
		imports = append(imports, path)
	}

	return imports
}

// collectReferences processes the AST to find references to symbols
func (c *GoChunker) collectReferences(fset *token.FileSet, file *ast.File, chunks []model.Chunk, symbolTable *model.SymbolTable) {
	// Visitor to find identifier references
	visitor := &referenceVisitor{
		fset:        fset,
		symbolTable: symbolTable,
		chunks:      chunks,
		filePath:    file.Name.Name,
	}

	// Walk the AST to find references
	ast.Walk(visitor, file)
}

// referenceVisitor implements the ast.Visitor interface to find references
type referenceVisitor struct {
	fset         *token.FileSet
	symbolTable  *model.SymbolTable
	chunks       []model.Chunk
	filePath     string
	currentChunk string
}

// Visit implements the ast.Visitor interface
func (v *referenceVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	// Update current chunk based on position
	pos := v.fset.Position(node.Pos())
	for _, chunk := range v.chunks {
		if pos.Line >= chunk.StartLine && pos.Line <= chunk.EndLine {
			v.currentChunk = chunk.ID
			break
		}
	}

	// Check for identifier references
	if ident, ok := node.(*ast.Ident); ok {
		name := ident.Name

		// Skip basic types, keywords, etc.
		if isBasicType(name) || isKeyword(name) {
			return v
		}

		// Check if this identifier is a reference to a known symbol
		if _, exists := v.symbolTable.Definitions[name]; exists {
			// Add reference if we're in a valid chunk
			if v.currentChunk != "" {
				v.symbolTable.AddReference(name, model.SymbolReference{
					Name:     name,
					ChunkID:  v.currentChunk,
					FilePath: v.filePath,
					Line:     pos.Line,
				})
			}
		}
	}

	return v
}

// isBasicType checks if a name is a Go basic type
func isBasicType(name string) bool {
	basicTypes := map[string]bool{
		"bool":       true,
		"byte":       true,
		"complex128": true,
		"complex64":  true,
		"error":      true,
		"float32":    true,
		"float64":    true,
		"int":        true,
		"int16":      true,
		"int32":      true,
		"int64":      true,
		"int8":       true,
		"rune":       true,
		"string":     true,
		"uint":       true,
		"uint16":     true,
		"uint32":     true,
		"uint64":     true,
		"uint8":      true,
		"uintptr":    true,
	}
	return basicTypes[name]
}

// isKeyword checks if a name is a Go keyword
func isKeyword(name string) bool {
	keywords := map[string]bool{
		"break":       true,
		"case":        true,
		"chan":        true,
		"const":       true,
		"continue":    true,
		"default":     true,
		"defer":       true,
		"else":        true,
		"fallthrough": true,
		"for":         true,
		"func":        true,
		"go":          true,
		"goto":        true,
		"if":          true,
		"import":      true,
		"interface":   true,
		"map":         true,
		"package":     true,
		"range":       true,
		"return":      true,
		"select":      true,
		"struct":      true,
		"switch":      true,
		"type":        true,
		"var":         true,
	}
	return keywords[name]
}

// fallbackChunking provides a line-based chunking as a fallback
func (c *GoChunker) fallbackChunking(filePath string, content []byte, symbolTable *model.SymbolTable, options ChunkingOptions) ([]model.Chunk, error) {
	lines := strings.Split(string(content), "\n")
	var chunks []model.Chunk

	// Simple line-based chunking
	for i := 0; i < len(lines); i += options.MaxChunkSize {
		end := i + options.MaxChunkSize
		if end > len(lines) {
			end = len(lines)
		}

		chunkLines := lines[i:end]
		chunkContent := strings.Join(chunkLines, "\n")

		// Generate ID
		chunkID := util.GenerateID(filePath, chunkContent)

		chunk := model.Chunk{
			ID:         chunkID,
			FilePath:   filePath,
			StartLine:  i + 1,
			EndLine:    end,
			Content:    chunkContent,
			Language:   "go",
			TokenCount: util.EstimateTokenCount(chunkContent),
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// FindRelatedChunks finds chunks related to the given chunk
func (c *GoChunker) FindRelatedChunks(chunk model.Chunk, symbolTable *model.SymbolTable) []string {
	relatedChunks := make(map[string]bool)

	// For each symbol in this chunk
	for _, symbol := range chunk.Symbols {
		// Find where this symbol is referenced
		for _, ref := range symbolTable.References[symbol] {
			if ref.ChunkID != chunk.ID {
				relatedChunks[ref.ChunkID] = true
			}
		}
	}

	// For each symbol referenced in this chunk
	for symbol, refs := range symbolTable.References {
		for _, ref := range refs {
			if ref.ChunkID == chunk.ID {
				// Find where this symbol is defined
				for _, def := range symbolTable.Definitions[symbol] {
					if def.ChunkID != chunk.ID {
						relatedChunks[def.ChunkID] = true
					}
				}
			}
		}
	}

	// Convert map to slice
	var result []string
	for id := range relatedChunks {
		result = append(result, id)
	}

	return result
}
