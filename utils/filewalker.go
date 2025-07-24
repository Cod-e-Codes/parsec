package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// SupportedExtensions defines the file extensions we support for summarization
var SupportedExtensions = map[string]bool{
	// Programming languages
	".go":    true,
	".py":    true,
	".rs":    true,
	".js":    true,
	".ts":    true,
	".jsx":   true,
	".tsx":   true,
	".java":  true,
	".c":     true,
	".cpp":   true,
	".h":     true,
	".hpp":   true,
	".cs":    true,
	".php":   true,
	".rb":    true,
	".swift": true,
	".kt":    true,
	".scala": true,

	// Markup and documentation
	".md":       true,
	".markdown": true,
	".txt":      true,
	".rst":      true,

	// Configuration files
	".json":       true,
	".yaml":       true,
	".yml":        true,
	".toml":       true,
	".ini":        true,
	".cfg":        true,
	".conf":       true,
	".env":        true,
	".properties": true,

	// Data files
	".xml": true,
	".csv": true,
	".log": true,

	// Shell and scripts
	".sh":   true,
	".bash": true,
	".zsh":  true,
	".fish": true,
	".ps1":  true,
	".bat":  true,
	".cmd":  true,
}

// FileInfo represents basic information about a discovered file
type FileInfo struct {
	Path      string
	Name      string
	Extension string
	IsDir     bool
}

// Walker handles filesystem traversal and filtering
type Walker struct {
	basePath string
	files    []FileInfo
}

// NewWalker creates a new file walker for the given base path
func NewWalker(basePath string) *Walker {
	return &Walker{
		basePath: basePath,
		files:    make([]FileInfo, 0),
	}
}

// ListDirectory lists only the immediate children of the specified directory
func (w *Walker) ListDirectory(dirPath string) ([]FileInfo, error) {
	w.files = make([]FileInfo, 0)

	// Use dirPath directly as the directory to scan
	scanPath := dirPath

	entries, err := os.ReadDir(scanPath)
	if err != nil {
		return w.files, err
	}

	// Add parent directory entry if not at base
	if dirPath != w.basePath {
		parentPath := filepath.Dir(dirPath)
		relParentPath, _ := filepath.Rel(w.basePath, parentPath)
		if relParentPath == "." {
			relParentPath = ""
		}

		parentInfo := FileInfo{
			Path:      "..",
			Name:      "..",
			Extension: "",
			IsDir:     true,
		}
		w.files = append(w.files, parentInfo)
	}

	// Add all entries in current directory
	for _, entry := range entries {
		// Skip hidden files and directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Skip common non-source directories
		if entry.IsDir() {
			skipDirs := []string{"node_modules", "vendor", "target", "build", "dist", ".git"}
			shouldSkip := false
			for _, skipDir := range skipDirs {
				if entry.Name() == skipDir {
					shouldSkip = true
					break
				}
			}
			if shouldSkip {
				continue
			}
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		// Show all files and directories (not just supported extensions)
		fileInfo := FileInfo{
			Path:      entry.Name(),
			Name:      entry.Name(),
			Extension: ext,
			IsDir:     entry.IsDir(),
		}
		w.files = append(w.files, fileInfo)
	}

	return w.files, nil
}

// IsSourceFile checks if a file has a supported extension
func IsSourceFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return SupportedExtensions[ext]
}

// IsExecutableFile checks if a file is executable
func IsExecutableFile(filePath string) bool {
	// Check Windows executable extensions
	if runtime.GOOS == "windows" {
		ext := strings.ToLower(filepath.Ext(filePath))
		winExeExts := []string{".exe", ".com", ".bat", ".cmd", ".ps1", ".msi"}
		for _, exeExt := range winExeExts {
			if ext == exeExt {
				return true
			}
		}
		return false
	}

	// For Unix-like systems, check file permissions
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	// Check if any execute bit is set
	mode := info.Mode()
	return mode&0111 != 0 // Check if any execute permission bit is set
}
