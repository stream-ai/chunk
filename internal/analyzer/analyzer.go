package analyzer

// SymbolExtractor defines the contract for symbol extraction
type SymbolExtractor interface {
	// ExtractSymbols extracts symbols from content
	ExtractSymbols(content []byte, language string, framework string) []string

	// ExtractImports extracts imports from content
	ExtractImports(content []byte, language string) []string
}

// Analyzer combines multiple analysis functions
type Analyzer struct {
	SymbolExtractor SymbolExtractor
}

// NewAnalyzer creates a new analyzer
func NewAnalyzer(symbolExtractor SymbolExtractor) *Analyzer {
	return &Analyzer{
		SymbolExtractor: symbolExtractor,
	}
}
