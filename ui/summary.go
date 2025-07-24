package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"parsec/core"

	"github.com/charmbracelet/lipgloss"
)

// SummaryModel handles the right pane summary display
type SummaryModel struct {
	summary     *core.FileSummary
	content     string
	width       int
	height      int
	scrollPos   int
	baseStyle   lipgloss.Style
	titleStyle  lipgloss.Style
	errorStyle  lipgloss.Style
	loadingText string
	isLoading   bool
}

// NewSummaryModel creates a new summary model
func NewSummaryModel() SummaryModel {
	return SummaryModel{
		summary:     nil,
		content:     "Select a file to view its summary",
		scrollPos:   0,
		loadingText: "Loading...",
		isLoading:   false,
		baseStyle: lipgloss.NewStyle().
			Padding(1).
			MarginLeft(1),
		titleStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Underline(true),
		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),
	}
}

// SetSummary updates the displayed summary
func (m *SummaryModel) SetSummary(summary *core.FileSummary) {
	m.summary = summary
	m.isLoading = false
	m.scrollPos = 0

	if summary == nil {
		m.content = "No file selected"
		return
	}

	m.content = m.formatSummaryForDisplay(*summary)
}

// SetLoading sets the loading state
func (m *SummaryModel) SetLoading(loading bool) {
	m.isLoading = loading
	if loading {
		m.content = m.loadingText
	}
}

// SetContent sets custom content directly
func (m *SummaryModel) SetContent(content string) {
	m.summary = nil
	m.isLoading = false
	m.scrollPos = 0
	m.content = content
}

// SetDimensions updates the model dimensions
func (m *SummaryModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}

// Scroll adjusts the scroll position
func (m *SummaryModel) Scroll(delta int) {
	lines := strings.Split(m.content, "\n")

	// Use same height calculation as View method
	availableHeight := m.height - 4 // Account for padding only
	if availableHeight < 1 {
		availableHeight = 1
	}

	maxScroll := 0
	if len(lines) > availableHeight {
		maxScroll = len(lines) - availableHeight
	}

	newPos := m.scrollPos + delta
	if newPos < 0 {
		newPos = 0
	}
	if newPos > maxScroll {
		newPos = maxScroll
	}
	m.scrollPos = newPos
}

// View renders the summary
func (m SummaryModel) View() string {
	if m.isLoading {
		return m.baseStyle.Render(m.loadingText)
	}

	if m.summary != nil && m.summary.Error != "" {
		errorContent := m.errorStyle.Render("Error: ") + m.summary.Error
		return m.baseStyle.Render(errorContent)
	}

	// Split content into lines and apply scrolling
	lines := strings.Split(m.content, "\n")

	// Calculate available height - be more conservative
	availableHeight := m.height - 4 // Account for padding only
	if availableHeight < 1 {
		availableHeight = 1
	}

	// Calculate visible window
	startIdx := m.scrollPos
	endIdx := startIdx + availableHeight
	if endIdx > len(lines) {
		endIdx = len(lines)
	}

	// Build the visible content with fixed height
	var visibleLines []string
	for i := startIdx; i < endIdx; i++ {
		line := lines[i]
		// Truncate long lines to fit width
		maxLineLength := m.width - 4 // Account for padding
		if maxLineLength > 0 && len(line) > maxLineLength {
			line = line[:maxLineLength-3] + "..."
		}
		visibleLines = append(visibleLines, line)
	}

	// Pad with empty lines to maintain consistent height
	for len(visibleLines) < availableHeight {
		visibleLines = append(visibleLines, "")
	}

	content := strings.Join(visibleLines, "\n")
	return m.baseStyle.Render(content)
}

// formatSummaryForDisplay formats a FileSummary for display
func (m SummaryModel) formatSummaryForDisplay(summary core.FileSummary) string {
	var result strings.Builder

	// Title with proper wrapping for long paths
	formattedPath := m.formatPath(summary.Path)
	result.WriteString(m.titleStyle.Render(fmt.Sprintf("📄 %s", formattedPath)))
	result.WriteString("\n\n")

	// Basic info
	result.WriteString(fmt.Sprintf("Language: %s\n", summary.Language))
	result.WriteString(fmt.Sprintf("Lines: %d\n", summary.LineCount))
	if summary.FileSize > 0 {
		result.WriteString(fmt.Sprintf("Size: %s\n", formatFileSize(summary.FileSize)))
	}
	if summary.FunctionCount > 0 {
		result.WriteString(fmt.Sprintf("Functions: %d\n", summary.FunctionCount))
	}
	result.WriteString("\n")

	// Handle different content types
	if summary.IsExecutable {
		result.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true).Render("⚙️ Executable Help:"))
		result.WriteString("\n\n")
		result.WriteString(summary.ExecutableHelp)
		return result.String()
	}

	// For markdown files with rendered content
	if summary.IsRendered && summary.RenderedContent != "" {
		result.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true).Render("📝 Rendered Content:"))
		result.WriteString("\n\n")
		result.WriteString(summary.RenderedContent)
		return result.String()
	}

	// For text files with content preview
	if len(summary.Content) > 0 {
		result.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Bold(true).Render("📖 Content Preview:"))
		result.WriteString("\n\n")
		for _, line := range summary.Content {
			result.WriteString(line + "\n")
		}
		result.WriteString("\n")
	}

	// Headers for markdown files
	if len(summary.Headers) > 0 {
		result.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("171")).Bold(true).Render("📋 Headers:"))
		result.WriteString("\n")
		for i, header := range summary.Headers {
			if i < 10 { // Show max 10 headers
				result.WriteString(fmt.Sprintf("  %s\n", header))
			}
		}
		if len(summary.Headers) > 10 {
			result.WriteString(fmt.Sprintf("  ... and %d more\n", len(summary.Headers)-10))
		}
		result.WriteString("\n")
	}

	// Config keys for configuration files
	if len(summary.ConfigKeys) > 0 {
		result.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("🔧 Configuration Keys:"))
		result.WriteString("\n")
		for i, key := range summary.ConfigKeys {
			if i < 15 { // Show max 15 keys
				result.WriteString(fmt.Sprintf("  • %s\n", key))
			}
		}
		if len(summary.ConfigKeys) > 15 {
			result.WriteString(fmt.Sprintf("  ... and %d more\n", len(summary.ConfigKeys)-15))
		}
		result.WriteString("\n")
	}

	// Functions section
	if len(summary.Functions) > 0 {
		result.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true).Render("🔧 Functions:"))
		result.WriteString("\n")

		for i, fn := range summary.Functions {
			if i < 15 { // Show max 15 functions
				result.WriteString(fmt.Sprintf("  • %s\n", fn))
			}
		}
		if len(summary.Functions) > 15 {
			result.WriteString(fmt.Sprintf("  ... and %d more\n", len(summary.Functions)-15))
		}
		result.WriteString("\n")
	}

	// Imports section
	if len(summary.Imports) > 0 {
		result.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Bold(true).Render("📦 Imports:"))
		result.WriteString("\n")

		for i, imp := range summary.Imports {
			if i < 10 { // Show max 10 imports
				result.WriteString(fmt.Sprintf("  • %s\n", imp))
			}
		}
		if len(summary.Imports) > 10 {
			result.WriteString(fmt.Sprintf("  ... and %d more\n", len(summary.Imports)-10))
		}
		result.WriteString("\n")
	}

	// Types section
	if len(summary.Types) > 0 {
		result.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("207")).Bold(true).Render("🏷️  Types:"))
		result.WriteString("\n")

		for i, typ := range summary.Types {
			if i < 10 { // Show max 10 types
				result.WriteString(fmt.Sprintf("  • %s\n", typ))
			}
		}
		if len(summary.Types) > 10 {
			result.WriteString(fmt.Sprintf("  ... and %d more\n", len(summary.Types)-10))
		}
		result.WriteString("\n")
	}

	// Structs section (if different from types)
	if len(summary.Structs) > 0 && len(summary.Structs) != len(summary.Types) {
		result.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Bold(true).Render("🏗️  Structs:"))
		result.WriteString("\n")

		for i, str := range summary.Structs {
			if i < 10 { // Show max 10 structs
				result.WriteString(fmt.Sprintf("  • %s\n", str))
			}
		}
		if len(summary.Structs) > 10 {
			result.WriteString(fmt.Sprintf("  ... and %d more\n", len(summary.Structs)-10))
		}
	}

	// Links for markdown files
	if len(summary.Links) > 0 {
		result.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true).Render("🔗 Links:"))
		result.WriteString("\n")
		for i, link := range summary.Links {
			if i < 8 { // Show max 8 links
				result.WriteString(fmt.Sprintf("  %s\n", link))
			}
		}
		if len(summary.Links) > 8 {
			result.WriteString(fmt.Sprintf("  ... and %d more\n", len(summary.Links)-8))
		}
	}

	return result.String()
}

// GetScrollInfo returns current scroll information
func (m SummaryModel) GetScrollInfo() (current, maxScroll int) {
	lines := strings.Split(m.content, "\n")

	// Use same height calculation as other methods
	availableHeight := m.height - 4 // Account for padding only
	if availableHeight < 1 {
		availableHeight = 1
	}

	maxScrollPos := 0
	if len(lines) > availableHeight {
		maxScrollPos = len(lines) - availableHeight
	}
	return m.scrollPos, maxScrollPos
}

// formatFileSize formats file size in human-readable format
// formatPath intelligently formats long file paths to fit within the pane width
func (m SummaryModel) formatPath(path string) string {
	// Account for icon and spaces in the title: "📄 " takes about 3 characters
	maxWidth := m.width - 6 // Account for borders, padding, and icon

	if maxWidth <= 0 {
		maxWidth = 20 // Minimum reasonable width
	}

	if len(path) <= maxWidth {
		return path
	}

	// For very long paths, show the filename and parent directory with "..."
	filename := filepath.Base(path)
	if len(filename) >= maxWidth-3 {
		// If even just the filename is too long, truncate it
		return filename[:maxWidth-3] + "..."
	}

	// Try to show some parent directory context
	dir := filepath.Dir(path)
	if dir == "." || dir == "/" {
		return filename
	}

	// Calculate how much space we have for the directory part
	filenameLen := len(filename)
	remainingSpace := maxWidth - filenameLen - 4 // Account for ".../"

	if remainingSpace > 0 {
		// Try to show the most relevant parent directory
		parentDir := filepath.Base(dir)
		if len(parentDir) <= remainingSpace {
			return parentDir + "/" + filename
		} else if remainingSpace > 3 {
			return parentDir[:remainingSpace-3] + ".../" + filename
		}
	}

	// If we can't fit any parent context, just show "...filename"
	return "..." + filename
}

func formatFileSize(size int64) string {
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
