package core

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"parsec/utils"

	"github.com/charmbracelet/glamour"
)

// FileSummary contains parsed information about a source file
type FileSummary struct {
	Path          string
	Language      string
	LineCount     int
	FunctionCount int
	Functions     []string
	Imports       []string
	Types         []string
	Structs       []string
	Error         string

	// Additional fields for non-code files
	Headers         []string // For markdown headers
	Links           []string // For markdown links
	ConfigKeys      []string // For config file keys
	FileSize        int64    // File size in bytes
	IsExecutable    bool     // Whether file is executable
	ExecutableHelp  string   // Help text from executable
	Content         []string // First few lines for text files
	RenderedContent string   // Glamour-rendered markdown or formatted content
	IsRendered      bool     // Whether content has been rendered with glamour
}

// LanguageConfig holds regex patterns for different programming languages
type LanguageConfig struct {
	FunctionPattern *regexp.Regexp
	ImportPattern   *regexp.Regexp
	TypePattern     *regexp.Regexp
	StructPattern   *regexp.Regexp
}

// Language patterns for different file types
var languageConfigs = map[string]LanguageConfig{
	".go": {
		FunctionPattern: regexp.MustCompile(`^func\s+(\w+)`),
		ImportPattern:   regexp.MustCompile(`^import\s+["']([^"']+)["']|^\s+["']([^"']+)["']`),
		TypePattern:     regexp.MustCompile(`^type\s+(\w+)\s+`),
		StructPattern:   regexp.MustCompile(`^type\s+(\w+)\s+struct`),
	},
	".py": {
		FunctionPattern: regexp.MustCompile(`^def\s+(\w+)`),
		ImportPattern:   regexp.MustCompile(`^(?:from\s+(\S+)\s+)?import\s+(.+)`),
		TypePattern:     regexp.MustCompile(`^class\s+(\w+)`),
		StructPattern:   regexp.MustCompile(`^class\s+(\w+)`),
	},
	".js": {
		FunctionPattern: regexp.MustCompile(`^(?:function\s+(\w+)|const\s+(\w+)\s*=.*=>|(\w+)\s*:\s*function)`),
		ImportPattern:   regexp.MustCompile(`^import.*from\s+['"]([^'"]+)['"]|^const\s+.*=\s+require\(['"]([^'"]+)['"]\)`),
		TypePattern:     regexp.MustCompile(`^(?:class\s+(\w+)|interface\s+(\w+))`),
		StructPattern:   regexp.MustCompile(`^(?:class\s+(\w+)|interface\s+(\w+))`),
	},
	".ts": {
		FunctionPattern: regexp.MustCompile(`^(?:function\s+(\w+)|const\s+(\w+)\s*=.*=>|(\w+)\s*:\s*function)`),
		ImportPattern:   regexp.MustCompile(`^import.*from\s+['"]([^'"]+)['"]`),
		TypePattern:     regexp.MustCompile(`^(?:class\s+(\w+)|interface\s+(\w+)|type\s+(\w+))`),
		StructPattern:   regexp.MustCompile(`^(?:class\s+(\w+)|interface\s+(\w+))`),
	},
	".rs": {
		FunctionPattern: regexp.MustCompile(`^(?:pub\s+)?fn\s+(\w+)`),
		ImportPattern:   regexp.MustCompile(`^use\s+([^;]+);`),
		TypePattern:     regexp.MustCompile(`^(?:pub\s+)?(?:struct\s+(\w+)|enum\s+(\w+)|type\s+(\w+))`),
		StructPattern:   regexp.MustCompile(`^(?:pub\s+)?struct\s+(\w+)`),
	},
	".cpp": {
		FunctionPattern: regexp.MustCompile(`^(?:[a-zA-Z_][a-zA-Z0-9_]*\s+)?(?:[a-zA-Z_][a-zA-Z0-9_]*::)?\w+\s+(\w+)\s*\(`),
		ImportPattern:   regexp.MustCompile(`^#include\s+[<"]([^>"]+)[>"]`),
		TypePattern:     regexp.MustCompile(`^(?:class|struct|enum|union)\s+(\w+)`),
		StructPattern:   regexp.MustCompile(`^(?:class|struct)\s+(\w+)`),
	},
	".cc": {
		FunctionPattern: regexp.MustCompile(`^(?:[a-zA-Z_][a-zA-Z0-9_]*\s+)?(?:[a-zA-Z_][a-zA-Z0-9_]*::)?\w+\s+(\w+)\s*\(`),
		ImportPattern:   regexp.MustCompile(`^#include\s+[<"]([^>"]+)[>"]`),
		TypePattern:     regexp.MustCompile(`^(?:class|struct|enum|union)\s+(\w+)`),
		StructPattern:   regexp.MustCompile(`^(?:class|struct)\s+(\w+)`),
	},
}

// Summarizer handles file analysis and summary generation
type Summarizer struct {
	basePath string
}

// NewSummarizer creates a new file summarizer
func NewSummarizer(basePath string) *Summarizer {
	return &Summarizer{basePath: basePath}
}

// SummarizeFile analyzes a file and returns its summary
func (s *Summarizer) SummarizeFile(filePath string) FileSummary {
	summary := FileSummary{
		Path:       filePath,
		Language:   getLanguage(filePath),
		Functions:  make([]string, 0),
		Imports:    make([]string, 0),
		Types:      make([]string, 0),
		Structs:    make([]string, 0),
		Headers:    make([]string, 0),
		Links:      make([]string, 0),
		ConfigKeys: make([]string, 0),
		Content:    make([]string, 0),
	}

	// Construct full path
	fullPath := filepath.Join(s.basePath, filePath)

	// Get file info
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		summary.Error = fmt.Sprintf("Error getting file info: %v", err)
		return summary
	}
	summary.FileSize = fileInfo.Size()

	// Check if file is executable
	summary.IsExecutable = utils.IsExecutableFile(fullPath)
	if summary.IsExecutable {
		summary.ExecutableHelp = s.getExecutableHelp(fullPath)
		return summary
	}

	// Get file extension and determine parsing strategy
	ext := strings.ToLower(filepath.Ext(filePath))

	// Handle different file types
	switch ext {
	case ".md", ".markdown":
		return s.parseMarkdown(fullPath, summary)
	case ".json":
		return s.parseJSON(fullPath, summary)
	case ".yaml", ".yml":
		return s.parseYAML(fullPath, summary)
	case ".ini", ".cfg", ".conf":
		return s.parseINI(fullPath, summary)
	case ".env":
		return s.parseEnv(fullPath, summary)
	case ".txt", ".log", ".rst", ".xml", ".csv":
		return s.parseTextFile(fullPath, summary)
	default:
		// Try to parse as source code using language-specific parsers
		switch ext {
		case ".go":
			return s.parseGoFile(fullPath, summary)
		case ".py":
			return s.parsePythonFile(fullPath, summary)
		case ".js", ".jsx", ".ts", ".tsx":
			return s.parseJavaScriptFile(fullPath, summary)
		case ".rs":
			return s.parseRustFile(fullPath, summary)
		case ".cpp", ".cc":
			return s.parseCppFile(fullPath, summary)
		default:
			// Try to parse as source code using generic parser
			config, exists := languageConfigs[ext]
			if !exists {
				// If not a known source file, treat as text
				return s.parseTextFile(fullPath, summary)
			}
			return s.parseSourceCode(fullPath, summary, config)
		}
	}
}

// parseSourceCode handles traditional programming language files
func (s *Summarizer) parseSourceCode(fullPath string, summary FileSummary, config LanguageConfig) FileSummary {
	// TODO: Refactor into language-specific parsing functions
	// This is a temporary placeholder while we implement the refactoring
	// The current implementation will be replaced with language-specific parsers

	file, err := os.Open(fullPath)
	if err != nil {
		summary.Error = fmt.Sprintf("Error opening file: %v", err)
		return summary
	}
	defer file.Close()

	// Parse file line by line
	scanner := bufio.NewScanner(file)
	lineCount := 0
	inMultiLineComment := false

	for scanner.Scan() {
		lineCount++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments (basic)
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle multi-line comments (Go style)
		if strings.Contains(line, "/*") {
			inMultiLineComment = true
		}
		if inMultiLineComment {
			if strings.Contains(line, "*/") {
				inMultiLineComment = false
			}
			continue
		}

		// Extract functions
		if matches := config.FunctionPattern.FindStringSubmatch(line); matches != nil {
			for _, match := range matches[1:] {
				if match != "" {
					summary.Functions = append(summary.Functions, match)
					summary.FunctionCount++
					break
				}
			}
		}

		// Extract imports
		if matches := config.ImportPattern.FindStringSubmatch(line); matches != nil {
			for _, match := range matches[1:] {
				if match != "" {
					summary.Imports = append(summary.Imports, match)
					break
				}
			}
		}

		// Extract types
		if matches := config.TypePattern.FindStringSubmatch(line); matches != nil {
			for _, match := range matches[1:] {
				if match != "" {
					summary.Types = append(summary.Types, match)
					break
				}
			}
		}

		// Extract structs
		if matches := config.StructPattern.FindStringSubmatch(line); matches != nil {
			for _, match := range matches[1:] {
				if match != "" {
					summary.Structs = append(summary.Structs, match)
					break
				}
			}
		}
	}

	summary.LineCount = lineCount

	if err := scanner.Err(); err != nil {
		summary.Error = fmt.Sprintf("Error reading file: %v", err)
	}

	return summary
}

// parseMarkdown extracts headers, links, and structure from markdown files
func (s *Summarizer) parseMarkdown(fullPath string, summary FileSummary) FileSummary {
	// Read the entire markdown file
	content, err := os.ReadFile(fullPath)
	if err != nil {
		summary.Error = fmt.Sprintf("Error reading file: %v", err)
		return summary
	}

	contentStr := string(content)
	summary.LineCount = strings.Count(contentStr, "\n") + 1

	// Parse for metadata (headers, links)
	lines := strings.Split(contentStr, "\n")
	headerPattern := regexp.MustCompile(`^(#{1,6})\s+(.+)`)
	linkPattern := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Extract headers
		if matches := headerPattern.FindStringSubmatch(line); matches != nil {
			level := len(matches[1])
			header := fmt.Sprintf("%s %s", strings.Repeat("#", level), matches[2])
			summary.Headers = append(summary.Headers, header)
		}

		// Extract links
		linkMatches := linkPattern.FindAllStringSubmatch(line, -1)
		for _, match := range linkMatches {
			if len(match) >= 3 {
				link := fmt.Sprintf("[%s](%s)", match[1], match[2])
				summary.Links = append(summary.Links, link)
			}
		}
	}

	// Render markdown with glamour
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80), // Default width, will be adjusted in UI
	)
	if err == nil {
		rendered, renderErr := renderer.Render(contentStr)
		if renderErr == nil {
			summary.RenderedContent = rendered
			summary.IsRendered = true
		} else {
			// Fallback to plain text if rendering fails
			summary.Content = lines
			if len(summary.Content) > 50 {
				summary.Content = summary.Content[:50]
				summary.Content = append(summary.Content, "... (truncated)")
			}
		}
	} else {
		// Fallback to plain text if glamour fails
		summary.Content = lines
		if len(summary.Content) > 50 {
			summary.Content = summary.Content[:50]
			summary.Content = append(summary.Content, "... (truncated)")
		}
	}

	return summary
}

// parseJSON analyzes JSON structure
func (s *Summarizer) parseJSON(fullPath string, summary FileSummary) FileSummary {
	file, err := os.Open(fullPath)
	if err != nil {
		summary.Error = fmt.Sprintf("Error opening file: %v", err)
		return summary
	}
	defer file.Close()

	// Read the entire file for JSON parsing
	content, err := os.ReadFile(fullPath)
	if err != nil {
		summary.Error = fmt.Sprintf("Error reading file: %v", err)
		return summary
	}

	var jsonData interface{}
	if err := json.Unmarshal(content, &jsonData); err != nil {
		summary.Error = fmt.Sprintf("Invalid JSON: %v", err)
		return summary
	}

	// Extract keys from JSON
	summary.ConfigKeys = extractJSONKeys(jsonData, "")
	summary.LineCount = strings.Count(string(content), "\n") + 1

	return summary
}

// parseTextFile handles plain text files
func (s *Summarizer) parseTextFile(fullPath string, summary FileSummary) FileSummary {
	content, err := os.ReadFile(fullPath)
	if err != nil {
		summary.Error = fmt.Sprintf("Error reading file: %v", err)
		return summary
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")
	summary.LineCount = len(lines)

	// Store first 50 lines as content preview
	maxLines := 50
	if len(lines) < maxLines {
		maxLines = len(lines)
	}

	summary.Content = lines[:maxLines]

	// Add truncation indicator if needed
	if len(lines) > 50 {
		summary.Content = append(summary.Content, fmt.Sprintf("... (%d more lines)", len(lines)-50))
	}

	return summary
}

// getExecutableHelp attempts to get help text from an executable
func (s *Summarizer) getExecutableHelp(fullPath string) string {
	fileName := filepath.Base(fullPath)

	// Special case: if this is likely the current running executable, provide help directly
	if fileName == "parsec.exe" || fileName == "parsec" {
		return `Usage: parsec [directory]

Navigate and summarize files in a terminal-based interface.

Examples:
  parsec                       # Scan current directory  
  parsec C:\Projects           # Scan C:\Projects (Windows)
  parsec /home/user/code       # Scan /home/user/code (Unix)
  parsec "C:\Program Files"    # Use quotes for paths with spaces

Parsec is a terminal-based file summarizer that provides:
- Split-screen interface with file navigation
- Multi-language source code analysis  
- Markdown rendering with syntax highlighting
- Configuration file parsing
- Directory navigation support

Keyboard Controls:
  ↑/↓ or k/j    Navigate file list
  Enter         Enter directory or open file  
  PgUp/PgDn     Scroll summary content
  Home/End      Jump to first/last file
  t             Toggle directory visibility
  r             Refresh current directory
  q or Ctrl+C   Quit`
	}

	// Try common help flags for other executables
	helpFlags := []string{"--help", "-h", "help", "/?"}

	for _, flag := range helpFlags {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, fullPath, flag)
		output, err := cmd.CombinedOutput()

		if err == nil && len(output) > 0 {
			// Limit output to first 25 lines
			lines := strings.Split(string(output), "\n")
			if len(lines) > 25 {
				lines = lines[:25]
				lines = append(lines, "... (truncated)")
			}
			return strings.Join(lines, "\n")
		}
	}

	// If no help flags work, show basic executable info
	fileInfo, err := os.Stat(fullPath)
	sizeStr := "unknown"
	if err == nil {
		sizeStr = formatBytes(fileInfo.Size())
	}

	return fmt.Sprintf("Executable: %s\n\nThis is an executable file.\nHelp flags (--help, -h, help, /?) did not produce output.\n\nFile size: %s\nType: Binary executable",
		fileName,
		sizeStr)
}

func (s *Summarizer) parseYAML(fullPath string, summary FileSummary) FileSummary {
	// For now, treat as text file
	return s.parseTextFile(fullPath, summary)
}

func (s *Summarizer) parseINI(fullPath string, summary FileSummary) FileSummary {
	content, err := os.ReadFile(fullPath)
	if err != nil {
		summary.Error = fmt.Sprintf("Error reading file: %v", err)
		return summary
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")
	summary.LineCount = len(lines)

	sectionPattern := regexp.MustCompile(`^\[([^\]]+)\]`)
	keyPattern := regexp.MustCompile(`^([^=]+)=(.*)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") || line == "" {
			continue
		}

		// Extract sections
		if matches := sectionPattern.FindStringSubmatch(line); matches != nil {
			summary.ConfigKeys = append(summary.ConfigKeys, fmt.Sprintf("[%s]", matches[1]))
		}

		// Extract keys
		if matches := keyPattern.FindStringSubmatch(line); matches != nil {
			key := strings.TrimSpace(matches[1])
			summary.ConfigKeys = append(summary.ConfigKeys, key)
		}
	}

	// Store content preview (first 40 lines)
	maxLines := 40
	if len(lines) < maxLines {
		maxLines = len(lines)
	}
	summary.Content = lines[:maxLines]

	if len(lines) > 40 {
		summary.Content = append(summary.Content, fmt.Sprintf("... (%d more lines)", len(lines)-40))
	}

	return summary
}

func (s *Summarizer) parseEnv(fullPath string, summary FileSummary) FileSummary {
	file, err := os.Open(fullPath)
	if err != nil {
		summary.Error = fmt.Sprintf("Error opening file: %v", err)
		return summary
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	envPattern := regexp.MustCompile(`^([A-Z_][A-Z0-9_]*)=`)

	for scanner.Scan() {
		lineCount++
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// Extract environment variable names
		if matches := envPattern.FindStringSubmatch(line); matches != nil {
			summary.ConfigKeys = append(summary.ConfigKeys, matches[1])
		}

		// Store content preview
		if len(summary.Content) < 30 {
			summary.Content = append(summary.Content, line)
		}
	}

	summary.LineCount = lineCount
	return summary
}

// extractJSONKeys recursively extracts keys from JSON data
func extractJSONKeys(data interface{}, prefix string) []string {
	var keys []string

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			fullKey := key
			if prefix != "" {
				fullKey = prefix + "." + key
			}
			keys = append(keys, fullKey)

			// Recursively extract nested keys (limited depth)
			if strings.Count(fullKey, ".") < 3 {
				keys = append(keys, extractJSONKeys(value, fullKey)...)
			}
		}
	case []interface{}:
		if len(v) > 0 {
			// Analyze first array element
			keys = append(keys, extractJSONKeys(v[0], prefix+"[0]")...)
		}
	}

	return keys
}

// parseGoFile handles Go-specific parsing with proper comment handling
func (s *Summarizer) parseGoFile(fullPath string, summary FileSummary) FileSummary {
	file, err := os.Open(fullPath)
	if err != nil {
		summary.Error = fmt.Sprintf("Error opening file: %v", err)
		return summary
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	inMultiLineComment := false

	for scanner.Scan() {
		lineCount++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and single-line comments
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Handle multi-line comments
		if strings.Contains(line, "/*") {
			inMultiLineComment = true
		}
		if inMultiLineComment {
			if strings.Contains(line, "*/") {
				inMultiLineComment = false
			}
			continue
		}

		// Extract Go-specific patterns
		if matches := regexp.MustCompile(`^func\s+(\w+)`).FindStringSubmatch(line); matches != nil {
			summary.Functions = append(summary.Functions, matches[1])
			summary.FunctionCount++
		}

		if matches := regexp.MustCompile(`^import\s+["']([^"']+)["']|^\s+["']([^"']+)["']`).FindStringSubmatch(line); matches != nil {
			for _, match := range matches[1:] {
				if match != "" {
					summary.Imports = append(summary.Imports, match)
					break
				}
			}
		}

		if matches := regexp.MustCompile(`^type\s+(\w+)\s+`).FindStringSubmatch(line); matches != nil {
			summary.Types = append(summary.Types, matches[1])
		}

		if matches := regexp.MustCompile(`^type\s+(\w+)\s+struct`).FindStringSubmatch(line); matches != nil {
			summary.Structs = append(summary.Structs, matches[1])
		}
	}

	summary.LineCount = lineCount
	if err := scanner.Err(); err != nil {
		summary.Error = fmt.Sprintf("Error reading file: %v", err)
	}

	return summary
}

// parsePythonFile handles Python-specific parsing
func (s *Summarizer) parsePythonFile(fullPath string, summary FileSummary) FileSummary {
	file, err := os.Open(fullPath)
	if err != nil {
		summary.Error = fmt.Sprintf("Error opening file: %v", err)
		return summary
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	inMultiLineString := false

	for scanner.Scan() {
		lineCount++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle multi-line strings (docstrings)
		if strings.Contains(line, `"""`) || strings.Contains(line, `'''`) {
			inMultiLineString = !inMultiLineString
			continue
		}
		if inMultiLineString {
			continue
		}

		// Extract Python-specific patterns
		if matches := regexp.MustCompile(`^def\s+(\w+)`).FindStringSubmatch(line); matches != nil {
			summary.Functions = append(summary.Functions, matches[1])
			summary.FunctionCount++
		}

		if matches := regexp.MustCompile(`^(?:from\s+(\S+)\s+)?import\s+(.+)`).FindStringSubmatch(line); matches != nil {
			importStr := ""
			if matches[1] != "" {
				importStr = matches[1] + "." + matches[2]
			} else {
				importStr = matches[2]
			}
			summary.Imports = append(summary.Imports, importStr)
		}

		if matches := regexp.MustCompile(`^class\s+(\w+)`).FindStringSubmatch(line); matches != nil {
			summary.Types = append(summary.Types, matches[1])
			summary.Structs = append(summary.Structs, matches[1])
		}
	}

	summary.LineCount = lineCount
	if err := scanner.Err(); err != nil {
		summary.Error = fmt.Sprintf("Error reading file: %v", err)
	}

	return summary
}

// parseJavaScriptFile handles JavaScript/TypeScript-specific parsing
func (s *Summarizer) parseJavaScriptFile(fullPath string, summary FileSummary) FileSummary {
	file, err := os.Open(fullPath)
	if err != nil {
		summary.Error = fmt.Sprintf("Error opening file: %v", err)
		return summary
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	inMultiLineComment := false

	for scanner.Scan() {
		lineCount++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and single-line comments
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Handle multi-line comments
		if strings.Contains(line, "/*") {
			inMultiLineComment = true
		}
		if inMultiLineComment {
			if strings.Contains(line, "*/") {
				inMultiLineComment = false
			}
			continue
		}

		// Extract JavaScript/TypeScript patterns
		if matches := regexp.MustCompile(`^(?:function\s+(\w+)|const\s+(\w+)\s*=.*=>|(\w+)\s*:\s*function)`).FindStringSubmatch(line); matches != nil {
			for _, match := range matches[1:] {
				if match != "" {
					summary.Functions = append(summary.Functions, match)
					summary.FunctionCount++
					break
				}
			}
		}

		if matches := regexp.MustCompile(`^import.*from\s+['"]([^'"]+)['"]|^const\s+.*=\s+require\(['"]([^'"]+)['"]\)`).FindStringSubmatch(line); matches != nil {
			for _, match := range matches[1:] {
				if match != "" {
					summary.Imports = append(summary.Imports, match)
					break
				}
			}
		}

		if matches := regexp.MustCompile(`^(?:class\s+(\w+)|interface\s+(\w+))`).FindStringSubmatch(line); matches != nil {
			for _, match := range matches[1:] {
				if match != "" {
					summary.Types = append(summary.Types, match)
					summary.Structs = append(summary.Structs, match)
					break
				}
			}
		}
	}

	summary.LineCount = lineCount
	if err := scanner.Err(); err != nil {
		summary.Error = fmt.Sprintf("Error reading file: %v", err)
	}

	return summary
}

// parseRustFile handles Rust-specific parsing
func (s *Summarizer) parseRustFile(fullPath string, summary FileSummary) FileSummary {
	file, err := os.Open(fullPath)
	if err != nil {
		summary.Error = fmt.Sprintf("Error opening file: %v", err)
		return summary
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	inMultiLineComment := false

	for scanner.Scan() {
		lineCount++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and single-line comments
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Handle multi-line comments
		if strings.Contains(line, "/*") {
			inMultiLineComment = true
		}
		if inMultiLineComment {
			if strings.Contains(line, "*/") {
				inMultiLineComment = false
			}
			continue
		}

		// Extract Rust-specific patterns
		if matches := regexp.MustCompile(`^(?:pub\s+)?fn\s+(\w+)`).FindStringSubmatch(line); matches != nil {
			summary.Functions = append(summary.Functions, matches[1])
			summary.FunctionCount++
		}

		if matches := regexp.MustCompile(`^use\s+([^;]+);`).FindStringSubmatch(line); matches != nil {
			summary.Imports = append(summary.Imports, matches[1])
		}

		if matches := regexp.MustCompile(`^(?:pub\s+)?(?:struct\s+(\w+)|enum\s+(\w+)|type\s+(\w+))`).FindStringSubmatch(line); matches != nil {
			for _, match := range matches[1:] {
				if match != "" {
					summary.Types = append(summary.Types, match)
					break
				}
			}
		}

		if matches := regexp.MustCompile(`^(?:pub\s+)?struct\s+(\w+)`).FindStringSubmatch(line); matches != nil {
			summary.Structs = append(summary.Structs, matches[1])
		}
	}

	summary.LineCount = lineCount
	if err := scanner.Err(); err != nil {
		summary.Error = fmt.Sprintf("Error reading file: %v", err)
	}

	return summary
}

// parseCppFile handles C++-specific parsing
func (s *Summarizer) parseCppFile(fullPath string, summary FileSummary) FileSummary {
	file, err := os.Open(fullPath)
	if err != nil {
		summary.Error = fmt.Sprintf("Error opening file: %v", err)
		return summary
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	inMultiLineComment := false

	for scanner.Scan() {
		lineCount++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and single-line comments
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Handle multi-line comments
		if strings.Contains(line, "/*") {
			inMultiLineComment = true
		}
		if inMultiLineComment {
			if strings.Contains(line, "*/") {
				inMultiLineComment = false
			}
			continue
		}

		// Extract C++-specific patterns
		if matches := regexp.MustCompile(`^(?:[a-zA-Z_][a-zA-Z0-9_]*\s+)?(?:[a-zA-Z_][a-zA-Z0-9_]*::)?\w+\s+(\w+)\s*\(`).FindStringSubmatch(line); matches != nil {
			summary.Functions = append(summary.Functions, matches[1])
			summary.FunctionCount++
		}

		if matches := regexp.MustCompile(`^#include\s+[<"]([^>"]+)[>"]`).FindStringSubmatch(line); matches != nil {
			summary.Imports = append(summary.Imports, matches[1])
		}

		if matches := regexp.MustCompile(`^(?:class|struct|enum|union)\s+(\w+)`).FindStringSubmatch(line); matches != nil {
			summary.Types = append(summary.Types, matches[1])
		}

		if matches := regexp.MustCompile(`^(?:class|struct)\s+(\w+)`).FindStringSubmatch(line); matches != nil {
			summary.Structs = append(summary.Structs, matches[1])
		}
	}

	summary.LineCount = lineCount
	if err := scanner.Err(); err != nil {
		summary.Error = fmt.Sprintf("Error reading file: %v", err)
	}

	return summary
}

// formatBytes formats file size in human-readable format
func formatBytes(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	if size >= GB {
		return fmt.Sprintf("%.1f GB", float64(size)/GB)
	} else if size >= MB {
		return fmt.Sprintf("%.1f MB", float64(size)/MB)
	} else if size >= KB {
		return fmt.Sprintf("%.1f KB", float64(size)/KB)
	}
	return fmt.Sprintf("%d B", size)
}

// getLanguage determines the programming language based on file extension
func getLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	languageMap := map[string]string{
		// Programming languages
		".go":    "Go",
		".py":    "Python",
		".js":    "JavaScript",
		".ts":    "TypeScript",
		".jsx":   "React/JSX",
		".tsx":   "React/TSX",
		".rs":    "Rust",
		".java":  "Java",
		".c":     "C",
		".cpp":   "C++",
		".cc":    "C++",
		".h":     "C Header",
		".hpp":   "C++ Header",
		".cs":    "C#",
		".php":   "PHP",
		".rb":    "Ruby",
		".swift": "Swift",
		".kt":    "Kotlin",
		".scala": "Scala",

		// Markup and documentation
		".md":       "Markdown",
		".markdown": "Markdown",
		".txt":      "Text",
		".rst":      "reStructuredText",

		// Configuration files
		".json":       "JSON",
		".yaml":       "YAML",
		".yml":        "YAML",
		".toml":       "TOML",
		".ini":        "INI",
		".cfg":        "Config",
		".conf":       "Config",
		".env":        "Environment",
		".properties": "Properties",

		// Data files
		".xml": "XML",
		".csv": "CSV",
		".log": "Log",

		// Shell and scripts
		".sh":   "Shell",
		".bash": "Bash",
		".zsh":  "Zsh",
		".fish": "Fish",
		".ps1":  "PowerShell",
		".bat":  "Batch",
		".cmd":  "Command",
	}

	if lang, exists := languageMap[ext]; exists {
		return lang
	}
	return "Unknown"
}
