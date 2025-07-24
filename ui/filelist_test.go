package ui

import (
	"strings"
	"testing"

	"parsec/utils"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFileListModel(t *testing.T) {
	m := NewFileListModel()

	// Test initial state
	t.Run("Initial state", func(t *testing.T) {
		if m.cursor != 0 {
			t.Errorf("Expected cursor to be 0, got %d", m.cursor)
		}
		if !m.showDirs {
			t.Error("Expected showDirs to be true")
		}
		if len(m.files) != 0 {
			t.Errorf("Expected empty files list, got %d files", len(m.files))
		}
	})

	// Test file list updates
	t.Run("SetFiles", func(t *testing.T) {
		files := []utils.FileInfo{
			{Path: "test1.go", IsDir: false, Extension: ".go"},
			{Path: "test2.go", IsDir: false, Extension: ".go"},
		}
		m.SetFiles(files)

		if len(m.files) != 2 {
			t.Errorf("Expected 2 files, got %d", len(m.files))
		}
		if m.cursor >= len(m.files) {
			t.Errorf("Cursor %d out of bounds for %d files", m.cursor, len(m.files))
		}
	})

	// Test cursor movement
	t.Run("Cursor movement", func(t *testing.T) {
		files := []utils.FileInfo{
			{Path: "test1.go", IsDir: false, Extension: ".go"},
			{Path: "test2.go", IsDir: false, Extension: ".go"},
			{Path: "test3.go", IsDir: false, Extension: ".go"},
		}
		m.SetFiles(files)
		m.cursor = 0

		// Test down movement
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		if m.cursor != 1 {
			t.Errorf("Expected cursor to be 1, got %d", m.cursor)
		}

		// Test up movement
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
		if m.cursor != 0 {
			t.Errorf("Expected cursor to be 0, got %d", m.cursor)
		}

		// Test bounds
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp}) // Should stay at 0
		if m.cursor != 0 {
			t.Errorf("Expected cursor to stay at 0, got %d", m.cursor)
		}

		// Test end key
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnd})
		if m.cursor != 2 {
			t.Errorf("Expected cursor to be at end (2), got %d", m.cursor)
		}

		// Test home key
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyHome})
		if m.cursor != 0 {
			t.Errorf("Expected cursor to be at start (0), got %d", m.cursor)
		}
	})

	// Test directory toggling
	t.Run("Directory toggling", func(t *testing.T) {
		files := []utils.FileInfo{
			{Path: "dir1", IsDir: true},
			{Path: "test1.go", IsDir: false},
			{Path: "dir2", IsDir: true},
		}
		m.SetFiles(files)
		m.showDirs = true
		m.SetDimensions(40, 10) // Ensure we have enough space to show all files

		// Toggle directories off
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
		if m.showDirs {
			t.Error("Expected showDirs to be false after toggle")
		}

		// Check view doesn't contain directories
		view := m.View()
		if strings.Contains(view, "📁 dir1") || strings.Contains(view, "📁 dir2") {
			t.Error("View should not contain directories when hidden")
		}

		// Toggle directories back on
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
		if !m.showDirs {
			t.Error("Expected showDirs to be true after second toggle")
		}

		// Check view contains directories
		view = m.View()
		if !strings.Contains(view, "📁 dir1") || !strings.Contains(view, "📁 dir2") {
			t.Error("View should contain directories when shown")
		}
	})

	// Test view rendering
	t.Run("View rendering", func(t *testing.T) {
		m.SetDimensions(40, 10)
		files := []utils.FileInfo{
			{Path: "test.go", IsDir: false, Extension: ".go"},
			{Path: "test.py", IsDir: false, Extension: ".py"},
		}
		m.SetFiles(files)

		view := m.View()

		// Check for file icons
		if !strings.Contains(view, "🐹") { // Go icon
			t.Error("View should contain Go file icon")
		}
		if !strings.Contains(view, "🐍") { // Python icon
			t.Error("View should contain Python file icon")
		}

		// Check for cursor
		if !strings.Contains(view, ">") {
			t.Error("View should contain cursor indicator")
		}

		// Check for file count
		if !strings.Contains(view, "2 files") {
			t.Error("View should contain file count")
		}
	})
}

func TestGetFileIcon(t *testing.T) {
	tests := []struct {
		ext  string
		want string
	}{
		{".go", "🐹"},
		{".py", "🐍"},
		{".js", "📄"},
		{".ts", "📘"},
		{".md", "📝"},
		{".unknown", "📄"},
		{"", "📄"},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			got := getFileIcon(tt.ext)
			if got != tt.want {
				t.Errorf("getFileIcon(%q) = %q, want %q", tt.ext, got, tt.want)
			}
		})
	}
}
