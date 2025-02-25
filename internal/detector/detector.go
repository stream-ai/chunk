package detector

// LanguageDetector defines the contract for language detection
type LanguageDetector interface {
	// DetectLanguage determines the programming language of a file
	DetectLanguage(filePath string, content []byte) string
}

// FrameworkDetector defines the contract for framework detection
type FrameworkDetector interface {
	// DetectFramework determines the framework used in a file
	DetectFramework(filePath string, content []byte, language string) string
}
