// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If a separate written commercial agreement exists, you may use this file under those terms instead.

package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"

	"github.com/Charles546/hd-cli/internal/util"
)

// Generator renders Honeydipper config files from templates using a WizardConfig.
type Generator struct {
	templates *template.Template
	funcMap    template.FuncMap
}

// NewGenerator creates a new Generator, loading and parsing all embedded templates.
func NewGenerator() *Generator {
	funcMap := sprig.TxtFuncMap()
	// Add custom template functions
	funcMap["hasStr"] = func(s string) bool { return strings.TrimSpace(s) != "" }

	g := &Generator{
		funcMap: funcMap,
	}

	// Parse all embedded templates
	entries, err := templatesFS.ReadDir("configs")
	if err != nil {
		// If no templates dir, create a minimal generator
		g.templates = template.New("empty").Funcs(funcMap)
		return g
	}

	// Build a template with all files
	root := template.New("root").Funcs(funcMap)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := templatesFS.ReadFile("configs/" + entry.Name())
		if err != nil {
			continue
		}
		t := root.New(entry.Name()).Funcs(funcMap)
		_, err = t.Parse(string(content))
		if err != nil {
			continue
		}
	}
	g.templates = root

	return g
}

// Generate renders all config files and writes them to the output directory.
// If dryRun is true, files are printed to stdout instead of writing to disk.
func (g *Generator) Generate(cfg *WizardConfig, outputDir string, dryRun bool) error {
	if cfg == nil {
		return fmt.Errorf("wizard config is nil")
	}

	// Apply defaults for any missing values
	applyDefaults(cfg)

	// Ensure output directory exists (unless dry run)
	if !dryRun {
		if err := util.EnsureDir(outputDir); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Define the files to generate
	files := []struct {
		name     string
		template string
		cond     bool
	}{
		{"init.yaml", "init.yaml.tmpl", true},
		{"integrations.yaml", "integrations.yaml.tmpl", true},
		{"drivers.yaml", "drivers.yaml.tmpl", true},
		{"daemon.yaml", "daemon.yaml.tmpl", true},
		{"workflows.yaml", "workflows.yaml.tmpl", true},
		{"ai/agents.yaml", "ai-agents.yaml.tmpl", cfg.AIEnabled},
		{"ai/contexts.yaml", "ai-contexts.yaml.tmpl", cfg.AIEnabled},
		{"ai/mcp.yaml", "ai-mcp.yaml.tmpl", cfg.AIEnabled},
		{"ai/engines.yaml", "ai-engines.yaml.tmpl", cfg.AIEnabled},
	}

	for _, f := range files {
		if !f.cond {
			continue
		}

		tmpl := g.templates.Lookup(f.template)
		if tmpl == nil {
			return fmt.Errorf("template %q not found", f.template)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, cfg); err != nil {
			return fmt.Errorf("failed to render %s: %w", f.name, err)
		}

		// Validate YAML before writing
		if err := ValidateYAML(buf.Bytes()); err != nil {
			return fmt.Errorf("generated %s has invalid YAML: %w", f.name, err)
		}

		if dryRun {
			fmt.Printf("===== %s =====\n", f.name)
			fmt.Println(buf.String())
			fmt.Println()
		} else {
			filePath := filepath.Join(outputDir, f.name)
			if err := util.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", f.name, err)
			}
		}
	}

	// Generate docker-compose.yaml for Docker mode
	if cfg.DeploymentMode == "docker" {
		tmpl := g.templates.Lookup("docker-compose.yaml.tmpl")
		if tmpl != nil {
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, cfg); err != nil {
				return fmt.Errorf("failed to render docker-compose.yaml: %w", err)
			}
			if dryRun {
				fmt.Printf("===== docker-compose.yaml =====\n")
				fmt.Println(buf.String())
			} else {
				filePath := filepath.Join(outputDir, "docker-compose.yaml")
				if err := util.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
					return fmt.Errorf("failed to write docker-compose.yaml: %w", err)
				}
			}
		}
	}

	return nil
}

// GenerateFromAnswersFile reads a YAML answers file and generates config without the TUI.
func (g *Generator) GenerateFromAnswersFile(answersPath, outputDir string, dryRun bool) error {
	data, err := os.ReadFile(answersPath)
	if err != nil {
		return fmt.Errorf("failed to read answers file: %w", err)
	}

	cfg := &WizardConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to parse answers file: %w", err)
	}

	// Apply defaults for any missing values
	applyDefaults(cfg)

	return g.Generate(cfg, outputDir, dryRun)
}

// applyDefaults fills in default values for fields that are empty.
func applyDefaults(cfg *WizardConfig) {
	def := NewDefaultWizardConfig()

	if cfg.ProjectName == "" {
		cfg.ProjectName = def.ProjectName
	}
	if cfg.ConfigDir == "" {
		cfg.ConfigDir = "./" + cfg.ProjectName
	}
	if abs, err := filepath.Abs(cfg.ConfigDir); err == nil {
		cfg.ConfigDirAbs = abs
	} else {
		cfg.ConfigDirAbs = cfg.ConfigDir
	}
	if cfg.DeploymentMode == "" {
		cfg.DeploymentMode = def.DeploymentMode
	}
	if cfg.EssentialsRepoURL == "" {
		cfg.EssentialsRepoURL = def.EssentialsRepoURL
	}
	if cfg.EssentialsBranch == "" {
		cfg.EssentialsBranch = def.EssentialsBranch
	}
	if cfg.BootstrapCloneCredentialType == "" {
		cfg.BootstrapCloneCredentialType = def.BootstrapCloneCredentialType
	}
	if cfg.EssentialsCloneType == "" {
		cfg.EssentialsCloneType = def.EssentialsCloneType
	}
	if cfg.SecretsBackend == "" {
		cfg.SecretsBackend = def.SecretsBackend
	}
	if cfg.RedisMode == "" {
		cfg.RedisMode = def.RedisMode
	}
	if cfg.DockerImageTag == "" {
		cfg.DockerImageTag = def.DockerImageTag
	}
	if cfg.DockerAPIPort == 0 {
		cfg.DockerAPIPort = def.DockerAPIPort
	}
	if cfg.DockerWebhookPort == 0 {
		cfg.DockerWebhookPort = def.DockerWebhookPort
	}
	if cfg.K8sNamespace == "" {
		cfg.K8sNamespace = def.K8sNamespace
	}
	if cfg.K8sRepoStrategy == "" {
		cfg.K8sRepoStrategy = def.K8sRepoStrategy
	}
	if cfg.SourceBranch == "" {
		cfg.SourceBranch = def.SourceBranch
	}
	if cfg.AIModel == "" {
		cfg.AIModel = def.AIModel
	}
	if cfg.AIBaseURL == "" {
		cfg.AIBaseURL = def.AIBaseURL
	}
	if cfg.AIEngineName == "" {
		cfg.AIEngineName = def.AIEngineName
	}
}

// ApplyDefaults fills in default values for fields that are empty.
// This is the exported version of applyDefaults for use by the CLI layer.
func ApplyDefaults(cfg *WizardConfig) {
	applyDefaults(cfg)
}
