package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetTestDataPath(t *testing.T) {
	path := GetTestDataPath(t)
	if path == "" {
		t.Error("Expected non-empty test data path")
	}
	if !filepath.IsAbs(path) {
		t.Error("Expected absolute path")
	}

	// Get the project root directory
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Failed to get caller information")
	}
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filename)))

	// Check if the testdata path is under the project root
	relPath, err := filepath.Rel(projectRoot, path)
	if err != nil {
		t.Errorf("Failed to get relative path: %v", err)
	}
	if relPath != "testdata" {
		t.Errorf("Expected testdata path to be 'testdata', got %q", relPath)
	}
}

func TestCreateTempDir(t *testing.T) {
	dir := CreateTempDir(t)
	defer os.RemoveAll(dir) // Cleanup in case t.Cleanup fails

	// Check directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("Temp directory was not created")
	}

	// Check directory is empty
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Errorf("Failed to read temp directory: %v", err)
	}
	if len(entries) > 0 {
		t.Error("Expected empty directory")
	}

	// Write a file to test cleanup
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Errorf("Failed to write test file: %v", err)
	}
}

func TestCopyFile(t *testing.T) {
	// Create source file
	srcDir := CreateTempDir(t)
	srcPath := filepath.Join(srcDir, "source.txt")
	content := []byte("test content")
	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy to destination
	dstDir := CreateTempDir(t)
	dstPath := filepath.Join(dstDir, "nested", "dest.txt")
	CopyFile(t, srcPath, dstPath)

	// Verify copy
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Errorf("Failed to read destination file: %v", err)
	}
	if string(dstContent) != string(content) {
		t.Error("Destination content does not match source")
	}
}

func TestSetupTestFiles(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := CreateTempDir(t)

	// Set up test files in the temporary directory
	testDir := SetupTestFiles(t)

	// Verify that the test files were created
	expectedFiles := []string{
		"go/sample.go",
		"markdown/sample.md",
		"config/sample.json",
	}

	for _, file := range expectedFiles {
		fullPath := filepath.Join(testDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", file)
		}
	}

	// Clean up
	os.RemoveAll(tempDir)
}

// mockT implements TestingT for assertion tests
type mockT struct {
	failed bool
}

func (m *mockT) Helper()                                   {}
func (m *mockT) Error(args ...interface{})                 { m.failed = true }
func (m *mockT) Errorf(format string, args ...interface{}) { m.failed = true }

func TestAssertions(t *testing.T) {
	// Test AssertEqual
	t.Run("AssertEqual", func(t *testing.T) {
		mock := &mockT{}
		AssertEqual(mock, 1, 1) // Should pass
		if mock.failed {
			t.Error("AssertEqual failed for equal values")
		}

		mock = &mockT{}         // Reset mock
		AssertEqual(mock, 1, 2) // Should fail
		if !mock.failed {
			t.Error("AssertEqual passed for unequal values")
		}
	})

	// Test AssertContains
	t.Run("AssertContains", func(t *testing.T) {
		mock := &mockT{}
		AssertContains(mock, "hello world", "world") // Should pass
		if mock.failed {
			t.Error("AssertContains failed for contained substring")
		}

		mock = &mockT{}                            // Reset mock
		AssertContains(mock, "hello world", "xyz") // Should fail
		if !mock.failed {
			t.Error("AssertContains passed for non-contained substring")
		}
	})

	// Test AssertNoError
	t.Run("AssertNoError", func(t *testing.T) {
		mock := &mockT{}
		AssertNoError(mock, nil) // Should pass
		if mock.failed {
			t.Error("AssertNoError failed for nil error")
		}

		mock = &mockT{}                     // Reset mock
		AssertNoError(mock, os.ErrNotExist) // Should fail
		if !mock.failed {
			t.Error("AssertNoError passed for non-nil error")
		}
	})

	// Test AssertError
	t.Run("AssertError", func(t *testing.T) {
		mock := &mockT{}
		AssertError(mock, os.ErrNotExist) // Should pass
		if mock.failed {
			t.Error("AssertError failed for non-nil error")
		}

		mock = &mockT{}        // Reset mock
		AssertError(mock, nil) // Should fail
		if !mock.failed {
			t.Error("AssertError passed for nil error")
		}
	})
}
