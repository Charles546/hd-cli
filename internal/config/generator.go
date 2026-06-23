// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If a separate written commercial agreement exists, you may use this file under those terms instead.

package config

import (
	"bytes"
	"fmt"
	"os/exec"
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
	funcMap["envRef"] = func(name string) string { return "{% .env." + name + " %}" }
	funcMap["hasNewline"] = func(s string) bool { return strings.Contains(s, "\n") }
	funcMap["yamlEscape"] = func(s string) string {
		s = strings.ReplaceAll(s, "\\", "\\\\")
		s = strings.ReplaceAll(s, "\n", "\\n")
		s = strings.ReplaceAll(s, "\"", "\\\"")
		return s
	}

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

	// Populate DevEnvVars for dev mode docker-compose env passthrough
	if cfg.SecretsBackend == "dev" {
		cfg.DevEnvVars = cfg.CollectDevEnvVars()
	}

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

	// Build config repo clone env vars for the docker-compose template
	cfg.ConfigRepoCloneEnvVars = cfg.BuildConfigRepoCloneEnvVars()

	// Run git init and git remote add if user chose to create a repo
	if cfg.GithubCreateRepo {
		if err := runGitInit(outputDir); err != nil {
			return fmt.Errorf("failed to run git init: %w", err)
		}
		cfg.GitInit = true

		if cfg.GitRemoteURL != "" {
			if err := runGitRemoteAdd(outputDir, cfg.GitRemoteURL); err != nil {
				return fmt.Errorf("failed to run git remote add: %w", err)
			}
		}
	}

	// Generate .env file for docker deployments with non-source modes
	// Contains placeholder values for sensitive data like PATs and SSH keys
	if !dryRun && cfg.DeploymentMode != "source" {
		envContent := generateEnvFile(cfg)
		if envContent != "" {
			envPath := filepath.Join(outputDir, ".env")
			if err := util.WriteFile(envPath, []byte(envContent), 0644); err != nil {
				return fmt.Errorf("failed to write .env file: %w", err)
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


// runGitInit runs git init in the given directory.
func runGitInit(dir string) error {
	cmd := exec.Command("git", "init", dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run git init in %s: %w", dir, err)
	}
	return nil
}

// runGitRemoteAdd adds or updates the "origin" remote in the given directory.
// If the remote already exists, it updates the URL with `git remote set-url`;
// otherwise it adds a new remote with `git remote add`.
func runGitRemoteAdd(dir, url string) error {
	// Check if the "origin" remote already exists
	checkCmd := exec.Command("git", "-C", dir, "remote", "get-url", "origin")
	if err := checkCmd.Run(); err == nil {
		// Remote exists, update the URL
		setCmd := exec.Command("git", "-C", dir, "remote", "set-url", "origin", url)
		setCmd.Stdout = os.Stdout
		setCmd.Stderr = os.Stderr
		if err := setCmd.Run(); err != nil {
			return fmt.Errorf("failed to run git remote set-url origin %s in %s: %w", url, dir, err)
		}
		return nil
	}

	// Remote does not exist, add it
	addCmd := exec.Command("git", "-C", dir, "remote", "add", "origin", url)
	addCmd.Stdout = os.Stdout
	addCmd.Stderr = os.Stderr
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("failed to run git remote add origin %s in %s: %w", url, dir, err)
	}
	return nil
}

// generateEnvFile creates a .env file content with placeholder values
// for environment variables used by clone authentication.
func generateEnvFile(cfg *WizardConfig) string {
	if cfg == nil {
		return ""
	}

	envVars := cfg.BuildConfigRepoCloneEnvVars()
	if len(envVars) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, "# Environment variables for Honeydipper config repo clone authentication")
	lines = append(lines, "# NOTE: Replace the placeholder values with actual secrets before deploying.")
	lines = append(lines, "")
	for k, v := range envVars {
		lines = append(lines, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(lines, "\n") + "\n"
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
	if cfg.ConfigRepoCloneAuth == "" {
		cfg.ConfigRepoCloneAuth = def.ConfigRepoCloneAuth
	}
}

// ApplyDefaults fills in default values for fields that are empty.
// This is the exported version of applyDefaults for use by the CLI layer.
func ApplyDefaults(cfg *WizardConfig) {
	applyDefaults(cfg)
}

