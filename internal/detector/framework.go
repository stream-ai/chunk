package detector

import (
	"path/filepath"
	"strings"
)

// DefaultFrameworkDetector implements the FrameworkDetector interface
type DefaultFrameworkDetector struct {
	// frameworkPatterns maps frameworks to byte patterns to look for in the content
	frameworkPatterns map[string][]string
}

// NewDefaultFrameworkDetector creates a new DefaultFrameworkDetector
func NewDefaultFrameworkDetector() *DefaultFrameworkDetector {
	detector := &DefaultFrameworkDetector{
		frameworkPatterns: make(map[string][]string),
	}

	// Initialize pattern mappings
	detector.registerPatterns()

	return detector
}

// registerPatterns sets up the framework detection patterns
func (d *DefaultFrameworkDetector) registerPatterns() {
	// JavaScript/TypeScript frameworks
	d.frameworkPatterns["react"] = []string{
		"import React",
		"from 'react'",
		"from \"react\"",
		"React.Component",
		"extends Component",
		"useState",
		"useEffect",
		"createContext",
	}

	d.frameworkPatterns["angular"] = []string{
		"@angular/core",
		"@Component",
		"@NgModule",
		"@Injectable",
		"moduleId",
		"platformBrowserDynamic",
	}

	d.frameworkPatterns["vue"] = []string{
		"import Vue",
		"from 'vue'",
		"from \"vue\"",
		"new Vue",
		"createApp",
		"<template>",
		"Vue.component",
		"defineComponent",
	}

	d.frameworkPatterns["svelte"] = []string{
		"<script>",
		"<style>",
		"export let",
		"svelte:",
	}

	d.frameworkPatterns["nextjs"] = []string{
		"import { NextPage }",
		"GetServerSideProps",
		"GetStaticPaths",
		"GetStaticProps",
		"next/router",
		"next/link",
		"NextApiRequest",
		"NextApiResponse",
	}

	// Mobile frameworks
	d.frameworkPatterns["flutter"] = []string{
		"package:flutter",
		"extends StatelessWidget",
		"extends StatefulWidget",
		"BuildContext",
		"MaterialApp",
	}

	d.frameworkPatterns["reactnative"] = []string{
		"from 'react-native'",
		"import { View, Text } from 'react-native'",
		"StyleSheet.create",
		"AppRegistry",
	}
}

// DetectFramework implements the FrameworkDetector interface
func (d *DefaultFrameworkDetector) DetectFramework(filePath string, content []byte, language string) string {
	// Check file extension first
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".jsx" || ext == ".tsx" {
		return "react"
	}
	if ext == ".vue" {
		return "vue"
	}
	if ext == ".svelte" {
		return "svelte"
	}

	// No need to check for frameworks in non-JS/TS files
	if language != "javascript" && language != "typescript" && language != "jsx" && language != "tsx" && language != "dart" {
		return ""
	}

	// Check config files
	basename := strings.ToLower(filepath.Base(filePath))
	if basename == "package.json" {
		contentStr := string(content)

		// Check for dependencies in package.json
		if strings.Contains(contentStr, "\"react\"") && !strings.Contains(contentStr, "\"react-native\"") {
			return "react"
		}
		if strings.Contains(contentStr, "\"@angular/core\"") {
			return "angular"
		}
		if strings.Contains(contentStr, "\"vue\"") {
			return "vue"
		}
		if strings.Contains(contentStr, "\"svelte\"") {
			return "svelte"
		}
		if strings.Contains(contentStr, "\"next\"") {
			return "nextjs"
		}
		if strings.Contains(contentStr, "\"react-native\"") {
			return "reactnative"
		}
	}

	// Check for patterns in the content
	contentStr := string(content)
	for framework, patterns := range d.frameworkPatterns {
		for _, pattern := range patterns {
			if strings.Contains(contentStr, pattern) {
				return framework
			}
		}
	}

	return ""
}
