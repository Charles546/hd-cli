// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If a separate written commercial agreement exists, you may use this file under those terms instead.

package config

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed configs/*
var templatesFS embed.FS

// templatesDir is the embedded directory containing all config templates.
const templatesDir = "configs"

// GetTemplate returns the content of a named template file.
func GetTemplate(name string) ([]byte, error) {
	content, err := templatesFS.ReadFile(templatesDir + "/" + name)
	if err != nil {
		return nil, fmt.Errorf("template %q not found: %w", name, err)
	}
	return content, nil
}

// ListTemplates returns the names of all available template files.
func ListTemplates() ([]string, error) {
	entries, err := fs.ReadDir(templatesFS, templatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}
