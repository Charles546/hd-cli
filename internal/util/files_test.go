// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If you have a separate written commercial agreement, you may use this file under those terms instead.

package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	// Test with existing file
	tmpFile, err := os.CreateTemp("", "test-file-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name()) //nolint:errcheck
	tmpFile.Close()                //nolint:errcheck

	if !FileExists(tmpFile.Name()) {
		t.Error("Expected FileExists to return true for existing file")
	}

	// Test with non-existent file
	if FileExists("/nonexistent/path/file.txt") {
		t.Error("Expected FileExists to return false for non-existent file")
	}

	// Test with directory
	if FileExists(os.TempDir()) {
		t.Error("Expected FileExists to return false for directory")
	}
}

func TestDirExists(t *testing.T) {
	// Test with existing directory
	if !DirExists(os.TempDir()) {
		t.Error("Expected DirExists to return true for existing directory")
	}

	// Test with non-existent directory
	if DirExists("/nonexistent/directory/path") {
		t.Error("Expected DirExists to return false for non-existent directory")
	}

	// Test with file
	tmpFile, err := os.CreateTemp("", "test-file-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name()) //nolint:errcheck
	tmpFile.Close()                //nolint:errcheck

	if DirExists(tmpFile.Name()) {
		t.Error("Expected DirExists to return false for file")
	}
}

func TestWriteFile(t *testing.T) {
	// Test writing a file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test-write.txt")
	content := []byte("Hello, World!")

	err := WriteFile(filePath, content, 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Verify file exists
	if !FileExists(filePath) {
		t.Error("Expected file to exist after WriteFile")
	}

	// Verify content
	readContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(readContent) != "Hello, World!" {
		t.Errorf("Expected content 'Hello, World!', got: %s", string(readContent))
	}
}

func TestWriteFileInNestedDir(t *testing.T) {
	// Test writing file in nested directory structure
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nested", "dir", "test.txt")
	content := []byte("Nested content")

	err := WriteFile(filePath, content, 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Verify directories were created
	if !DirExists(filepath.Join(tmpDir, "nested")) {
		t.Error("Expected nested directory to exist")
	}
	if !DirExists(filepath.Join(tmpDir, "nested", "dir")) {
		t.Error("Expected deeply nested directory to exist")
	}

	// Verify file exists
	if !FileExists(filePath) {
		t.Error("Expected file to exist")
	}
}

func TestEnsureDir(t *testing.T) {
	// Test creating nested directories
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "new", "nested", "dir")

	err := EnsureDir(newDir)
	if err != nil {
		t.Fatalf("EnsureDir failed: %v", err)
	}

	if !DirExists(newDir) {
		t.Error("Expected directory to exist after EnsureDir")
	}

	// Test calling EnsureDir on existing directory (should not fail)
	err = EnsureDir(newDir)
	if err != nil {
		t.Fatalf("EnsureDir on existing dir failed: %v", err)
	}
}

func TestCopyFile(t *testing.T) {
	// Test copying a file
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "source.txt")
	dstPath := filepath.Join(tmpDir, "destination.txt")

	// Create source file
	content := []byte("Source content")
	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy file
	err := CopyFile(srcPath, dstPath)
	if err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	// Verify destination exists
	if !FileExists(dstPath) {
		t.Error("Expected destination file to exist")
	}

	// Verify content
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(dstContent) != "Source content" {
		t.Errorf("Expected 'Source content', got: %s", string(dstContent))
	}
}

func TestCopyFileToNestedDir(t *testing.T) {
	// Test copying file to nested directory
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "source.txt")
	dstPath := filepath.Join(tmpDir, "nested", "dir", "destination.txt")

	// Create source file
	content := []byte("Nested copy")
	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy file
	err := CopyFile(srcPath, dstPath)
	if err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	// Verify destination exists
	if !FileExists(dstPath) {
		t.Error("Expected destination file to exist")
	}
}

func TestReadFile(t *testing.T) {
	// Test reading a file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test-read.txt")
	content := []byte("Read test content")

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Read file
	readContent, err := ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(readContent) != "Read test content" {
		t.Errorf("Expected 'Read test content', got: %s", string(readContent))
	}
}

func TestReadFileNonexistent(t *testing.T) {
	// Test reading non-existent file
	_, err := ReadFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("Expected error when reading non-existent file")
	}
}
