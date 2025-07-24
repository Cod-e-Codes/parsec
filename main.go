package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"parsec/core"
	"parsec/ui"
	"parsec/utils"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

// UI dimension constants
const (
	headerHeight = 1
	footerHeight = 1
	borderWidth  = 2
	borderHeight = 2
)

// Application model containing the main state
type model struct {
	fileListModel ui.FileListModel
	summaryModel  ui.SummaryModel
	summarizer    *core.Summarizer
	walker        *utils.Walker
	selectedPath  string
	width         int
	height        int
	basePath      string
	currentDir    string // Current directory being browsed

	// Search state
	searchMode    bool
	searchQuery   string
	allFiles      []utils.FileInfo // All files in current directory for filtering
	filteredFiles []utils.FileInfo // Files after fuzzy search filter
}

// SummaryMsg is sent when a file summary is ready
type SummaryMsg struct {
	summary      core.FileSummary
	path         string // The relative path from base directory
	selectedPath string // The path as selected in the file list
}

func (m model) Init() tea.Cmd {
	return loadFilesCmd(m.walker, m.currentDir)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.sizeComponents()
		return m, nil

	case LoadFilesMsg:
		// Store all files and apply search filter if active
		m.allFiles = msg.files
		if m.searchMode && m.searchQuery != "" {
			m.filteredFiles = m.fuzzyFilter(m.allFiles, m.searchQuery)
			m.fileListModel.SetFiles(m.filteredFiles)
		} else {
			m.filteredFiles = m.allFiles
			m.fileListModel.SetFiles(m.allFiles)
		}

		// Check if we have a new selection after loading files
		var cmds []tea.Cmd
		if selected := m.fileListModel.GetSelectedFile(); selected != nil && selected.Path != m.selectedPath {
			m.selectedPath = selected.Path
			if cmd := m.handleFileSelection(selected); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return m, tea.Batch(cmds...)

	case SummaryMsg:
		// Update summary when async summary is ready
		if msg.selectedPath == m.selectedPath {
			m.summaryModel.SetSummary(&msg.summary)
		}
		return m, nil

	case DirectoryPreviewMsg:
		// Update summary with directory preview
		if selected := m.fileListModel.GetSelectedFile(); selected != nil && selected.Path == msg.dirName {
			m.summaryModel.SetSummary(nil)
			m.summaryModel.SetContent(msg.content)
		}
		return m, nil

	case tea.KeyMsg:
		// Handle search mode input
		if m.searchMode {
			// Check for escape key using key type for better reliability
			if msg.Type == tea.KeyEscape {
				// Exit search mode
				m.searchMode = false
				m.searchQuery = ""
				m.fileListModel.SetFiles(m.allFiles)
				return m, nil
			}

			// Handle other search mode keys
			switch msg.Type {
			case tea.KeyEnter:
				// Exit search mode but keep filter
				m.searchMode = false
				return m, nil
			case tea.KeyBackspace:
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					if m.searchQuery == "" {
						m.filteredFiles = m.allFiles
						m.fileListModel.SetFiles(m.allFiles)
					} else {
						m.filteredFiles = m.fuzzyFilter(m.allFiles, m.searchQuery)
						m.fileListModel.SetFiles(m.filteredFiles)
					}
				}
				return m, nil
			default:
				// Add character to search query
				if len(msg.String()) == 1 && msg.String() >= " " && msg.String() <= "~" {
					m.searchQuery += msg.String()
					m.filteredFiles = m.fuzzyFilter(m.allFiles, m.searchQuery)
					m.fileListModel.SetFiles(m.filteredFiles)
				}
				return m, nil
			}
		}

		// Normal mode keyboard handling
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "/":
			// Enter search mode
			m.searchMode = true
			m.searchQuery = ""
			return m, nil

		case "r":
			// Refresh file list
			return m, loadFilesCmd(m.walker, m.currentDir)

		case "pgup":
			// Scroll summary up
			m.summaryModel.Scroll(-5)
			return m, nil

		case "pgdown":
			// Scroll summary down
			m.summaryModel.Scroll(5)
			return m, nil

		case "enter":
			// Navigate into directory or select file
			if selected := m.fileListModel.GetSelectedFile(); selected != nil && selected.IsDir {
				if selected.Path == ".." {
					// Go up one directory
					m.currentDir = filepath.Dir(m.currentDir)
				} else {
					// Go into the selected directory
					m.currentDir = filepath.Join(m.currentDir, selected.Path)
				}
				// Load files from new directory
				return m, loadFilesCmd(m.walker, m.currentDir)
			}
			return m, nil

		default:
			// Handle file list navigation
			var cmd tea.Cmd
			m.fileListModel, cmd = m.fileListModel.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}

			// Check if selection changed
			if selected := m.fileListModel.GetSelectedFile(); selected != nil && selected.Path != m.selectedPath {
				m.selectedPath = selected.Path
				if cmd := m.handleFileSelection(selected); cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// fuzzyFilter applies fuzzy search to filter files
func (m model) fuzzyFilter(files []utils.FileInfo, query string) []utils.FileInfo {
	if query == "" {
		return files
	}

	// Create a slice of file paths for fuzzy matching
	var filePaths []string
	for _, file := range files {
		filePaths = append(filePaths, file.Path)
	}

	// Perform fuzzy search
	matches := fuzzy.Find(query, filePaths)

	// Convert matches back to FileInfo slice
	var filtered []utils.FileInfo
	for _, match := range matches {
		for _, file := range files {
			if file.Path == match.Str {
				filtered = append(filtered, file)
				break
			}
		}
	}

	return filtered
}

// DirectoryPreviewMsg is sent when directory preview is ready
type DirectoryPreviewMsg struct {
	dirName string
	content string
}

// showDirectoryPreview shows a preview of directory contents in the summary pane
func (m *model) showDirectoryPreview(dirName string) tea.Cmd {
	return func() tea.Msg {
		// Construct full path to the directory
		dirPath := filepath.Join(m.currentDir, dirName)

		// Get directory contents
		files, err := m.walker.ListDirectory(dirPath)
		if err != nil {
			return DirectoryPreviewMsg{
				dirName: dirName,
				content: fmt.Sprintf("📁 Directory: %s\n\nError reading directory: %v", dirName, err),
			}
		}

		// Format directory contents for display
		content := m.formatDirectoryPreview(dirName, files, dirPath)
		return DirectoryPreviewMsg{
			dirName: dirName,
			content: content,
		}
	}
}

// formatDirectoryPreview formats directory contents for display
func (m model) formatDirectoryPreview(dirName string, files []utils.FileInfo, dirPath string) string {
	var result strings.Builder

	// Get relative path for display
	relDirPath, _ := filepath.Rel(m.basePath, dirPath)
	if relDirPath == "." {
		relDirPath = "/"
	} else {
		relDirPath = "/" + relDirPath
	}

	result.WriteString(fmt.Sprintf("📁 Directory: %s\n", dirName))
	result.WriteString(fmt.Sprintf("Path: %s\n\n", relDirPath))

	// Filter out ".." entry for preview since it's just navigation
	var previewFiles []utils.FileInfo
	for _, file := range files {
		if file.Path != ".." {
			previewFiles = append(previewFiles, file)
		}
	}

	if len(previewFiles) == 0 {
		result.WriteString("📭 This directory is empty.\n\nPress Enter to navigate into this directory.")
		return result.String()
	}

	// Count files and directories
	fileCount := 0
	dirCount := 0
	for _, file := range previewFiles {
		if file.IsDir {
			dirCount++
		} else {
			fileCount++
		}
	}

	result.WriteString(fmt.Sprintf("Contains: %d files", fileCount))
	if dirCount > 0 {
		result.WriteString(fmt.Sprintf(", %d directories", dirCount))
	}
	result.WriteString("\n\n")

	// Show first several items
	maxItems := 20
	if len(previewFiles) > maxItems {
		result.WriteString(fmt.Sprintf("First %d items:\n", maxItems))
	} else {
		result.WriteString("Contents:\n")
	}

	for i, file := range previewFiles {
		if i >= maxItems {
			break
		}

		// Get file icon
		var icon string
		if file.IsDir {
			icon = "📁"
		} else {
			// Get file icon based on extension
			icon = getFileIcon(file.Extension)
		}

		result.WriteString(fmt.Sprintf("  %s %s", icon, file.Path))
		if file.IsDir {
			result.WriteString("/")
		}
		result.WriteString("\n")
	}

	if len(previewFiles) > maxItems {
		result.WriteString(fmt.Sprintf("  ... and %d more items\n", len(previewFiles)-maxItems))
	}

	result.WriteString("\nPress Enter to navigate into this directory.")
	return result.String()
}

// getFileIcon returns an appropriate icon for the file extension
func getFileIcon(ext string) string {
	icons := map[string]string{
		// Programming languages
		".go":    "🐹",
		".py":    "🐍",
		".js":    "📄",
		".ts":    "📘",
		".jsx":   "⚛️",
		".tsx":   "⚛️",
		".rs":    "🦀",
		".java":  "☕",
		".c":     "📄",
		".cpp":   "📄",
		".h":     "📄",
		".hpp":   "📄",
		".cs":    "🔷",
		".php":   "🐘",
		".rb":    "💎",
		".swift": "🍎",
		".kt":    "📱",
		".scala": "⚖️",

		// Documentation and markup
		".md":       "📝",
		".markdown": "📝",
		".txt":      "📄",
		".rst":      "📜",

		// Configuration files
		".json":       "🔧",
		".yaml":       "⚙️",
		".yml":        "⚙️",
		".toml":       "⚙️",
		".ini":        "⚙️",
		".cfg":        "⚙️",
		".conf":       "⚙️",
		".env":        "🌿",
		".properties": "⚙️",

		// Data files
		".xml": "📋",
		".csv": "📊",
		".log": "📜",

		// Shell and scripts
		".sh":   "🐚",
		".bash": "🐚",
		".zsh":  "🐚",
		".fish": "🐠",
		".ps1":  "💻",
		".bat":  "💻",
		".cmd":  "💻",

		// Executables
		".exe": "⚙️",
		".bin": "⚙️",
	}

	if icon, exists := icons[ext]; exists {
		return icon
	}
	return "📄"
}

// handleFileSelection processes file selection and starts summarization if appropriate
func (m *model) handleFileSelection(selected *utils.FileInfo) tea.Cmd {
	if selected == nil {
		m.summaryModel.SetSummary(nil)
		return nil
	}

	// Handle directory selection - show directory contents preview
	if selected.IsDir {
		if selected.Path == ".." {
			// Show parent directory info
			parentPath := filepath.Dir(m.currentDir)
			relParentPath, _ := filepath.Rel(m.basePath, parentPath)
			if relParentPath == "." {
				relParentPath = "/"
			} else {
				relParentPath = "/" + relParentPath
			}

			m.summaryModel.SetSummary(nil)
			m.summaryModel.SetContent(fmt.Sprintf("📁 Parent Directory\n\nPath: %s\n\nPress Enter to navigate up to this directory.", relParentPath))
			return nil
		} else {
			// Show directory contents preview
			return m.showDirectoryPreview(selected.Path)
		}
	}

	// Handle file selection
	fullPath := filepath.Join(m.currentDir, selected.Path)
	if utils.IsSourceFile(selected.Path) || utils.IsExecutableFile(fullPath) {
		// Start loading summary for the new selection
		m.summaryModel.SetLoading(true)
		// Create relative path for summarization
		relPath, _ := filepath.Rel(m.basePath, fullPath)
		return summarizeFileCmd(m.summarizer, relPath, selected.Path)
	} else {
		// Show a simple message for unsupported files
		m.summaryModel.SetSummary(nil)
		m.summaryModel.SetContent(fmt.Sprintf("File: %s\n\nThis file type is not supported for summarization.", selected.Path))
		return nil
	}
}

// sizeComponents updates component dimensions based on current window size
func (m model) sizeComponents() model {
	if m.width <= 0 || m.height <= 0 {
		return m
	}

	// Calculate available space accounting for header, borders and footer

	availableWidth := m.width
	availableHeight := m.height - headerHeight - footerHeight

	// Each pane gets half the width minus minimal border space
	paneWidth := (availableWidth / 2) - borderWidth
	paneHeight := availableHeight - borderHeight

	// Ensure minimum sizes
	if paneWidth < 10 {
		paneWidth = 10
	}
	if paneHeight < 5 {
		paneHeight = 5
	}

	m.fileListModel.SetDimensions(paneWidth, paneHeight)
	m.summaryModel.SetDimensions(paneWidth, paneHeight)
	return m
}

func (m model) View() string {
	// If we don't have valid dimensions yet, return a simple message
	if m.width <= 0 || m.height <= 0 {
		return "Initializing..."
	}

	// Calculate available space accounting for header, borders and footer

	availableWidth := m.width
	availableHeight := m.height - headerHeight - footerHeight

	paneWidth := (availableWidth / 2) - borderWidth
	paneHeight := availableHeight - borderHeight

	// Ensure minimum sizes
	if paneWidth < 10 {
		paneWidth = 10
	}
	if paneHeight < 5 {
		paneHeight = 5
	}

	// Define styles for the split panes
	leftPaneStyle := lipgloss.NewStyle().
		Width(paneWidth).
		Height(paneHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	rightPaneStyle := lipgloss.NewStyle().
		Width(paneWidth).
		Height(paneHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	// Render file list
	fileListView := m.fileListModel.View()
	leftPane := leftPaneStyle.Render(fileListView)

	// Render summary
	summaryView := m.summaryModel.View()
	rightPane := rightPaneStyle.Render(summaryView)

	// Join the panes horizontally
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	// Add current directory header
	relCurrentDir, _ := filepath.Rel(m.basePath, m.currentDir)
	if relCurrentDir == "." {
		relCurrentDir = "/"
	} else {
		relCurrentDir = "/" + relCurrentDir
	}
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true).
		Render(fmt.Sprintf("📁 %s", relCurrentDir))

	// Add footer with help text or search input
	var footer string
	if m.searchMode {
		// Show search input
		searchStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

		cursor := ""
		if len(m.searchQuery)%2 == 0 { // Simple blinking cursor simulation
			cursor = "█"
		}

		footer = searchStyle.Render(fmt.Sprintf("Search: %s%s", m.searchQuery, cursor)) +
			lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(" (ESC to cancel, Enter to confirm)")
	} else {
		// Show regular help
		footer = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render("↑/↓ navigate • Enter to open • / search • PgUp/PgDn scroll • t toggle dirs • r refresh • q quit")
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

func initialModel(basePath string) model {
	m := model{
		fileListModel: ui.NewFileListModel(),
		summaryModel:  ui.NewSummaryModel(),
		summarizer:    core.NewSummarizer(basePath),
		walker:        utils.NewWalker(basePath),
		basePath:      basePath,
		currentDir:    basePath, // Start in the base directory
		selectedPath:  "",
		width:         80, // Set reasonable default dimensions
		height:        24,

		// Initialize search state
		searchMode:    false,
		searchQuery:   "",
		allFiles:      make([]utils.FileInfo, 0),
		filteredFiles: make([]utils.FileInfo, 0),
	}
	return m.sizeComponents()
}

// LoadFilesMsg is sent when files are loaded from disk
type LoadFilesMsg struct {
	files []utils.FileInfo
}

// loadFilesCmd loads files from the specified directory
func loadFilesCmd(walker *utils.Walker, dirPath string) tea.Cmd {
	return func() tea.Msg {
		files, err := walker.ListDirectory(dirPath)
		if err != nil {
			// Return empty list on error
			files = make([]utils.FileInfo, 0)
		}
		return LoadFilesMsg{files: files}
	}
}

// summarizeFileCmd creates a summary for the specified file
func summarizeFileCmd(summarizer *core.Summarizer, filePath string, selectedPath string) tea.Cmd {
	return func() tea.Msg {
		summary := summarizer.SummarizeFile(filePath)
		return SummaryMsg{summary: summary, path: filePath, selectedPath: selectedPath}
	}
}

func main() {
	// Set custom usage function to show comprehensive help
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `Usage: parsec [directory]

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
- Directory navigation with live preview
- Smart directory content inspection

Keyboard Controls:
  ↑/↓ or k/j    Navigate file list
  Enter         Enter directory or open file  
  /             Start fuzzy search
  PgUp/PgDn     Scroll summary content
  Home/End      Jump to first/last file
  t             Toggle directory visibility
  r             Refresh current directory
  q or Ctrl+C   Quit

Search Mode:
  Type          Add characters to search query
  Backspace     Remove last character
  Enter         Confirm search and stay filtered
  ESC           Cancel search and show all files
`)
	}

	flag.Parse()

	// Get directory from positional argument or use current directory
	basePath := "."
	if flag.NArg() > 0 {
		basePath = flag.Arg(0)
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(basePath)
	if err != nil {
		fmt.Printf("Error getting absolute path: %v\n", err)
		os.Exit(1)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Printf("Directory does not exist: %s\n", absPath)
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel(absPath), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
