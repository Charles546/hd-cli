// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If a separate written commercial agreement exists, you may use this file under those terms instead.

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/Charles546/hd-cli/internal/util"
)

// ValidationError represents a config validation issue.
type ValidationError struct {
	File    string `json:"file"`
	Line    int    `json:"line,omitempty"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (v *ValidationError) Error() string {
	if v.Line > 0 {
		return fmt.Sprintf("%s:%d: %s", v.File, v.Line, v.Message)
	}
	return fmt.Sprintf("%s: %s", v.File, v.Message)
}

// ValidationResult holds the outcome of validating a config directory.
type ValidationResult struct {
	Valid   bool              `json:"valid"`
	Errors  []ValidationError `json:"errors,omitempty"`
	Warnings []string          `json:"warnings,omitempty"`
}

// ValidateYAML checks if the given content is valid YAML.
// Returns nil if valid, or a ValidationError describing the syntax error.
func ValidateYAML(content []byte) error {
	var doc interface{}
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}
	return nil
}

// ValidateYAMLDetailed checks YAML and returns line-level error info.
func ValidateYAMLDetailed(fileName string, content []byte) []ValidationError {
	var doc interface{}
	if err := yaml.Unmarshal(content, &doc); err != nil {
		// Try to extract line number from yaml.TypeError
		var typeErr *yaml.TypeError
		if errorsAs(err, &typeErr) {
			var errs []ValidationError
			for _, e := range typeErr.Errors {
				line := 0
				// Parse line from error message: "line N: ..."
				_, _ = fmt.Sscanf(e, "line %d:", &line)
				errs = append(errs, ValidationError{
					File:    fileName,
					Line:    line,
					Message: e,
				})
			}
			return errs
		}
		return []ValidationError{
			{File: fileName, Message: err.Error()},
		}
	}
	return nil
}

// ValidateConfig checks that a config directory has all required files
// and that all YAML files are syntactically valid.
func ValidateConfig(dir string) ValidationResult {
	result := ValidationResult{Valid: true}

	// Required files for a minimal Honeydipper config
	requiredFiles := []string{
		"init.yaml",
		"integrations.yaml",
		"drivers.yaml",
		"daemon.yaml",
		"workflows.yaml",
	}

	// Check each required file
	for _, f := range requiredFiles {
		path := filepath.Join(dir, f)
		if !util.FileExists(path) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				File:    f,
				Message: "required file is missing",
			})
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				File:    f,
				Message: fmt.Sprintf("failed to read: %v", err),
			})
			continue
		}

		if err := ValidateYAML(content); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				File:    f,
				Message: fmt.Sprintf("invalid YAML: %v", err),
			})
		}
	}

	// Check optional files if they exist
	optionalFiles := []string{
		"ai/basic.yaml",
		"ai/engines.yaml",
		"docker-compose.yaml",
	}

	for _, f := range optionalFiles {
		path := filepath.Join(dir, f)
		if !util.FileExists(path) {
			continue
		}
		content, err := os.ReadFile(path)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("failed to read %s: %v", f, err))
			continue
		}
		if err := ValidateYAML(content); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("%s has invalid YAML: %v", f, err))
		}
	}

	// Check for recommended fields in init.yaml
	initPath := filepath.Join(dir, "init.yaml")
	if util.FileExists(initPath) {
		content, err := os.ReadFile(initPath)
		if err == nil {
			var initDoc map[string]interface{}
			if err := yaml.Unmarshal(content, &initDoc); err == nil {
				if _, hasRepo := initDoc["repo"]; !hasRepo {
					result.Warnings = append(result.Warnings, "init.yaml missing 'repo' field")
				}
				if _, hasInclude := initDoc["include"]; !hasInclude {
					if _, hasGlobs := initDoc["globs"]; !hasGlobs {
						result.Warnings = append(result.Warnings, "init.yaml missing 'include' or 'globs' field")
					}
				}
			}
		}
	}

	return result
}

// ValidateAnswersFile checks if a non-interactive answers YAML file is valid.
func ValidateAnswersFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read answers file: %w", err)
	}

	if err := ValidateYAML(content); err != nil {
		return fmt.Errorf("answers file has invalid YAML: %w", err)
	}

	// Check for recommended fields
	var doc map[string]interface{}
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return err
	}

	var missing []string
	for _, field := range []string{"project_name", "config_dir", "deployment_mode"} {
		if _, ok := doc[field]; !ok {
			missing = append(missing, field)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("answers file missing recommended fields: %s", strings.Join(missing, ", "))
	}

	return nil
}

// errorsAs is a helper to unwrap errors (simplified version of errors.As)
func errorsAs(err error, target interface{}) bool {
	// Simple type assertion for yaml.TypeError
	if te, ok := err.(*yaml.TypeError); ok {
		if t, ok := target.(**yaml.TypeError); ok {
			*t = te
			return true
		}
	}
	return false
}
