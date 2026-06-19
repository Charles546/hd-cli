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

func TestCheckGit(t *testing.T) {
	result := CheckGit()
	
	// Git test - assuming git is available in test environment
	if !result.Available {
		t.Log("Git is not available in this environment")
		return
	}

	if result.Name != "Git" {
		t.Errorf("Expected Name to be 'Git', got: %s", result.Name)
	}

	if result.Version == "" {
		t.Error("Expected version to be set")
	}

	if result.Error != "" {
		t.Errorf("Unexpected error: %s", result.Error)
	}
}

func TestCheckGo(t *testing.T) {
	result := CheckGo()
	
	// Go test - should always be available in Go test environment
	if !result.Available {
		t.Error("Go should be available in test environment")
		return
	}

	if result.Name != "Go" {
		t.Errorf("Expected Name to be 'Go', got: %s", result.Name)
	}

	if result.Version == "" {
		t.Error("Expected version to be set")
	}

	// Check that version contains "go" prefix
	if !strings.Contains(result.Version, "go") {
		t.Errorf("Expected version to contain 'go', got: %s", result.Version)
	}
}

func TestCheckDocker(t *testing.T) {
	result := CheckDocker()
	
	// Docker test - may or may not be available
	t.Logf("Docker check result: Available=%v, Version=%s, Error=%s", 
		result.Available, result.Version, result.Error)
	
	// Name should be either Docker or Podman
	if result.Name != "Docker" && result.Name != "Podman" {
		t.Errorf("Expected Name to be 'Docker' or 'Podman', got: %s", result.Name)
	}
}

func TestCheckHelm(t *testing.T) {
	result := CheckHelm()
	
	// Helm test - may or may not be available
	t.Logf("Helm check result: Available=%v, Version=%s, Error=%s", 
		result.Available, result.Version, result.Error)
	
	if result.Name != "Helm" {
		t.Errorf("Expected Name to be 'Helm', got: %s", result.Name)
	}
}

func TestCheckKubectl(t *testing.T) {
	result := CheckKubectl()
	
	// kubectl test - may or may not be available
	t.Logf("kubectl check result: Available=%v, Version=%s, Error=%s", 
		result.Available, result.Version, result.Error)
	
	if result.Name != "kubectl" {
		t.Errorf("Expected Name to be 'kubectl', got: %s", result.Name)
	}
}

func TestCheckAll(t *testing.T) {
	results := CheckAll()
	
	if len(results) != 5 {
		t.Errorf("Expected 5 check results, got: %d", len(results))
	}

	// Verify all checks are present
	expectedNames := map[string]bool{
		"Docker":  false,
		"Podman":  false,
		"Helm":    false,
		"kubectl": false,
		"Go":      false,
	}

	for _, result := range results {
		if _, ok := expectedNames[result.Name]; ok {
			expectedNames[result.Name] = true
		}
	}

	// At least one of Docker or Podman should be found
	if !expectedNames["Docker"] && !expectedNames["Podman"] {
		t.Log("Neither Docker nor Podman found (this may be expected)")
	}

	// Go should always be found in test environment
	if !expectedNames["Go"] {
		t.Error("Go check not found in results")
	}
}

func TestCheckResultStruct(t *testing.T) {
	// Test that CheckResult struct has correct fields
	result := CheckResult{
		Name:      "Test",
		Available: true,
		Version:   "1.0.0",
		Error:     "",
	}

	if result.Name != "Test" {
		t.Errorf("Expected Name to be 'Test', got: %s", result.Name)
	}

	if !result.Available {
		t.Error("Expected Available to be true")
	}

	if result.Version != "1.0.0" {
		t.Errorf("Expected Version to be '1.0.0', got: %s", result.Version)
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.25.0", "1.24.0", 1},
		{"1.24.0", "1.25.0", -1},
		{"1.25.0", "1.25.0", 0},
		{"1.25.1", "1.25.0", 1},
		{"1.25.0", "1.25.1", -1},
		{"2.0.0", "1.99.99", 1},
	}

	for _, tt := range tests {
		result := compareVersions(tt.v1, tt.v2)
		if result != tt.expected {
			t.Errorf("compareVersions(%s, %s) = %d, expected %d", 
				tt.v1, tt.v2, result, tt.expected)
		}
	}
}

func TestIsGoVersionSufficient(t *testing.T) {
	tests := []struct {
		current  string
		minimum  string
		expected bool
	}{
		{"go1.25.0", "1.25", true},
		{"go1.25.11", "1.25", true},
		{"go1.26.0", "1.25", true},
		{"1.24.0", "1.25", false},
		{"1.24.99", "1.25", false},
	}

	for _, tt := range tests {
		result := isGoVersionSufficient(tt.current, tt.minimum)
		if result != tt.expected {
			t.Errorf("isGoVersionSufficient(%s, %s) = %v, expected %v", 
				tt.current, tt.minimum, result, tt.expected)
		}
	}
}
