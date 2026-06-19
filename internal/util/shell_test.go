// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If you have a separate written commercial agreement, you may use this file under those terms instead.

package util

import (
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	// Test successful command execution
	output, err := Run("echo", "hello")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !strings.Contains(output, "hello") {
		t.Errorf("Expected output to contain 'hello', got: %s", output)
	}
}

func TestRunError(t *testing.T) {
	// Test command that doesn't exist
	_, err := Run("nonexistent-command-xyz")
	if err == nil {
		t.Error("Expected error for nonexistent command")
	}
}

func TestRunInDir(t *testing.T) {
	// Test running command in a specific directory
	output, err := RunInDir("/tmp", "pwd")
	if err != nil {
		t.Fatalf("RunInDir failed: %v", err)
	}
	if !strings.Contains(output, "/tmp") {
		t.Errorf("Expected output to contain '/tmp', got: %s", output)
	}
}

func TestIsCommandAvailable(t *testing.T) {
	// Test with a command that exists
	if !IsCommandAvailable("echo") {
		t.Error("Expected 'echo' to be available")
	}

	// Test with a command that doesn't exist
	if IsCommandAvailable("nonexistent-command-xyz") {
		t.Error("Expected 'nonexistent-command-xyz' to not be available")
	}
}

func TestRunStreaming(t *testing.T) {
	// Test streaming output (should not error)
	err := RunStreaming("echo", "test")
	if err != nil {
		t.Fatalf("RunStreaming failed: %v", err)
	}
}

func TestRunInDirStreaming(t *testing.T) {
	// Test streaming output in directory
	err := RunInDirStreaming("/tmp", "pwd")
	if err != nil {
		t.Fatalf("RunInDirStreaming failed: %v", err)
	}
}
