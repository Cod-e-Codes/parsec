package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"parsec/internal/testutil"
)

func TestListDirectory(t *testing.T) {
	testDir := testutil.SetupTestFiles(t)
	walker := NewWalker(testDir)

	// Test root directory listing
	t.Run("Root directory", func(t *testing.T) {
		files, err := walker.ListDirectory(testDir)
		testutil.AssertNoError(t, err)

		// Should find our test directories
		foundGo := false
		foundMarkdown := false
		foundConfig := false

		for _, file := range files {
			switch file.Path {
			case "go":
				foundGo = true
				testutil.AssertEqual(t, file.IsDir, true)
			case "markdown":
				foundMarkdown = true
				testutil.AssertEqual(t, file.IsDir, true)
			case "config":
				foundConfig = true
				testutil.AssertEqual(t, file.IsDir, true)
			}
		}

		testutil.AssertEqual(t, foundGo, true)
		testutil.AssertEqual(t, foundMarkdown, true)
		testutil.AssertEqual(t, foundConfig, true)
	})

	// Test subdirectory listing
	t.Run("Subdirectory", func(t *testing.T) {
		files, err := walker.ListDirectory(filepath.Join(testDir, "go"))
		testutil.AssertNoError(t, err)

		// Should find parent directory and sample.go
		foundParent := false
		foundSample := false

		for _, file := range files {
			switch file.Path {
			case "..":
				foundParent = true
				testutil.AssertEqual(t, file.IsDir, true)
			case "sample.go":
				foundSample = true
				testutil.AssertEqual(t, file.IsDir, false)
				testutil.AssertEqual(t, file.Extension, ".go")
			}
		}

		testutil.AssertEqual(t, foundParent, true)
		testutil.AssertEqual(t, foundSample, true)
	})

	// Test error cases
	t.Run("Non-existent directory", func(t *testing.T) {
		_, err := walker.ListDirectory(filepath.Join(testDir, "does-not-exist"))
		testutil.AssertError(t, err)
	})

	t.Run("File instead of directory", func(t *testing.T) {
		_, err := walker.ListDirectory(filepath.Join(testDir, "go", "sample.go"))
		testutil.AssertError(t, err)
	})
}

func TestIsSourceFile(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{"test.go", true},
		{"test.py", true},
		{"test.js", true},
		{"test.ts", true},
		{"test.md", true},
		{"test.json", true},
		{"test.unknown", false},
		{"test", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := IsSourceFile(tt.filename)
			testutil.AssertEqual(t, got, tt.want)
		})
	}
}

func TestIsExecutableFile(t *testing.T) {
	testDir := testutil.CreateTempDir(t)

	// Create a regular file
	regularPath := filepath.Join(testDir, "regular.txt")
	err := os.WriteFile(regularPath, []byte("test"), 0644)
	testutil.AssertNoError(t, err)

	// Create an executable file
	var execPath string
	var execContent []byte
	if runtime.GOOS == "windows" {
		execPath = filepath.Join(testDir, "test.exe")
		execContent = []byte{0x4D, 0x5A} // DOS/Windows executable header
	} else {
		execPath = filepath.Join(testDir, "test.sh")
		execContent = []byte("#!/bin/sh\necho test")
	}
	err = os.WriteFile(execPath, execContent, 0755)
	testutil.AssertNoError(t, err)

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "Regular file",
			path: regularPath,
			want: false,
		},
		{
			name: "Executable file",
			path: execPath,
			want: true,
		},
		{
			name: "Non-existent file",
			path: filepath.Join(testDir, "does-not-exist"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsExecutableFile(tt.path)
			testutil.AssertEqual(t, got, tt.want)
		})
	}
}
