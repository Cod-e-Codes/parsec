package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestingT is a minimal interface for test assertions
type TestingT interface {
	Helper()
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

// GetTestDataPath returns the absolute path to the testdata directory
func GetTestDataPath(t testing.TB) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Failed to get caller information")
	}
	return filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filename))), "testdata")
}

// CreateTempDir creates a temporary directory and returns its path.
// The directory will be automatically cleaned up when the test finishes.
func CreateTempDir(t testing.TB) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "parsec-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

// CopyFile copies a file from src to dst
func CopyFile(t testing.TB, src, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("Failed to read source file: %v", err)
	}

	err = os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		t.Fatalf("Failed to create destination directory: %v", err)
	}

	err = os.WriteFile(dst, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write destination file: %v", err)
	}
}

// SetupTestFiles copies test files from testdata to a temporary directory
func SetupTestFiles(t testing.TB) string {
	t.Helper()
	tempDir := CreateTempDir(t)
	testDataPath := GetTestDataPath(t)

	// Walk the testdata/samples directory and copy all files
	err := filepath.Walk(filepath.Join(testDataPath, "samples"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(filepath.Join(testDataPath, "samples"), path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(tempDir, relPath)
		CopyFile(t, path, dstPath)
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to copy test files: %v", err)
	}

	return tempDir
}

// AssertEqual compares two values and fails the test if they are not equal
func AssertEqual[T comparable](t TestingT, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

// AssertContains checks if a string contains a substring
func AssertContains(t TestingT, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("got %q, want it to contain %q", got, want)
	}
}

// AssertNoError fails the test if err is not nil
func AssertNoError(t TestingT, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// AssertError fails the test if err is nil
func AssertError(t TestingT, err error) {
	t.Helper()
	if err == nil {
		t.Error("expected error, got nil")
	}
}
