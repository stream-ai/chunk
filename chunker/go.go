package chunker

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
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

	// 1) Add a chunk for the entire package if you want
	if file.Name != nil {
		pkgName := file.Name.Name
		pkgChunk := Chunk{
			FilePath:  filepath,
			ChunkID:   fmt.Sprintf("package_%s", pkgName),
			ChunkType: "package",
			Code:      extractSource(source, 0, len(source)),
			StartLine: 1,
			EndLine:   countLines(source),
		}
		chunks = append(chunks, pkgChunk)
	}

	// 2) Top-level declarations
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			c := extractFuncChunk(d, filepath, fset, source)
			chunks = append(chunks, c)
		case *ast.GenDecl:
			// Could handle var/const/import/type blocks
			genChunks := extractGenDeclChunks(d, filepath, fset, source)
			chunks = append(chunks, genChunks...)
		}
	}

	return chunks, nil
}

func extractFuncChunk(fn *ast.FuncDecl, filepath string, fset *token.FileSet, source []byte) Chunk {
	start := fset.Position(fn.Pos())
	end := fset.Position(fn.End())
	chunkText := extractSource(source, start.Offset, end.Offset)

	var chunkID, chunkType string
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		// method
		recv := nodeToString(fn.Recv.List[0].Type)
		chunkID = fmt.Sprintf("method_%s_%s", recv, fn.Name.Name)
		chunkType = "method"
	} else {
		chunkID = fmt.Sprintf("func_%s", fn.Name.Name)
		chunkType = "function"
	}

	return Chunk{
		FilePath:  filepath,
		ChunkID:   chunkID,
		ChunkType: chunkType,
		StartLine: start.Line,
		EndLine:   end.Line,
		Code:      chunkText,
	}
}

func extractGenDeclChunks(gd *ast.GenDecl, filepath string, fset *token.FileSet, source []byte) []Chunk {
	var results []Chunk
	start := fset.Position(gd.Pos())
	end := fset.Position(gd.End())
	chunkText := extractSource(source, start.Offset, end.Offset)

	// Could refine further for each spec
	chunkType := gd.Tok.String() // "import", "var", "const", "type"
	c := Chunk{
		FilePath:  filepath,
		ChunkID:   fmt.Sprintf("%s_decl_%d_%d", chunkType, start.Line, end.Line),
		ChunkType: chunkType,
		StartLine: start.Line,
		EndLine:   end.Line,
		Code:      chunkText,
	}
	results = append(results, c)
	return results
}

func extractSource(source []byte, start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(source) {
		end = len(source)
	}
	return string(source[start:end])
}

func nodeToString(expr ast.Expr) string {
	// E.g., "*MyStruct", "MyStruct", etc.
	return strings.TrimSpace(fmt.Sprintf("%v", expr))
}

func countLines(data []byte) int {
	return strings.Count(string(data), "\n") + 1
}
