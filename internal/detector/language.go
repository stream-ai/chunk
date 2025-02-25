package detector

import (
	"bytes"
	"path/filepath"
	"strings"
)

// DefaultLanguageDetector implements the LanguageDetector interface
type DefaultLanguageDetector struct {
	// extensionMap maps file extensions to languages
	extensionMap map[string]string

	// specialFilesMap maps specific filenames to languages
	specialFilesMap map[string]string
}

// NewDefaultLanguageDetector creates a new DefaultLanguageDetector
func NewDefaultLanguageDetector() *DefaultLanguageDetector {
	detector := &DefaultLanguageDetector{
		extensionMap:    make(map[string]string),
		specialFilesMap: make(map[string]string),
	}

	// Initialize extension mappings
	detector.registerExtensions()
	detector.registerSpecialFiles()

	return detector
}

// registerExtensions sets up the extension to language mappings
func (d *DefaultLanguageDetector) registerExtensions() {
	// Programming languages
	d.extensionMap[".go"] = "go"
	d.extensionMap[".js"] = "javascript"
	d.extensionMap[".jsx"] = "jsx"
	d.extensionMap[".ts"] = "typescript"
	d.extensionMap[".tsx"] = "tsx"
	d.extensionMap[".py"] = "python"
	d.extensionMap[".rb"] = "ruby"
	d.extensionMap[".java"] = "java"
	d.extensionMap[".kt"] = "kotlin"
	d.extensionMap[".kts"] = "kotlin"
	d.extensionMap[".swift"] = "swift"
	d.extensionMap[".c"] = "c"
	d.extensionMap[".h"] = "c"
	d.extensionMap[".cpp"] = "cpp"
	d.extensionMap[".cc"] = "cpp"
	d.extensionMap[".cxx"] = "cpp"
	d.extensionMap[".hpp"] = "cpp"
	d.extensionMap[".cs"] = "csharp"
	d.extensionMap[".rs"] = "rust"
	d.extensionMap[".php"] = "php"
	d.extensionMap[".dart"] = "dart"

	// Shell scripts and config files
	d.extensionMap[".sh"] = "shell"
	d.extensionMap[".bash"] = "shell"
	d.extensionMap[".zsh"] = "shell"
	d.extensionMap[".fish"] = "shell"
	d.extensionMap[".ps1"] = "powershell"
	d.extensionMap[".bat"] = "batch"
	d.extensionMap[".cmd"] = "batch"

	// Web/markup languages
	d.extensionMap[".html"] = "html"
	d.extensionMap[".css"] = "css"
	d.extensionMap[".scss"] = "scss"
	d.extensionMap[".less"] = "less"
	d.extensionMap[".vue"] = "vue"
	d.extensionMap[".svelte"] = "svelte"

	// Data/config formats
	d.extensionMap[".json"] = "json"
	d.extensionMap[".yaml"] = "yaml"
	d.extensionMap[".yml"] = "yaml"
	d.extensionMap[".toml"] = "toml"
	d.extensionMap[".md"] = "markdown"
	d.extensionMap[".markdown"] = "markdown"
	d.extensionMap[".ini"] = "ini"
	d.extensionMap[".conf"] = "conf"
	d.extensionMap[".env"] = "env"
}

func (d *DefaultLanguageDetector) registerSpecialFiles() {
	d.specialFilesMap["dockerfile"] = "dockerfile"
	d.specialFilesMap["makefile"] = "makefile"
	d.specialFilesMap["jenkinsfile"] = "jenkinsfile"
	d.specialFilesMap["gemfile"] = "ruby"
	d.specialFilesMap["rakefile"] = "ruby"
	d.specialFilesMap["cmakelists.txt"] = "cmake"
	d.specialFilesMap[".gitignore"] = "gitignore"
	d.specialFilesMap[".dockerignore"] = "dockerignore"
	d.specialFilesMap[".bashrc"] = "shell"
	d.specialFilesMap[".zshrc"] = "shell"
	d.specialFilesMap[".bash_profile"] = "shell"
	d.specialFilesMap[".profile"] = "shell"
}

// DetectLanguage implements the LanguageDetector interface
func (d *DefaultLanguageDetector) DetectLanguage(filePath string, content []byte) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	baseName := strings.ToLower(filepath.Base(filePath))

	// Special case for Dockerfile (which has no extension)
	if baseName == "dockerfile" {
		return "dockerfile"
	}

	// Check if the file starts with a shebang for shell scripts
	if len(content) > 2 && content[0] == '#' && content[1] == '!' {
		firstLine := string(bytes.SplitN(content, []byte("\n"), 2)[0])
		if strings.Contains(firstLine, "/bin/bash") || strings.Contains(firstLine, "/bin/sh") {
			return "shell"
		}
		if strings.Contains(firstLine, "/bin/zsh") {
			return "shell"
		}
		if strings.Contains(firstLine, "python") {
			return "python"
		}
		if strings.Contains(firstLine, "ruby") || strings.Contains(firstLine, "ruby") {
			return "ruby"
		}
		if strings.Contains(firstLine, "node") || strings.Contains(firstLine, "nodejs") {
			return "javascript"
		}
		// If we found a shebang but don't recognize it, mark as shell anyway
		return "shell"
	}

	// Check extension mappings
	if language, ok := d.extensionMap[ext]; ok {
		return language
	}

	// Check special filenames
	if language, ok := d.specialFilesMap[baseName]; ok {
		return language
	}

	// Check for special filenames with extensions
	for specialFile, language := range d.specialFilesMap {
		if strings.HasPrefix(baseName, specialFile+".") {
			return language
		}
	}

	// Default fallback
	return "unknown"
}
