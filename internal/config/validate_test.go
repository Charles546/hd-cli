// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If a separate written commercial agreement exists, you may use this file under those terms instead.

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateYAML(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid YAML",
			content: "key: value\nlist:\n  - item1\n  - item2\n",
			wantErr: false,
		},
		{
			name:    "invalid YAML syntax",
			content: "key: value\n  bad_indent: true\n",
			wantErr: true,
		},
		{
			name:    "empty content",
			content: "",
			wantErr: false,
		},
		{
			name:    "valid complex YAML",
			content: "nested:\n  inner:\n    key: value\n",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateYAML([]byte(tt.content))
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	// Create a temp directory with valid config files
	tmpDir := t.TempDir()

	files := map[string]string{
		"init.yaml":         "repo:\n  - url: https://example.com\ninclude:\n  - '*.yaml'\n",
		"integrations.yaml": "systems:\n  github:\n    type: github\n",
		"drivers.yaml":      "redis:\n  type: redis\n  address: localhost\n",
		"daemon.yaml":       "daemon:\n  name: test\n",
		"workflows.yaml":    "workflows:\n  - name: test\n",
	}

	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}
	}

	result := ValidateConfig(tmpDir)
	if !result.Valid {
		t.Errorf("ValidateConfig() = false, want true; errors: %v", result.Errors)
	}
}

func TestValidateConfigMissingFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Only write some files
	if err := os.WriteFile(filepath.Join(tmpDir, "init.yaml"), []byte("repo: test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	result := ValidateConfig(tmpDir)
	if result.Valid {
		t.Error("ValidateConfig() should be invalid with missing files")
	}

	// Check that we got errors for missing files
	foundMissing := false
	for _, err := range result.Errors {
		if strings.Contains(err.Message, "missing") {
			foundMissing = true
		}
	}
	if !foundMissing {
		t.Error("expected missing file error")
	}
}

func TestValidateConfigInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	// Write a file with invalid YAML
	if err := os.WriteFile(filepath.Join(tmpDir, "init.yaml"), []byte("key: value\n  bad: indent\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "integrations.yaml"), []byte("valid: yaml\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "drivers.yaml"), []byte("valid: yaml\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "daemon.yaml"), []byte("valid: yaml\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "workflows.yaml"), []byte("valid: yaml\n"), 0644); err != nil {
		t.Fatal(err)
	}

	result := ValidateConfig(tmpDir)
	if result.Valid {
		t.Error("ValidateConfig() should be invalid with invalid YAML")
	}
}

func TestValidateAnswersFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid answers",
			content: "project_name: test\nconfig_dir: ./test\ndeployment_mode: docker\n",
			wantErr: false,
		},
		{
			name:    "invalid YAML",
			content: "key: value\n  bad: indent\n",
			wantErr: true,
		},
		{
			name:    "missing recommended fields",
			content: "project_name: test\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, "answers.yaml")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}
			err := ValidateAnswersFile(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAnswersFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAnswersFileNotFound(t *testing.T) {
	err := ValidateAnswersFile("/nonexistent/path.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
