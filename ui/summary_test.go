package ui

import (
	"strings"
	"testing"

	"parsec/core"
)

func TestSummaryModel(t *testing.T) {
	m := NewSummaryModel()

	// Test initial state
	t.Run("Initial state", func(t *testing.T) {
		if m.summary != nil {
			t.Error("Expected nil summary in initial state")
		}
		if m.content != "Select a file to view its summary" {
			t.Error("Expected default content message")
		}
	})

	// Test loading state
	t.Run("Loading state", func(t *testing.T) {
		m.SetLoading(true)
		view := m.View()
		if !strings.Contains(view, "Loading...") {
			t.Error("View should show loading state")
		}
	})

	// Test summary display
	t.Run("Summary display", func(t *testing.T) {
		m.SetDimensions(80, 40)
		summary := &core.FileSummary{
			Path:          "test.go",
			Language:      "Go",
			LineCount:     100,
			FunctionCount: 5,
			Functions:     []string{"main", "test"},
			Imports:       []string{"fmt", "strings"},
		}

		m.SetSummary(summary)
		view := m.View()

		if !strings.Contains(view, "Language: Go") {
			t.Error("View should contain language info")
		}
		if !strings.Contains(view, "Lines: 100") {
			t.Error("View should contain line count")
		}
		if !strings.Contains(view, "Functions: 5") {
			t.Error("View should contain function count")
		}
	})

	// Test markdown rendering
	t.Run("Markdown rendering", func(t *testing.T) {
		m.SetDimensions(80, 40)
		summary := &core.FileSummary{
			Path:            "test.md",
			Language:        "Markdown",
			LineCount:       10,
			Headers:         []string{"# Title", "## Section"},
			Links:           []string{"[Link](https://example.com)"},
			IsRendered:      true,
			RenderedContent: "# Title\n\nRendered content",
		}

		m.SetSummary(summary)
		view := m.View()

		// Check for rendered content
		if !strings.Contains(view, "# Title") {
			t.Error("View should contain rendered markdown")
		}

		// Check for headers section
		if !strings.Contains(view, "📋 Headers") || !strings.Contains(view, "# Title") {
			t.Error("View should contain Headers section")
		}

		// Check for links section
		if !strings.Contains(view, "🔗 Links") || !strings.Contains(view, "[Link](https://example.com)") {
			t.Error("View should contain Links section")
		}
	})

	// Test error display
	t.Run("Error display", func(t *testing.T) {
		summary := &core.FileSummary{
			Path:  "error.txt",
			Error: "Failed to read file",
		}

		m.SetSummary(summary)
		view := m.View()

		if !strings.Contains(view, "Error:") {
			t.Error("View should show error message")
		}
		if !strings.Contains(view, "Failed to read file") {
			t.Error("View should contain error details")
		}
	})

	// Test scrolling
	t.Run("Scrolling", func(t *testing.T) {
		m.SetDimensions(40, 10)
		content := strings.Repeat("Line\n", 20)
		m.SetContent(content)

		// Test scroll down
		m.Scroll(5)
		if m.scrollPos != 5 {
			t.Errorf("Expected scroll position 5, got %d", m.scrollPos)
		}

		// Test scroll up
		m.Scroll(-2)
		if m.scrollPos != 3 {
			t.Errorf("Expected scroll position 3, got %d", m.scrollPos)
		}

		// Test scroll bounds
		m.Scroll(-10) // Try to scroll past top
		if m.scrollPos != 0 {
			t.Errorf("Expected scroll position 0, got %d", m.scrollPos)
		}

		// Test scroll past bottom
		maxScroll := strings.Count(content, "\n") - (m.height - 4)
		m.scrollPos = 0         // Reset scroll position
		m.Scroll(maxScroll + 1) // Try to scroll one past max
		if m.scrollPos != maxScroll {
			t.Errorf("Scroll position %d should be capped at max scroll %d", m.scrollPos, maxScroll)
		}
	})
}

func TestFormatPath(t *testing.T) {
	m := NewSummaryModel()
	m.SetDimensions(40, 10)

	tests := []struct {
		path string
		want string
	}{
		{"short.go", "short.go"},
		{"very/long/path/to/file.go", "file.go"},
		{"src/project/main.go", "project/main.go"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := m.formatPath(tt.path)
			if !strings.HasSuffix(got, tt.want) {
				t.Errorf("formatPath(%q) = %q, want suffix %q", tt.path, got, tt.want)
			}
			if len(got) > m.width-6 { // Account for icon and padding
				t.Errorf("Formatted path %q exceeds max width %d", got, m.width-6)
			}
		})
	}
}
