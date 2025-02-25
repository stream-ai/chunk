package util

import (
	"path/filepath"
	"strings"
)

// IsBinaryFile determines if a file is likely binary based on extension
func IsBinaryFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	binaryExts := map[string]bool{
		".exe":   true,
		".dll":   true,
		".so":    true,
		".dylib": true,
		".bin":   true,
		".obj":   true,
		".o":     true,
		".a":     true,
		".lib":   true,
		".png":   true,
		".jpg":   true,
		".jpeg":  true,
		".gif":   true,
		".bmp":   true,
		".tiff":  true,
		".ico":   true,
		".zip":   true,
		".tar":   true,
		".gz":    true,
		".bz2":   true,
		".7z":    true,
		".rar":   true,
		".pdf":   true,
		".doc":   true,
		".docx":  true,
		".xls":   true,
		".xlsx":  true,
		".ppt":   true,
		".pptx":  true,
	}

	return binaryExts[ext]
}

// ShouldSkipDirectory checks if a directory should be skipped
func ShouldSkipDirectory(dirName string) bool {
	dirsToSkip := map[string]bool{
		".git":         true,
		"node_modules": true,
		"vendor":       true,
		"dist":         true,
		"build":        true,
		"target":       true,
		"cdk.out":      true,
		".next":        true,
		".angular":     true,
	}

	return dirsToSkip[dirName]
}
