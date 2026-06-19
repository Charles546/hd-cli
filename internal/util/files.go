// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If you have a separate written commercial agreement, you may use this file under those terms instead.

package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FileExists checks if a file exists and is not a directory
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// WriteFile writes content to a file atomically (write to temp, then rename)
func WriteFile(path string, content []byte, perm os.FileMode) error {
	// Create the directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := EnsureDir(dir); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create temp file in the same directory
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	// Ensure cleanup on any failure
	defer func() {
		if _, err := os.Stat(tmpPath); err == nil {
			os.Remove(tmpPath) //nolint:errcheck
		}
	}()

	// Write content
	if _, err := tmp.Write(content); err != nil {
		tmp.Close() //nolint:errcheck
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Set permissions before closing
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close() //nolint:errcheck
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	// Close the file
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Rename to final path (atomic on most filesystems)
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// EnsureDir creates a directory tree if it doesn't exist
func EnsureDir(path string) error {
	if DirExists(path) {
		return nil
	}

	err := os.MkdirAll(path, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}

	return nil
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	// Open source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close() //nolint:errcheck

	// Ensure destination directory exists
	if err := EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}

	// Create destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close() //nolint:errcheck

	// Copy content
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Copy permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set destination file permissions: %w", err)
	}

	return nil
}

// ReadFile reads the entire content of a file
func ReadFile(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return content, nil
}
