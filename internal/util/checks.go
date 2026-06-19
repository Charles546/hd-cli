// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If you have a separate written commercial agreement, you may use this file under those terms instead.

package util

import (
	"fmt"
	"strings"
)

// CheckResult represents the result of a prerequisite check
type CheckResult struct {
	Name      string `json:"name"`
	Available bool   `json:"available"`
	Version   string `json:"version,omitempty"`
	Error     string `json:"error,omitempty"`
}

// CheckDocker detects Docker or Podman and verifies daemon is running
func CheckDocker() CheckResult {
	result := CheckResult{Name: "Docker"}

	// Try Docker first
	if IsCommandAvailable("docker") {
		result.Available = true
		version, err := getDockerVersion()
		if err != nil {
			result.Error = err.Error()
			// Still available but daemon might not be running
		}
		result.Version = version

		// Check if daemon is running
		_, err = Run("docker", "info")
		if err != nil {
			result.Error = "Docker daemon is not running"
			result.Available = false
		}

		return result
	}

	// Try Podman
	if IsCommandAvailable("podman") {
		result.Available = true
		result.Name = "Podman"
		version, err := getPodmanVersion()
		if err != nil {
			result.Error = err.Error()
		}
		result.Version = version

		// Check if daemon is running
		_, err = Run("podman", "info")
		if err != nil {
			result.Error = "Podman daemon is not running"
			result.Available = false
		}

		return result
	}

	result.Error = "Neither Docker nor Podman found"
	return result
}

// CheckHelm detects Helm and returns version
func CheckHelm() CheckResult {
	result := CheckResult{Name: "Helm"}

	if !IsCommandAvailable("helm") {
		result.Error = "Helm not found"
		return result
	}

	result.Available = true
	output, err := Run("helm", "version", "--short")
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.Version = strings.TrimSpace(output)
	return result
}

// CheckKubectl detects kubectl and verifies cluster connectivity
func CheckKubectl() CheckResult {
	result := CheckResult{Name: "kubectl"}

	if !IsCommandAvailable("kubectl") {
		result.Error = "kubectl not found"
		return result
	}

	result.Available = true

	// Get current context
	ctxOutput, err := Run("kubectl", "config", "current-context")
	if err != nil {
		result.Error = "Failed to get current context: " + err.Error()
		return result
	}
	currentContext := strings.TrimSpace(ctxOutput)

	// Check cluster connectivity
	_, err = Run("kubectl", "cluster-info")
	if err != nil {
		result.Error = "Cannot connect to cluster"
		return result
	}

	result.Version = fmt.Sprintf("Context: %s", currentContext)
	return result
}

// CheckGit detects Git and returns version
func CheckGit() CheckResult {
	result := CheckResult{Name: "Git"}

	if !IsCommandAvailable("git") {
		result.Error = "Git not found"
		return result
	}

	result.Available = true
	output, err := Run("git", "--version")
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.Version = strings.TrimSpace(output)
	return result
}

// CheckGo detects Go and returns version (must be >= 1.25)
func CheckGo() CheckResult {
	result := CheckResult{Name: "Go"}

	if !IsCommandAvailable("go") {
		result.Error = "Go not found"
		return result
	}

	result.Available = true
	output, err := Run("go", "version")
	if err != nil {
		result.Error = err.Error()
		return result
	}

	// go version returns "go version go1.25.11 linux/amd64"
	// Extract just the version number
	versionStr := strings.TrimSpace(output)
	result.Version = versionStr

	// Parse out the actual version number
	// Format: "go version go1.25.11 linux/amd64"
	parts := strings.Fields(versionStr)
	var versionNum string
	for _, part := range parts {
		if strings.HasPrefix(part, "go") && len(part) > 2 {
			versionNum = strings.TrimPrefix(part, "go")
			break
		}
	}

	if versionNum == "" {
		result.Error = "Could not parse Go version"
		result.Available = false
		return result
	}

	// Check version requirement
	if !isGoVersionSufficient(versionNum, "1.25") {
		result.Error = fmt.Sprintf("Go version must be >= 1.25, found: %s", versionNum)
		result.Available = false
	}

	return result
}

// CheckAll runs all prerequisite checks and returns structured summary
func CheckAll() []CheckResult {
	return []CheckResult{
		CheckDocker(),
		CheckHelm(),
		CheckKubectl(),
		CheckGit(),
		CheckGo(),
	}
}

// Helper functions

func getDockerVersion() (string, error) {
	output, err := Run("docker", "--version")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func getPodmanVersion() (string, error) {
	output, err := Run("podman", "--version")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// isGoVersionSufficient checks if the Go version meets the minimum requirement
func isGoVersionSufficient(current, minimum string) bool {
	// Simple version comparison
	return compareVersions(current, minimum) >= 0
}

// compareVersions compares two semantic version strings
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 string
		if i < len(parts1) {
			p1 = parts1[i]
		}
		if i < len(parts2) {
			p2 = parts2[i]
		}

		// Compare numeric parts
		if p1 < p2 {
			return -1
		}
		if p1 > p2 {
			return 1
		}
	}

	return 0
}
