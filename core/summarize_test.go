package core

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"parsec/internal/testutil"
)

func TestSummarizeFile(t *testing.T) {
	testDir := testutil.SetupTestFiles(t)
	summarizer := NewSummarizer(testDir)

	tests := []struct {
		name     string
		filePath string
		want     FileSummary
	}{
		{
			name:     "Go file",
			filePath: "go/sample.go",
			want: FileSummary{
				Path:          "go/sample.go",
				Language:      "Go",
				LineCount:     25,
				FunctionCount: 3, // main, greet, processUser
				Functions:     []string{"main", "greet", "processUser"},
				Imports:       []string{"fmt"},
				Types:         []string{"User"},
				Structs:       []string{"User"},
			},
		},
		{
			name:     "Markdown file",
			filePath: "markdown/sample.md",
			want: FileSummary{
				Path:       "markdown/sample.md",
				Language:   "Markdown",
				LineCount:  25,
				Headers:    []string{"# Sample Document", "## Section 1", "### Subsection 1.1", "## Section 2", "### Code Sample"},
				Links:      []string{"[Link to Google](https://www.google.com)", "[Link to GitHub](https://www.github.com)"},
				IsRendered: true,
			},
		},
		{
			name:     "JSON file",
			filePath: "config/sample.json",
			want: FileSummary{
				Path:      "config/sample.json",
				Language:  "JSON",
				LineCount: 13,
				ConfigKeys: []string{
					"name",
					"version",
					"config",
					"config.port",
					"config.host",
					"config.database",
					"config.database.url",
					"config.database.maxConnections",
					"features",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := summarizer.SummarizeFile(tt.filePath)

			// Check basic fields
			testutil.AssertEqual(t, got.Path, tt.want.Path)
			testutil.AssertEqual(t, got.Language, tt.want.Language)
			testutil.AssertEqual(t, got.LineCount, tt.want.LineCount)
			testutil.AssertEqual(t, got.FunctionCount, tt.want.FunctionCount)

			// Check slices if they should exist
			if len(tt.want.Functions) > 0 {
				testutil.AssertEqual(t, len(got.Functions), len(tt.want.Functions))
				// Sort both slices to ensure consistent order
				gotFuncs := make([]string, len(got.Functions))
				wantFuncs := make([]string, len(tt.want.Functions))
				copy(gotFuncs, got.Functions)
				copy(wantFuncs, tt.want.Functions)
				sort.Strings(gotFuncs)
				sort.Strings(wantFuncs)

				for i, fn := range wantFuncs {
					if i < len(gotFuncs) {
						testutil.AssertEqual(t, gotFuncs[i], fn)
					}
				}
			}

			if len(tt.want.Imports) > 0 {
				testutil.AssertEqual(t, len(got.Imports), len(tt.want.Imports))
				// Sort both slices to ensure consistent order
				gotImports := make([]string, len(got.Imports))
				wantImports := make([]string, len(tt.want.Imports))
				copy(gotImports, got.Imports)
				copy(wantImports, tt.want.Imports)
				sort.Strings(gotImports)
				sort.Strings(wantImports)

				for i, imp := range wantImports {
					if i < len(gotImports) {
						testutil.AssertEqual(t, gotImports[i], imp)
					}
				}
			}

			if len(tt.want.Types) > 0 {
				testutil.AssertEqual(t, len(got.Types), len(tt.want.Types))
				// Sort both slices to ensure consistent order
				gotTypes := make([]string, len(got.Types))
				wantTypes := make([]string, len(tt.want.Types))
				copy(gotTypes, got.Types)
				copy(wantTypes, tt.want.Types)
				sort.Strings(gotTypes)
				sort.Strings(wantTypes)

				for i, typ := range wantTypes {
					if i < len(gotTypes) {
						testutil.AssertEqual(t, gotTypes[i], typ)
					}
				}
			}

			if len(tt.want.Headers) > 0 {
				testutil.AssertEqual(t, len(got.Headers), len(tt.want.Headers))
				for i, header := range tt.want.Headers {
					if i < len(got.Headers) {
						testutil.AssertEqual(t, got.Headers[i], header)
					}
				}
			}

			if len(tt.want.Links) > 0 {
				testutil.AssertEqual(t, len(got.Links), len(tt.want.Links))
				for i, link := range tt.want.Links {
					if i < len(got.Links) {
						testutil.AssertEqual(t, got.Links[i], link)
					}
				}
			}

			if len(tt.want.ConfigKeys) > 0 {
				testutil.AssertEqual(t, len(got.ConfigKeys), len(tt.want.ConfigKeys))

				// Sort both slices to ensure consistent order
				gotKeys := make([]string, len(got.ConfigKeys))
				wantKeys := make([]string, len(tt.want.ConfigKeys))
				copy(gotKeys, got.ConfigKeys)
				copy(wantKeys, tt.want.ConfigKeys)
				sort.Strings(gotKeys)
				sort.Strings(wantKeys)

				for i, key := range wantKeys {
					if i < len(gotKeys) {
						testutil.AssertEqual(t, gotKeys[i], key)
					}
				}
			}

			// Check rendered content for markdown
			if tt.want.IsRendered {
				if got.RenderedContent == "" {
					t.Error("Expected rendered content, got empty string")
				}
			}
		})
	}
}

func TestSummarizeFile_Errors(t *testing.T) {
	testDir := testutil.SetupTestFiles(t)
	summarizer := NewSummarizer(testDir)

	tests := []struct {
		name     string
		filePath string
	}{
		{
			name:     "Non-existent file",
			filePath: "does-not-exist.txt",
		},
		{
			name:     "Directory instead of file",
			filePath: "go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := summarizer.SummarizeFile(tt.filePath)
			if got.Error == "" {
				t.Error("Expected error, got none")
			}
		})
	}
}

func TestExecutableHelp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping executable help test in short mode")
	}

	testDir := testutil.CreateTempDir(t)
	summarizer := NewSummarizer(testDir)

	// Create a test executable
	execPath := filepath.Join(testDir, "test.exe")
	testutil.CopyFile(t, os.Args[0], execPath) // Copy the test binary itself

	summary := summarizer.SummarizeFile("test.exe")
	testutil.AssertEqual(t, summary.IsExecutable, true)
	testutil.AssertContains(t, summary.ExecutableHelp, "Usage")
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		size int64
		want string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{1024 * 1024 * 1024 * 2, "2.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatBytes(tt.size)
			testutil.AssertEqual(t, got, tt.want)
		})
	}
}

func TestGetLanguage(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"test.go", "Go"},
		{"test.py", "Python"},
		{"test.js", "JavaScript"},
		{"test.ts", "TypeScript"},
		{"test.md", "Markdown"},
		{"test.json", "JSON"},
		{"test.unknown", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := getLanguage(tt.filename)
			testutil.AssertEqual(t, got, tt.want)
		})
	}
}
