package main

import (
	"path/filepath"
	"testing"

	"parsec/internal/testutil"
	"parsec/utils"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInitialModel(t *testing.T) {
	testDir := testutil.SetupTestFiles(t)
	m := initialModel(testDir)

	// Test initial state
	if m.basePath != testDir {
		t.Errorf("Expected basePath %q, got %q", testDir, m.basePath)
	}
	if m.currentDir != testDir {
		t.Errorf("Expected currentDir %q, got %q", testDir, m.currentDir)
	}
	if m.searchMode {
		t.Error("Expected search mode to be false")
	}
	if m.searchQuery != "" {
		t.Error("Expected empty search query")
	}

	// Test initial dimensions
	if m.width < 10 || m.height < 5 {
		t.Error("Initial dimensions should be reasonable defaults")
	}
}

func TestModelUpdate(t *testing.T) {
	testDir := testutil.SetupTestFiles(t)
	m := initialModel(testDir)

	// Test window resize
	t.Run("Window resize", func(t *testing.T) {
		newModel, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
		m = newModel.(model)
		if m.width != 100 || m.height != 50 {
			t.Error("Model should update dimensions on window resize")
		}
	})

	// Test file loading
	t.Run("Load files", func(t *testing.T) {
		newModel, _ := m.Update(LoadFilesMsg{files: []utils.FileInfo{
			{Path: "test.go", IsDir: false},
			{Path: "test.md", IsDir: false},
		}})
		m = newModel.(model)
		if len(m.allFiles) != 2 {
			t.Error("Model should store loaded files")
		}
	})

	// Test search mode
	t.Run("Search mode", func(t *testing.T) {
		// Enter search mode
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		m = newModel.(model)
		if !m.searchMode {
			t.Error("'/' key should enter search mode")
		}

		// Type search query one character at a time
		chars := []rune{'t', 'e', 's', 't'}
		for _, ch := range chars {
			newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
			m = newModel.(model)
		}
		if m.searchQuery != "test" {
			t.Errorf("Expected search query 'test', got %q", m.searchQuery)
		}

		// Test backspace
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		m = newModel.(model)
		if m.searchQuery != "tes" {
			t.Errorf("Expected search query 'tes', got %q", m.searchQuery)
		}

		// Exit search mode with Escape
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
		m = newModel.(model)
		if m.searchMode {
			t.Error("Escape key should exit search mode")
		}
		if m.searchQuery != "" {
			t.Error("Search query should be cleared after escape")
		}
	})

	// Test directory navigation
	t.Run("Directory navigation", func(t *testing.T) {
		// Navigate into directory
		m.allFiles = []utils.FileInfo{{Path: "go", IsDir: true}}
		m.fileListModel.SetFiles(m.allFiles)
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m = newModel.(model)

		expectedDir := filepath.Join(testDir, "go")
		if m.currentDir != expectedDir {
			t.Errorf("Expected current directory %q, got %q", expectedDir, m.currentDir)
		}

		// Navigate back up
		m.allFiles = []utils.FileInfo{{Path: "..", IsDir: true}}
		m.fileListModel.SetFiles(m.allFiles)
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m = newModel.(model)

		if m.currentDir != testDir {
			t.Errorf("Expected to return to base directory %q, got %q", testDir, m.currentDir)
		}
	})
}

func TestFuzzyFilter(t *testing.T) {
	m := initialModel(".")
	files := []utils.FileInfo{
		{Path: "test.go"},
		{Path: "main.go"},
		{Path: "README.md"},
	}

	tests := []struct {
		query string
		want  int // Number of expected matches
	}{
		{"test", 1},
		{"go", 2},
		{"read", 1},
		{"xyz", 0},
		{"", 3}, // Empty query should return all files
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			filtered := m.fuzzyFilter(files, tt.query)
			if len(filtered) != tt.want {
				t.Errorf("fuzzyFilter(%q) returned %d files, want %d", tt.query, len(filtered), tt.want)
			}
		})
	}
}

func TestHandleFileSelection(t *testing.T) {
	testDir := testutil.SetupTestFiles(t)
	m := initialModel(testDir)

	tests := []struct {
		name    string
		file    *utils.FileInfo
		wantCmd bool // Whether a command should be returned
	}{
		{
			name:    "Nil file",
			file:    nil,
			wantCmd: false,
		},
		{
			name:    "Directory",
			file:    &utils.FileInfo{Path: "test", IsDir: true},
			wantCmd: true,
		},
		{
			name:    "Go file",
			file:    &utils.FileInfo{Path: "test.go", IsDir: false},
			wantCmd: true,
		},
		{
			name:    "Unsupported file",
			file:    &utils.FileInfo{Path: "test.xyz", IsDir: false},
			wantCmd: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := m.handleFileSelection(tt.file)
			if (cmd != nil) != tt.wantCmd {
				t.Errorf("handleFileSelection() returned command: %v, want command: %v", cmd != nil, tt.wantCmd)
			}
		})
	}
}
