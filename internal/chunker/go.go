package chunker

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
)

// ChunkGoFile parses a Go file, chunking by package, struct, func, etc.
func ChunkGoFile(filepath string) ([]Chunk, error) {
	source, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filepath, source, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var chunks []Chunk

	// Optionally, chunk for entire package
	if file.Name != nil {
		// pkgName := file.Name.Name
		log.Printf("Chunking Go Package: %s\n", file.Name.Name)
		code := string(source)
		chunkID := makeChunkID(filepath, code)

		pkgChunk := Chunk{
			FilePath:  filepath,
			ChunkID:   chunkID, // stable if package contents + file path unchanged
			ChunkType: "package",
			Code:      code,
			StartLine: 1,
			EndLine:   countLines(source),
		}
		chunks = append(chunks, pkgChunk)
	}

	// For top-level declarations
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			c := extractFuncChunk(d, filepath, fset, source)
			chunks = append(chunks, c)
		case *ast.GenDecl:
			genChunks := extractGenDeclChunks(d, filepath, fset, source)
			chunks = append(chunks, genChunks...)
		}
	}

	return chunks, nil
}

func extractFuncChunk(fn *ast.FuncDecl, filePath string, fset *token.FileSet, source []byte) Chunk {
	start := fset.Position(fn.Pos())
	end := fset.Position(fn.End())
	code := extractSource(source, start.Offset, end.Offset)

	var chunkID, chunkType string
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		// method
		recv := nodeToString(fn.Recv.List[0].Type)
		chunkType = "method"
		chunkID = fmt.Sprintf("%s:%s:%s", chunkType, recv, fn.Name.Name)
	} else {
		chunkType = "function"
		chunkID = fmt.Sprintf("%s:%s", chunkType, fn.Name.Name)
	}

	// Add path+code-based hashing to chunkID if you want the ID to reflect content changes:
	hashedID := makeChunkID(filePath, code)
	chunkID = chunkID + ":" + hashedID[:8] // e.g. short prefix to disambiguate

	return Chunk{
		FilePath:  filePath,
		ChunkID:   chunkID,
		ChunkType: chunkType,
		StartLine: start.Line,
		EndLine:   end.Line,
		Code:      code,
	}
}

func extractGenDeclChunks(gd *ast.GenDecl, filePath string, fset *token.FileSet, source []byte) []Chunk {
	var results []Chunk
	start := fset.Position(gd.Pos())
	end := fset.Position(gd.End())
	code := extractSource(source, start.Offset, end.Offset)

	chunkType := gd.Tok.String() // "import", "var", "const", "type"
	baseID := fmt.Sprintf("%s_decl_%d_%d", chunkType, start.Line, end.Line)

	hashedID := makeChunkID(filePath, code)
	finalID := baseID + ":" + hashedID[:8]

	c := Chunk{
		FilePath:  filePath,
		ChunkID:   finalID,
		ChunkType: chunkType,
		StartLine: start.Line,
		EndLine:   end.Line,
		Code:      code,
	}
	results = append(results, c)
	return results
}
