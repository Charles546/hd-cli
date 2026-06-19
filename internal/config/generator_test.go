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

func TestNewGenerator(t *testing.T) {
	g := NewGenerator()
	if g == nil {
		t.Fatal("NewGenerator returned nil")
	}
	if g.templates == nil {
		t.Fatal("generator template is nil")
	}
}

func TestGenerateDocker(t *testing.T) {
	g := NewGenerator()
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-docker"
	cfg.DeploymentMode = "docker"

	tmpDir := t.TempDir()
	err := g.Generate(cfg, tmpDir, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check required files exist
	requiredFiles := []string{
		"init.yaml",
		"integrations.yaml",
		"drivers.yaml",
		"daemon.yaml",
		"workflows.yaml",
		"docker-compose.yaml",
	}
	for _, f := range requiredFiles {
		path := filepath.Join(tmpDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("required file not found: %s", f)
		}
	}

	// AI files should NOT exist (AI disabled by default)
	aiFiles := []string{"ai/basic.yaml", "ai/engines.yaml"}
	for _, f := range aiFiles {
		path := filepath.Join(tmpDir, f)
		if _, err := os.Stat(path); err == nil {
			t.Errorf("AI file should not exist when AI disabled: %s", f)
		}
	}
}

func TestGenerateWithAI(t *testing.T) {
	g := NewGenerator()
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-ai"
	cfg.AIEnabled = true
	cfg.AIAPIKeyPath = "secrets/openai-key"
	cfg.AIBaseURL = "https://api.openai.com/v1"
	cfg.AIModel = "gpt-4o"
	cfg.AIEngineName = "default"

	tmpDir := t.TempDir()
	err := g.Generate(cfg, tmpDir, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// AI files should exist
	aiFiles := []string{"ai/basic.yaml", "ai/engines.yaml"}
	for _, f := range aiFiles {
		path := filepath.Join(tmpDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("AI file not found: %s", f)
		}
	}

	// Check AI content
	enginesPath := filepath.Join(tmpDir, "ai", "engines.yaml")
	content, err := os.ReadFile(enginesPath)
	if err != nil {
		t.Fatalf("failed to read ai/engines.yaml: %v", err)
	}
	if !strings.Contains(string(content), "gpt-4o") {
		t.Errorf("ai/engines.yaml does not contain model name")
	}
	if !strings.Contains(string(content), "openaiApiKey") {
		t.Errorf("ai/engines.yaml does not contain API key reference")
	}
}

func TestGenerateDryRun(t *testing.T) {
	g := NewGenerator()
	cfg := NewDefaultWizardConfig()

	// Dry run should not create any files
	tmpDir := filepath.Join(t.TempDir(), "nonexistent")
	err := g.Generate(cfg, tmpDir, true)
	if err != nil {
		t.Fatalf("Generate dry-run failed: %v", err)
	}

	// Directory should not have been created
	if _, err := os.Stat(tmpDir); err == nil {
		t.Error("dry-run should not create output directory")
	}
}

func TestGenerateKubernetes(t *testing.T) {
	g := NewGenerator()
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-k8s"
	cfg.DeploymentMode = "kubernetes"
	cfg.K8sNamespace = "honeydipper-prod"
	cfg.K8sRepoStrategy = "configmap"

	tmpDir := t.TempDir()
	err := g.Generate(cfg, tmpDir, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check daemon.yaml has k8s-relevant settings
	daemonPath := filepath.Join(tmpDir, "daemon.yaml")
	content, err := os.ReadFile(daemonPath)
	if err != nil {
		t.Fatalf("failed to read daemon.yaml: %v", err)
	}
	if !strings.Contains(string(content), "kubernetes") {
		t.Errorf("daemon.yaml does not reference kubernetes mode")
	}
}

func TestGenerateSource(t *testing.T) {
	g := NewGenerator()
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-source"
	cfg.DeploymentMode = "source"
	cfg.SourceClonePath = "/opt/honeydipper"
	cfg.SourceBranch = "v4"

	tmpDir := t.TempDir()
	err := g.Generate(cfg, tmpDir, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// docker-compose.yaml should NOT exist for source mode
	composePath := filepath.Join(tmpDir, "docker-compose.yaml")
	if _, err := os.Stat(composePath); err == nil {
		t.Error("docker-compose.yaml should not exist for source mode")
	}
}

func TestGenerateFromAnswersFile(t *testing.T) {
	g := NewGenerator()
	tmpDir := t.TempDir()

	answersPath := "../../test-fixtures/answers.yaml"
	err := g.GenerateFromAnswersFile(answersPath, tmpDir, false)
	if err != nil {
		t.Fatalf("GenerateFromAnswersFile failed: %v", err)
	}

	// Check files were generated
	requiredFiles := []string{
		"init.yaml",
		"integrations.yaml",
		"drivers.yaml",
		"daemon.yaml",
		"workflows.yaml",
	}
	for _, f := range requiredFiles {
		path := filepath.Join(tmpDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("required file not found: %s", f)
		}
	}
}

func TestApplyDefaults(t *testing.T) {
	cfg := &WizardConfig{
		ProjectName: "minimal",
	}
	applyDefaults(cfg)

	if cfg.DeploymentMode == "" {
		t.Error("DeploymentMode should have default")
	}
	if cfg.EssentialsRepoURL == "" {
		t.Error("EssentialsRepoURL should have default")
	}
	if cfg.DockerImageTag == "" {
		t.Error("DockerImageTag should have default")
	}
	if cfg.DockerAPIPort == 0 {
		t.Error("DockerAPIPort should have default")
	}
	if cfg.BootstrapCloneCredentialType != "none" {
		t.Errorf("BootstrapCloneCredentialType default should be 'none', got %q", cfg.BootstrapCloneCredentialType)
	}
}

func TestGenerateNilConfig(t *testing.T) {
	g := NewGenerator()
	err := g.Generate(nil, t.TempDir(), false)
	if err == nil {
		t.Error("expected error for nil config")
	}
}

// TestInitYamlUsesPluralReposAndIncludes verifies Fix 1: init.yaml uses "repos" and "includes" (plural).
func TestInitYamlUsesPluralReposAndIncludes(t *testing.T) {
	g := NewGenerator()
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-plural"
	cfg.EssentialsRepoURL = "https://github.com/example/essentials.git"
	cfg.EssentialsBranch = "main"

	tmpDir := t.TempDir()
	err := g.Generate(cfg, tmpDir, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	initPath := filepath.Join(tmpDir, "init.yaml")
	content, err := os.ReadFile(initPath)
	if err != nil {
		t.Fatalf("failed to read init.yaml: %v", err)
	}
	yamlStr := string(content)

	// Must use "repos:" (plural), not "repo:" (singular)
	if !strings.Contains(yamlStr, "repos:") {
		t.Error("init.yaml should contain 'repos:' (plural)")
	}
	if strings.Contains(yamlStr, "\nrepo:") {
		t.Error("init.yaml should not contain 'repo:' (singular)")
	}

	// Must use "includes:" (plural), not "include:" (singular)
	if !strings.Contains(yamlStr, "includes:") {
		t.Error("init.yaml should contain 'includes:' (plural)")
	}
	if strings.Contains(yamlStr, "\ninclude:") {
		t.Error("init.yaml should not contain 'include:' (singular)")
	}
}

// TestLookupStringsHaveNoSpaces verifies Fix 3: LOOKUP strings have no spaces after commas.
func TestLookupStringsHaveNoSpaces(t *testing.T) {
	g := NewGenerator()
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-lookup"
	cfg.AIEnabled = true

	tmpDir := t.TempDir()
	err := g.Generate(cfg, tmpDir, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check all generated YAML files for LOOKUP strings with spaces after commas
	files := []string{
		"init.yaml",
		"integrations.yaml",
		"drivers.yaml",
		"daemon.yaml",
		"workflows.yaml",
		"ai/basic.yaml",
		"ai/engines.yaml",
	}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		content, err := os.ReadFile(path)
		if err != nil {
			continue // file might not exist (e.g., ai files if AI disabled)
		}
		yamlStr := string(content)

		// Check for LOOKUP[vault, pattern (space after comma) - this is wrong
		if strings.Contains(yamlStr, "LOOKUP[vault, ") {
			t.Errorf("%s contains LOOKUP with space after comma", f)
		}
		if strings.Contains(yamlStr, "LOOKUP[secure_exec, ") {
			t.Errorf("%s contains LOOKUP with space after comma", f)
		}

		// Verify LOOKUP strings use no-space format: LOOKUP[driver,path]
		// Find all LOOKUP[ occurrences and ensure no space after comma
		lines := strings.Split(yamlStr, "\n")
		for lineNum, line := range lines {
			if strings.Contains(line, "LOOKUP[") {
				// Extract the LOOKUP content
				start := strings.Index(line, "LOOKUP[")
				end := strings.Index(line[start:], "]")
				if end > 0 {
					lookupContent := line[start+7 : start+end]
					if strings.Contains(lookupContent, ", ") {
						t.Errorf("%s line %d: LOOKUP content %q has space after comma", f, lineNum+1, lookupContent)
					}
				}
			}
		}
	}
}

// TestBootstrapCloneCredentialsInInitYaml verifies Fix 2: credentials in init.yaml.
func TestBootstrapCloneCredentialsInInitYaml(t *testing.T) {
	// Test with PAT credential type
	t.Run("pat credentials", func(t *testing.T) {
		g := NewGenerator()
		cfg := NewDefaultWizardConfig()
		cfg.ProjectName = "test-pat"
		cfg.BootstrapCloneCredentialType = "pat"
		cfg.BootstrapClonePassEnv = "DIPPER_GIT_PASS"

		tmpDir := t.TempDir()
		err := g.Generate(cfg, tmpDir, false)
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		initPath := filepath.Join(tmpDir, "init.yaml")
		content, err := os.ReadFile(initPath)
		if err != nil {
			t.Fatalf("failed to read init.yaml: %v", err)
		}
		yamlStr := string(content)

		if !strings.Contains(yamlStr, "pass_env:") {
			t.Error("init.yaml should contain 'pass_env:' for PAT credential type")
		}
		if !strings.Contains(yamlStr, "DIPPER_GIT_PASS") {
			t.Error("init.yaml should reference DIPPER_GIT_PASS env var")
		}
	})

	// Test with GitHub App credential type
	t.Run("github_app credentials", func(t *testing.T) {
		g := NewGenerator()
		cfg := NewDefaultWizardConfig()
		cfg.ProjectName = "test-ghapp"
		cfg.BootstrapCloneCredentialType = "github_app"

		tmpDir := t.TempDir()
		err := g.Generate(cfg, tmpDir, false)
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		initPath := filepath.Join(tmpDir, "init.yaml")
		content, err := os.ReadFile(initPath)
		if err != nil {
			t.Fatalf("failed to read init.yaml: %v", err)
		}
		yamlStr := string(content)

		if !strings.Contains(yamlStr, "token_source: github") {
			t.Error("init.yaml should contain 'token_source: github' for github_app credential type")
		}
	})

	// Test with no credentials (public repo)
	t.Run("no credentials", func(t *testing.T) {
		g := NewGenerator()
		cfg := NewDefaultWizardConfig()
		cfg.ProjectName = "test-none"
		cfg.BootstrapCloneCredentialType = "none"

		tmpDir := t.TempDir()
		err := g.Generate(cfg, tmpDir, false)
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		initPath := filepath.Join(tmpDir, "init.yaml")
		content, err := os.ReadFile(initPath)
		if err != nil {
			t.Fatalf("failed to read init.yaml: %v", err)
		}
		yamlStr := string(content)

		if strings.Contains(yamlStr, "token_source:") {
			t.Error("init.yaml should not contain 'token_source:' when credential type is 'none'")
		}
		if strings.Contains(yamlStr, "pass_env:") {
			t.Error("init.yaml should not contain 'pass_env:' when credential type is 'none'")
		}
	})
}

// TestBootstrapCloneCredentialsInDockerCompose verifies Fix 2: credentials in docker-compose.yaml.
func TestBootstrapCloneCredentialsInDockerCompose(t *testing.T) {
	// Test with PAT credential type
	t.Run("pat credentials in docker-compose", func(t *testing.T) {
		g := NewGenerator()
		cfg := NewDefaultWizardConfig()
		cfg.ProjectName = "test-dc-pat"
		cfg.DeploymentMode = "docker"
		cfg.BootstrapCloneCredentialType = "pat"
		cfg.BootstrapClonePassEnv = "DIPPER_GIT_PASS"

		tmpDir := t.TempDir()
		err := g.Generate(cfg, tmpDir, false)
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		composePath := filepath.Join(tmpDir, "docker-compose.yaml")
		content, err := os.ReadFile(composePath)
		if err != nil {
			t.Fatalf("failed to read docker-compose.yaml: %v", err)
		}
		yamlStr := string(content)

		if !strings.Contains(yamlStr, "DIPPER_GIT_PASS=${DIPPER_GIT_PASS}") {
			t.Error("docker-compose.yaml should pass through DIPPER_GIT_PASS env var")
		}
	})

	// Test with GitHub App credential type
	t.Run("github_app credentials in docker-compose", func(t *testing.T) {
		g := NewGenerator()
		cfg := NewDefaultWizardConfig()
		cfg.ProjectName = "test-dc-ghapp"
		cfg.DeploymentMode = "docker"
		cfg.BootstrapCloneCredentialType = "github_app"
		cfg.GithubAppID = "12345"
		cfg.GithubInstallationID = "67890"

		tmpDir := t.TempDir()
		err := g.Generate(cfg, tmpDir, false)
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		composePath := filepath.Join(tmpDir, "docker-compose.yaml")
		content, err := os.ReadFile(composePath)
		if err != nil {
			t.Fatalf("failed to read docker-compose.yaml: %v", err)
		}
		yamlStr := string(content)

		if !strings.Contains(yamlStr, "GH_APP_ID=${GH_APP_ID}") {
			t.Error("docker-compose.yaml should pass through GH_APP_ID env var")
		}
		if !strings.Contains(yamlStr, "GH_APP_INSTALLATION_ID=${GH_APP_INSTALLATION_ID}") {
			t.Error("docker-compose.yaml should pass through GH_APP_INSTALLATION_ID env var")
		}
		if !strings.Contains(yamlStr, "GH_APP_KEY=${GH_APP_KEY}") {
			t.Error("docker-compose.yaml should pass through GH_APP_KEY env var")
		}
	})

	// Test with no credentials
	t.Run("no credentials in docker-compose", func(t *testing.T) {
		g := NewGenerator()
		cfg := NewDefaultWizardConfig()
		cfg.ProjectName = "test-dc-none"
		cfg.DeploymentMode = "docker"
		cfg.BootstrapCloneCredentialType = "none"

		tmpDir := t.TempDir()
		err := g.Generate(cfg, tmpDir, false)
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		composePath := filepath.Join(tmpDir, "docker-compose.yaml")
		content, err := os.ReadFile(composePath)
		if err != nil {
			t.Fatalf("failed to read docker-compose.yaml: %v", err)
		}
		yamlStr := string(content)

		if strings.Contains(yamlStr, "DIPPER_GIT_PASS") {
			t.Error("docker-compose.yaml should not contain DIPPER_GIT_PASS when credential type is 'none'")
		}
		if strings.Contains(yamlStr, "GH_APP_") {
			t.Error("docker-compose.yaml should not contain GH_APP_ vars when credential type is 'none'")
		}
	})
}
