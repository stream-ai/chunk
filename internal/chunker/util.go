package chunker

import (
	"fmt"
	"go/ast"
	"strings"
)

// -----------------------------
// 7. Internal Utility Functions
// -----------------------------

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
	// E.g., "*MyStruct", "MyStruct"
	return strings.TrimSpace(fmt.Sprintf("%v", expr))
}

func countLines(data []byte) int {
	return strings.Count(string(data), "\n") + 1
}
