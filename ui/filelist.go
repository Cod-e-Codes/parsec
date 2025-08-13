package ui

import (
	"fmt"
	"strings"

	"parsec/utils"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FileListModel handles file navigation in the left pane
type FileListModel struct {
	files       []utils.FileInfo
	cursor      int
	selected    string
	width       int
	height      int
	showDirs    bool
	baseStyle   lipgloss.Style
	cursorStyle lipgloss.Style
}

// NewFileListModel creates a new file list model
func NewFileListModel() FileListModel {
	return FileListModel{
		files:    make([]utils.FileInfo, 0),
		cursor:   0,
		showDirs: true,
		baseStyle: lipgloss.NewStyle().
			Padding(1).
			MarginLeft(1),
		cursorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true),
	}
}

// SetFiles updates the file list
func (m *FileListModel) SetFiles(files []utils.FileInfo) {
	m.files = files
	if len(files) > 0 && m.cursor >= len(files) {
		m.cursor = len(files) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// GetSelectedFile returns the currently selected file
func (m FileListModel) GetSelectedFile() *utils.FileInfo {
	if len(m.files) == 0 || m.cursor >= len(m.files) {
		return nil
	}
	return &m.files[m.cursor]
}

// SetDimensions updates the model dimensions
func (m *FileListModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}

// Update handles file list specific updates
func (m FileListModel) Update(msg tea.Msg) (FileListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.cursor = max(0, m.cursor-1)
			if selected := m.GetSelectedFile(); selected != nil {
				m.selected = selected.Path
			}

		case "down", "j":
			m.cursor = min(len(m.files)-1, m.cursor+1)
			if selected := m.GetSelectedFile(); selected != nil {
				m.selected = selected.Path
			}

		case "home":
			m.cursor = 0
			if selected := m.GetSelectedFile(); selected != nil {
				m.selected = selected.Path
			}

		case "end":
			m.cursor = len(m.files) - 1
			if selected := m.GetSelectedFile(); selected != nil {
				m.selected = selected.Path
			}

		case "pageup":
			m.cursor = max(0, m.cursor-10)
			if selected := m.GetSelectedFile(); selected != nil {
				m.selected = selected.Path
			}

		case "pagedown":
			m.cursor = min(len(m.files)-1, m.cursor+10)
			if selected := m.GetSelectedFile(); selected != nil {
				m.selected = selected.Path
			}

		case "t":
			// Toggle directory visibility
			m.showDirs = !m.showDirs
		}
	}

	return m, nil
}

// View renders the file list
func (m FileListModel) View() string {
	if len(m.files) == 0 {
		return m.baseStyle.Render("No files found...")
	}

	// Calculate available height for file entries
	availableHeight := m.height - 4 // Account for padding
	if availableHeight < 1 {
		availableHeight = 1
	}

	// Reserve space for file count footer
	fileListHeight := availableHeight - 1
	if fileListHeight < 1 {
		fileListHeight = 1
	}

	// Get filtered files (excluding hidden directories if needed)
	var displayFiles []utils.FileInfo
	for _, file := range m.files {
		if m.showDirs || !file.IsDir {
			displayFiles = append(displayFiles, file)
		}
	}

	// Calculate visible range
	startIdx := 0
	endIdx := len(displayFiles)

	if len(displayFiles) > fileListHeight {
		// Center the cursor in the view
		halfVisible := fileListHeight / 2
		cursorInDisplayFiles := 0

		// Find cursor position in filtered list
		for i, file := range displayFiles {
			if m.cursor < len(m.files) && file.Path == m.files[m.cursor].Path {
				cursorInDisplayFiles = i
				break
			}
		}

		startIdx = max(0, cursorInDisplayFiles-halfVisible)
		endIdx = min(len(displayFiles), startIdx+fileListHeight)

		// Adjust if we're near the end
		if endIdx == len(displayFiles) {
			startIdx = max(0, endIdx-fileListHeight)
		}
	}

	// Build the file list with fixed height
	var lines []string
	for i := startIdx; i < endIdx; i++ {
		file := displayFiles[i]

		// Cursor indicator (ensure consistent 2-character width)
		cursor := "  "
		if m.cursor < len(m.files) && file.Path == m.files[m.cursor].Path {
			cursor = "> " // Already 2 characters
		}

		// File type indicator
		var indicator string
		if file.IsDir {
			if file.Path == ".." {
				indicator = "⬆️" // Special icon for parent directory
			} else {
				indicator = "📁"
			}
		} else {
			indicator = utils.GetFileIcon(file.Extension)
		}

		// Selection indicator (ensure consistent 1-character width)
		selected := " "
		if file.Path == m.selected {
			selected = "✓" // Single character
		}

		// File name (truncate if too long)
		name := file.Path

		// Calculate consistent width accounting for all components
		// cursor: 2 chars, brackets: 2 chars, selected: 1 char, space: 1 char, indicator: 2 chars, space: 1 char
		// Total fixed width: 9 characters
		maxNameLength := m.width - 9
		if maxNameLength > 0 && len(name) > maxNameLength {
			name = name[:maxNameLength-3] + "..."
		}

		// Use helper function to ensure consistent width formatting
		line := formatFileLine(cursor, selected, indicator, name)

		if m.cursor < len(m.files) && file.Path == m.files[m.cursor].Path {
			line = m.cursorStyle.Render(line)
		}

		lines = append(lines, line)
	}

	// Pad with empty lines to maintain consistent height
	for len(lines) < fileListHeight {
		lines = append(lines, "")
	}

	// Add file count footer
	footer := fmt.Sprintf("%d files", len(displayFiles))
	if !m.showDirs {
		footer += " (dirs hidden - press 't' to toggle)"
	}
	lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(footer))

	content := strings.Join(lines, "\n")
	return m.baseStyle.Render(content)
}

// Helper functions for min/max
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// formatFileLine ensures consistent width formatting for file list entries
func formatFileLine(cursor, selected, indicator, name string) string {
	// Ensure cursor is exactly 2 characters
	if len(cursor) < 2 {
		cursor = cursor + strings.Repeat(" ", 2-len(cursor))
	} else if len(cursor) > 2 {
		cursor = cursor[:2]
	}

	// Ensure selected is exactly 1 character
	if len(selected) < 1 {
		selected = " "
	} else if len(selected) > 1 {
		selected = selected[:1]
	}

	// Ensure indicator has consistent spacing (add space if needed)
	indicatorWithSpace := indicator
	if len(indicator) < 2 {
		indicatorWithSpace = indicator + " "
	}

	return fmt.Sprintf("%s[%s] %s %s", cursor, selected, indicatorWithSpace, name)
}
