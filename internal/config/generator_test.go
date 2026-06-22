// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If a separate written commercial agreement exists, you may use this file under those terms instead.

package config

import (
	"os"
	"os/exec"
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
	aiFiles := []string{"ai/agents.yaml", "ai/contexts.yaml", "ai/mcp.yaml", "ai/engines.yaml"}
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
	aiFiles := []string{"ai/agents.yaml", "ai/contexts.yaml", "ai/mcp.yaml", "ai/engines.yaml"}
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
	if !strings.Contains(string(content), "openrouter/owl-alpha") {
		t.Errorf("ai/engines.yaml does not contain expected model name")
	}
	if !strings.Contains(string(content), "openrouter.ai") {
		t.Errorf("ai/engines.yaml does not contain expected base URL")
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
	if !strings.Contains(string(content), "daemon") {
		t.Errorf("daemon.yaml does not reference daemon driver")
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
	if cfg.EssentialsCloneType != "none" {
		t.Errorf("EssentialsCloneType default should be 'none', got %q", cfg.EssentialsCloneType)
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

// TestInitYamlUsesPluralReposAndIncludes verifies init.yaml uses "repos" and "includes" (plural).
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

// TestLookupStringsHaveNoSpaces verifies LOOKUP strings have no spaces after commas.
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
		"ai/agents.yaml",
		"ai/contexts.yaml",
		"ai/mcp.yaml",
		"ai/engines.yaml",
	}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		content, err := os.ReadFile(path)
		if err != nil {
			continue // file might not exist
		}
		yamlStr := string(content)

		// Check for LOOKUP[vault, pattern (space after comma) - this is wrong
		if strings.Contains(yamlStr, "LOOKUP[vault, ") {
			t.Errorf("%s contains LOOKUP with space after comma", f)
		}

		// Find all LOOKUP[ occurrences and ensure no space after comma
		lines := strings.Split(yamlStr, "\n")
		for lineNum, line := range lines {
			if strings.Contains(line, "LOOKUP[") {
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

// TestBootstrapCloneCredentialsInInitYaml verifies credentials in init.yaml.
func TestBootstrapCloneCredentialsInInitYaml(t *testing.T) {
	// Test with PAT credential type
	t.Run("pat credentials", func(t *testing.T) {
		g := NewGenerator()
		cfg := NewDefaultWizardConfig()
		cfg.ProjectName = "test-pat"
		cfg.EssentialsCloneType = "pat"
		cfg.EssentialsClonePAT = "DIPPER_GIT_PAT"

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
		if !strings.Contains(yamlStr, "DIPPER_GIT_PAT") {
			t.Error("init.yaml should reference DIPPER_GIT_PAT env var")
		}
	})

	// Test with GitHub App credential type
	t.Run("github_app credentials", func(t *testing.T) {
		g := NewGenerator()
		cfg := NewDefaultWizardConfig()
		cfg.ProjectName = "test-ghapp"
		cfg.EssentialsCloneType = "github_app"

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
		cfg.EssentialsCloneType = "none"

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

	// Test with SSH credential type
	t.Run("ssh credentials", func(t *testing.T) {
		g := NewGenerator()
		cfg := NewDefaultWizardConfig()
		cfg.ProjectName = "test-ssh"
		cfg.EssentialsCloneType = "ssh"
		cfg.EssentialsCloneKey = "/home/user/.ssh/id_rsa"
		cfg.EssentialsCloneKeyPassEnv = "SSH_KEY_PASS"

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

		if !strings.Contains(yamlStr, "key_file:") {
			t.Error("init.yaml should contain 'key_file:' for SSH credential type")
		}
		if !strings.Contains(yamlStr, "/home/user/.ssh/id_rsa") {
			t.Error("init.yaml should contain the SSH key file path")
		}
		if !strings.Contains(yamlStr, "key_pass_env:") {
			t.Error("init.yaml should contain 'key_pass_env:' when key pass env is set")
		}
	})
}

// TestBootstrapCloneCredentialsInDockerCompose verifies credentials in docker-compose.yaml.
func TestBootstrapCloneCredentialsInDockerCompose(t *testing.T) {
	// Test with PAT credential type
	t.Run("pat credentials in docker-compose", func(t *testing.T) {
		g := NewGenerator()
		cfg := NewDefaultWizardConfig()
		cfg.ProjectName = "test-dc-pat"
		cfg.DeploymentMode = "docker"
		cfg.BootstrapCloneCredentialType = "pat"

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

		if !strings.Contains(yamlStr, "DIPPER_GIT_PAT=${DIPPER_GIT_PAT}") {
			t.Error("docker-compose.yaml should pass through DIPPER_GIT_PAT env var")
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

		if strings.Contains(yamlStr, "DIPPER_GIT_PAT") {
			t.Error("docker-compose.yaml should not contain DIPPER_GIT_PAT when credential type is 'none'")
		}
		if strings.Contains(yamlStr, "GH_APP_") {
			t.Error("docker-compose.yaml should not contain GH_APP_ vars when credential type is 'none'")
		}
	})
}

// TestAIEnabledIncludesAIYaml verifies that AI yaml files are generated when AI is enabled.
func TestAIEnabledIncludesAIYaml(t *testing.T) {
	g := NewGenerator()
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-ai-yaml"
	cfg.AIEnabled = true
	cfg.AIEngineName = "default"

	tmpDir := t.TempDir()
	err := g.Generate(cfg, tmpDir, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check ai/agents.yaml exists and has content
	agentsPath := filepath.Join(tmpDir, "ai", "agents.yaml")
	content, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("failed to read ai/agents.yaml: %v", err)
	}
	if !strings.Contains(string(content), "coordinator") {
		t.Error("ai/agents.yaml should contain coordinator agent definition")
	}
	if !strings.Contains(string(content), "developer") {
		t.Error("ai/agents.yaml should contain developer agent definition")
	}
	if !strings.Contains(string(content), "architect") {
		t.Error("ai/agents.yaml should contain architect agent definition")
	}

	// Check ai/mcp.yaml exists
	mcpPath := filepath.Join(tmpDir, "ai", "mcp.yaml")
	content, err = os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("failed to read ai/mcp.yaml: %v", err)
	}
	if !strings.Contains(string(content), "mcp") {
		t.Error("ai/mcp.yaml should contain MCP server configuration")
	}

	// Check ai/contexts.yaml exists
	contextsPath := filepath.Join(tmpDir, "ai", "contexts.yaml")
	content, err = os.ReadFile(contextsPath)
	if err != nil {
		t.Fatalf("failed to read ai/contexts.yaml: %v", err)
	}
	if !strings.Contains(string(content), "contexts") {
		t.Error("ai/contexts.yaml should contain contexts configuration")
	}
}

// TestInitYamlConditionalPath verifies that init.yaml includes path only when EssentialsPath is set.
func TestInitYamlConditionalPath(t *testing.T) {
	g := NewGenerator()

	// Without path
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-no-path"

	tmpDir := t.TempDir()
	err := g.Generate(cfg, tmpDir, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	initPath := filepath.Join(tmpDir, "init.yaml")
	content, _ := os.ReadFile(initPath)
	if strings.Contains(string(content), "path:") {
		t.Error("init.yaml should not contain 'path:' when EssentialsPath is empty")
	}

	// With path
	cfg2 := NewDefaultWizardConfig()
	cfg2.ProjectName = "test-with-path"
	cfg2.EssentialsPath = "essentials"

	tmpDir2 := t.TempDir()
	err = g.Generate(cfg2, tmpDir2, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	initPath2 := filepath.Join(tmpDir2, "init.yaml")
	content2, _ := os.ReadFile(initPath2)
	if !strings.Contains(string(content2), "path:") {
		t.Error("init.yaml should contain 'path:' when EssentialsPath is set")
	}
}

// TestInitYamlAIIncludes verifies that init.yaml includes ai/*.yaml only when AI is enabled.
func TestInitYamlAIIncludes(t *testing.T) {
	g := NewGenerator()

	// Without AI
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-no-ai"
	cfg.AIEnabled = false

	tmpDir := t.TempDir()
	err := g.Generate(cfg, tmpDir, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	initPath := filepath.Join(tmpDir, "init.yaml")
	content, _ := os.ReadFile(initPath)
	if strings.Contains(string(content), "ai/*.yaml") {
		t.Error("init.yaml should not contain 'ai/*.yaml' when AI is disabled")
	}

	// With AI
	cfg2 := NewDefaultWizardConfig()
	cfg2.ProjectName = "test-ai-includes"
	cfg2.AIEnabled = true

	tmpDir2 := t.TempDir()
	err = g.Generate(cfg2, tmpDir2, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	initPath2 := filepath.Join(tmpDir2, "init.yaml")
	content2, _ := os.ReadFile(initPath2)
	if !strings.Contains(string(content2), "ai/*.yaml") {
		t.Error("init.yaml should contain 'ai/*.yaml' when AI is enabled")
	}
}

// TestDockerComposeRepoEnvVar verifies the REPO env var in docker-compose.yaml.
func TestDockerComposeRepoEnvVar(t *testing.T) {
	// Test with GitHub repo creation
	t.Run("with github repo creation", func(t *testing.T) {
		g := NewGenerator()
		cfg := NewDefaultWizardConfig()
		cfg.ProjectName = "test-repo-gh"
		cfg.DeploymentMode = "docker"
		cfg.GithubCreateRepo = true
		cfg.GitRemoteURL = "git@github.com:myuser/hd-config.git"

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

		if !strings.Contains(yamlStr, "REPO=git@github.com:myuser/hd-config.git") {
			t.Errorf("docker-compose.yaml should contain REPO with git remote URL, got:\n%s", yamlStr)
		}
	})

	// Test without GitHub repo creation (local path)
	t.Run("without github repo creation", func(t *testing.T) {
		g := NewGenerator()
		cfg := NewDefaultWizardConfig()
		cfg.ProjectName = "test-repo-local"
		cfg.DeploymentMode = "docker"
		cfg.GithubCreateRepo = false

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

		if !strings.Contains(yamlStr, "REPO=/etc/honeydipper/config") {
			t.Errorf("docker-compose.yaml should contain REPO with local path, got:\n%s", yamlStr)
		}
	})
}

// TestDockerComposeVolumeMount verifies the config volume mount when not using GitHub repo.
func TestDockerComposeVolumeMount(t *testing.T) {
	// Test without GitHub repo creation - should have volume mount
	t.Run("without github repo creation has volume mount", func(t *testing.T) {
		g := NewGenerator()
		cfg := NewDefaultWizardConfig()
		cfg.ProjectName = "test-vol-local"
		cfg.DeploymentMode = "docker"
		cfg.GithubCreateRepo = false

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

		if !strings.Contains(yamlStr, "volumes:") {
			t.Errorf("docker-compose.yaml should contain volumes directive, got:\n%s", yamlStr)
		}
		if !strings.Contains(yamlStr, "/etc/honeydipper/config") {
			t.Errorf("docker-compose.yaml should mount config to /etc/honeydipper/config, got:\n%s", yamlStr)
		}
		// Verify the volume mount uses an absolute path (starts with /)
		lines := strings.Split(yamlStr, "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "- ") && strings.HasSuffix(trimmed, ":/etc/honeydipper/config") {
				hostPath := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "- "), ":/etc/honeydipper/config"))
				if !strings.HasPrefix(hostPath, "/") {
					t.Errorf("volume mount should use absolute path, got: %q", hostPath)
				}
			}
		}
	})

	// Test with GitHub repo creation - should NOT have volume mount
	t.Run("with github repo creation has no volume mount", func(t *testing.T) {
		g := NewGenerator()
		cfg := NewDefaultWizardConfig()
		cfg.ProjectName = "test-vol-gh"
		cfg.DeploymentMode = "docker"
		cfg.GithubCreateRepo = true
		cfg.GitRemoteURL = "git@github.com:user/repo.git"
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

		// The volumes: directive for config should NOT exist
		// (redis volume might still exist if RedisMode is local)
		// We check that there's no bind mount to /etc/honeydipper/config
		if strings.Contains(yamlStr, "/etc/honeydipper/config") {
			// It's OK if it's in the REPO env var line
			lines := strings.Split(yamlStr, "\n")
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "-") && strings.Contains(trimmed, "/etc/honeydipper/config") {
					t.Errorf("docker-compose.yaml should NOT have volume mount to /etc/honeydipper/config when using GitHub repo, line: %s", trimmed)
				}
			}
		}
	})
}

// TestSlackSecretPathsInIntegrations verifies user-provided Slack paths are used.
func TestSlackSecretPathsInIntegrations(t *testing.T) {
	// Test with user-provided paths
	t.Run("with custom slack paths", func(t *testing.T) {
		g := NewGenerator()
		cfg := NewDefaultWizardConfig()
		cfg.ProjectName = "test-slack-custom"
		cfg.SlackBotTokenPath = "secrets/slack/bot-token"
		cfg.SlackSigningSecretPath = "secrets/slack/signing-secret"
		cfg.SlackInteractionToken = "secrets/slack/interaction"
		cfg.SlackSlashCommandToken = "secrets/slack/slash"

		tmpDir := t.TempDir()
		err := g.Generate(cfg, tmpDir, false)
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		intPath := filepath.Join(tmpDir, "integrations.yaml")
		content, err := os.ReadFile(intPath)
		if err != nil {
			t.Fatalf("failed to read integrations.yaml: %v", err)
		}
		yamlStr := string(content)

		if !strings.Contains(yamlStr, "token: LOOKUP[vault,secrets/slack/bot-token]") {
			t.Errorf("integrations.yaml should contain custom bot token path, got:\n%s", yamlStr)
		}
		if !strings.Contains(yamlStr, "signatureSecret: LOOKUP[vault,secrets/slack/signing-secret]") {
			t.Errorf("integrations.yaml should contain custom signing secret path, got:\n%s", yamlStr)
		}
		if !strings.Contains(yamlStr, "interact_token: LOOKUP[vault,secrets/slack/interaction]") {
			t.Errorf("integrations.yaml should contain custom interaction token, got:\n%s", yamlStr)
		}
		if !strings.Contains(yamlStr, "slash_token: LOOKUP[vault,secrets/slack/slash]") {
			t.Errorf("integrations.yaml should contain custom slash command token, got:\n%s", yamlStr)
		}
	})

	// Test with default (empty) paths — should use LOOKUP
	t.Run("with default slack paths", func(t *testing.T) {
		g := NewGenerator()
		cfg := NewDefaultWizardConfig()
		cfg.ProjectName = "test-slack-default"
		// Slack paths left empty

		tmpDir := t.TempDir()
		err := g.Generate(cfg, tmpDir, false)
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		intPath := filepath.Join(tmpDir, "integrations.yaml")
		content, err := os.ReadFile(intPath)
		if err != nil {
			t.Fatalf("failed to read integrations.yaml: %v", err)
		}
		yamlStr := string(content)

		if !strings.Contains(yamlStr, "LOOKUP[vault,/secrets/data/test-slack-default/slack#botToken]") {
			t.Errorf("integrations.yaml should contain default LOOKUP for bot token, got:\n%s", yamlStr)
		}
		if !strings.Contains(yamlStr, "LOOKUP[vault,/secrets/data/test-slack-default/slack#signingSecret]") {
			t.Errorf("integrations.yaml should contain default LOOKUP for signing secret, got:\n%s", yamlStr)
		}
		if !strings.Contains(yamlStr, "$?nil") {
			t.Errorf("integrations.yaml should contain $?nil for default interact/slash tokens, got:\n%s", yamlStr)
		}
	})

	// Test with partial paths — some custom, some default
	t.Run("with partial slack paths", func(t *testing.T) {
		g := NewGenerator()
		cfg := NewDefaultWizardConfig()
		cfg.ProjectName = "test-slack-partial"
		cfg.SlackBotTokenPath = "custom/bot-token"
		// signing secret left empty
		cfg.SlackInteractionToken = "custom/interaction"
		// slash command token left empty

		tmpDir := t.TempDir()
		err := g.Generate(cfg, tmpDir, false)
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		intPath := filepath.Join(tmpDir, "integrations.yaml")
		content, err := os.ReadFile(intPath)
		if err != nil {
			t.Fatalf("failed to read integrations.yaml: %v", err)
		}
		yamlStr := string(content)

		if !strings.Contains(yamlStr, "token: LOOKUP[vault,custom/bot-token]") {
			t.Errorf("integrations.yaml should contain custom bot token path")
		}
		if !strings.Contains(yamlStr, "LOOKUP[vault,/secrets/data/test-slack-partial/slack#signingSecret]") {
			t.Errorf("integrations.yaml should contain default LOOKUP for signing secret")
		}
		if !strings.Contains(yamlStr, "interact_token: LOOKUP[vault,custom/interaction]") {
			t.Errorf("integrations.yaml should contain custom interaction token")
		}
		if !strings.Contains(yamlStr, "slash_token: $?nil") {
			t.Errorf("integrations.yaml should contain default $?nil for slash token")
		}
	})
}

// TestGithubTokenPathInIntegrations verifies user-provided GitHub paths are used.
func TestGithubTokenPathInIntegrations(t *testing.T) {
	t.Run("with custom github token path", func(t *testing.T) {
		g := NewGenerator()
		cfg := NewDefaultWizardConfig()
		cfg.ProjectName = "test-gh-custom"
		cfg.HasGithubPATIntegration = true
		cfg.GithubTokenPath = "secrets/github/my-pat"

		tmpDir := t.TempDir()
		err := g.Generate(cfg, tmpDir, false)
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		intPath := filepath.Join(tmpDir, "integrations.yaml")
		content, err := os.ReadFile(intPath)
		if err != nil {
			t.Fatalf("failed to read integrations.yaml: %v", err)
		}
		yamlStr := string(content)

		if !strings.Contains(yamlStr, "pat: LOOKUP[vault,secrets/github/my-pat]") {
			t.Errorf("integrations.yaml should contain custom GitHub token path, got:\n%s", yamlStr)
		}
	})

	t.Run("with custom github webhook secret path", func(t *testing.T) {
		g := NewGenerator()
		cfg := NewDefaultWizardConfig()
		cfg.ProjectName = "test-gh-webhook"
		cfg.HasGitHubAppIntegration = true
		cfg.GithubWebhookSecret = "secrets/github/webhook-secret"

		tmpDir := t.TempDir()
		err := g.Generate(cfg, tmpDir, false)
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		intPath := filepath.Join(tmpDir, "integrations.yaml")
		content, err := os.ReadFile(intPath)
		if err != nil {
			t.Fatalf("failed to read integrations.yaml: %v", err)
		}
		yamlStr := string(content)

		if !strings.Contains(yamlStr, "signatureSecret: LOOKUP[vault,secrets/github/webhook-secret]") {
			t.Errorf("integrations.yaml should contain custom webhook secret path, got:\n%s", yamlStr)
		}
	})
}

// ===== Dev mode secret template tests =====

func TestGenerateDevModeIntegrations(t *testing.T) {
	g := NewGenerator()
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-dev"
	cfg.SecretsBackend = "dev"
	cfg.HasGithubPATIntegration = true
	cfg.HasGitHubAppIntegration = true
	// In dev mode, values are stored with "$" prefix (stripped during template rendering)
	cfg.GithubTokenPath = "$HD_GITHUB_TOKEN"
	cfg.GithubWebhookSecret = "$HD_GITHUB_WEBHOOK_SECRET"
	cfg.SlackBotTokenPath = "$HD_SLACK_BOT_TOKEN"
	cfg.SlackSigningSecretPath = "$HD_SLACK_SIGNING_SECRET"
	cfg.SlackInteractionToken = "$HD_SLACK_INTERACTION_TOKEN"
	cfg.SlackSlashCommandToken = "$HD_SLACK_SLASH_TOKEN"

	tmpDir := t.TempDir()
	err := g.Generate(cfg, tmpDir, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	intPath := filepath.Join(tmpDir, "integrations.yaml")
	content, err := os.ReadFile(intPath)
	if err != nil {
		t.Fatalf("failed to read integrations.yaml: %v", err)
	}
	yamlStr := string(content)

	// Dev mode: should NOT contain LOOKUP[vault,...]
	if strings.Contains(yamlStr, "LOOKUP[vault") {
		t.Errorf("dev mode integrations.yaml should not contain LOOKUP[vault, got:\n%s", yamlStr)
	}

	// Should contain {% .env.* %} references for HD_ prefixed values
	if !strings.Contains(yamlStr, "{% .env.GITHUB_TOKEN %}") {
		t.Errorf("dev mode integrations.yaml should contain {%% .env.GITHUB_TOKEN %%}, got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "{% .env.GITHUB_WEBHOOK_SECRET %}") {
		t.Errorf("dev mode integrations.yaml should contain {%% .env.GITHUB_WEBHOOK_SECRET %%}, got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "{% .env.SLACK_BOT_TOKEN %}") {
		t.Errorf("dev mode integrations.yaml should contain {%% .env.SLACK_BOT_TOKEN %%}, got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "{% .env.SLACK_SIGNING_SECRET %}") {
		t.Errorf("dev mode integrations.yaml should contain {%% .env.SLACK_SIGNING_SECRET %%}, got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "{% .env.SLACK_INTERACTION_TOKEN %}") {
		t.Errorf("dev mode integrations.yaml should contain {%% .env.SLACK_INTERACTION_TOKEN %%}, got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "{% .env.SLACK_SLASH_TOKEN %}") {
		t.Errorf("dev mode integrations.yaml should contain {%% .env.SLACK_SLASH_TOKEN %%}, got:\n%s", yamlStr)
	}
}

func TestGenerateVaultModeIntegrations(t *testing.T) {
	g := NewGenerator()
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-vault"
	cfg.SecretsBackend = "vault"
	cfg.HasGithubPATIntegration = true
	cfg.HasGitHubAppIntegration = true
	cfg.GithubTokenPath = "secrets/github/pat"
	cfg.GithubWebhookSecret = "secrets/github/webhook"
	cfg.SlackBotTokenPath = "secrets/slack/bot-token"
	cfg.SlackSigningSecretPath = "secrets/slack/signing"

	tmpDir := t.TempDir()
	err := g.Generate(cfg, tmpDir, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	intPath := filepath.Join(tmpDir, "integrations.yaml")
	content, err := os.ReadFile(intPath)
	if err != nil {
		t.Fatalf("failed to read integrations.yaml: %v", err)
	}
	yamlStr := string(content)

	// Vault mode: should wrap in LOOKUP[vault,...]
	if !strings.Contains(yamlStr, "LOOKUP[vault,secrets/github/pat]") {
		t.Errorf("vault mode integrations.yaml should contain LOOKUP[vault,secrets/github/pat], got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "LOOKUP[vault,secrets/github/webhook]") {
		t.Errorf("vault mode integrations.yaml should contain LOOKUP[vault,secrets/github/webhook], got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "LOOKUP[vault,secrets/slack/bot-token]") {
		t.Errorf("vault mode integrations.yaml should contain LOOKUP[vault,secrets/slack/bot-token], got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "LOOKUP[vault,secrets/slack/signing]") {
		t.Errorf("vault mode integrations.yaml should contain LOOKUP[vault,secrets/slack/signing], got:\n%s", yamlStr)
	}
}

func TestGenerateDevModePlainValues(t *testing.T) {
	g := NewGenerator()
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-dev-plain"
	cfg.SecretsBackend = "dev"
	cfg.HasGithubPATIntegration = true
	cfg.GithubTokenPath = "my-plain-token-value"

	tmpDir := t.TempDir()
	err := g.Generate(cfg, tmpDir, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	intPath := filepath.Join(tmpDir, "integrations.yaml")
	content, err := os.ReadFile(intPath)
	if err != nil {
		t.Fatalf("failed to read integrations.yaml: %v", err)
	}
	yamlStr := string(content)

	// Dev mode with plain value: should use the value directly for the token
	if !strings.Contains(yamlStr, "pat: my-plain-token-value") {
		t.Errorf("dev mode should use plain value directly for token, got:\n%s", yamlStr)
	}
	// The plain value should NOT be wrapped in LOOKUP and NOT be an env ref
	if strings.Contains(yamlStr, "LOOKUP[vault,my-plain-token-value]") {
		t.Errorf("dev mode should not wrap plain values in LOOKUP, got:\n%s", yamlStr)
	}
	if strings.Contains(yamlStr, "{% .env") {
		t.Errorf("dev mode should not render plain values as env refs, got:\n%s", yamlStr)
	}
}

func TestGenerateDevModeDefaultPaths(t *testing.T) {
	// When dev mode has empty secret paths, the template should fall back
	// to default LOOKUP paths (same as vault mode defaults)
	g := NewGenerator()
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-dev-default"
	cfg.SecretsBackend = "dev"
	cfg.HasGithubPATIntegration = true
	// Leave GithubTokenPath empty

	tmpDir := t.TempDir()
	err := g.Generate(cfg, tmpDir, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	intPath := filepath.Join(tmpDir, "integrations.yaml")
	content, err := os.ReadFile(intPath)
	if err != nil {
		t.Fatalf("failed to read integrations.yaml: %v", err)
	}
	yamlStr := string(content)

	// When path is empty, the template falls back to default LOOKUP path
	// This is the existing behavior for empty paths
	if !strings.Contains(yamlStr, "LOOKUP[vault,/secrets/data/test-dev-default/github#pat]") {
		t.Errorf("dev mode with empty path should fall back to default LOOKUP, got:\n%s", yamlStr)
	}
}

// ===== Dev mode docker-compose env var tests =====

func TestDockerComposeDevModeEnvVars(t *testing.T) {
	g := NewGenerator()
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-dev-compose"
	cfg.DeploymentMode = "docker"
	cfg.SecretsBackend = "dev"
	cfg.HasGithubPATIntegration = true
	cfg.HasGitHubAppIntegration = true
	// Dev mode values stored with "$" prefix
	cfg.GithubTokenPath = "$HD_GITHUB_TOKEN"
	cfg.GithubWebhookSecret = "$HD_GITHUB_WEBHOOK_SECRET"
	cfg.SlackBotTokenPath = "$HD_SLACK_BOT_TOKEN"
	cfg.SlackSigningSecretPath = "$HD_SLACK_SIGNING_SECRET"

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

	// Should contain HD_* env var passthroughs
	if !strings.Contains(yamlStr, "HD_GITHUB_TOKEN=${HD_GITHUB_TOKEN}") {
		t.Errorf("docker-compose.yaml should contain HD_GITHUB_TOKEN passthrough, got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "HD_GITHUB_WEBHOOK_SECRET=${HD_GITHUB_WEBHOOK_SECRET}") {
		t.Errorf("docker-compose.yaml should contain HD_GITHUB_WEBHOOK_SECRET passthrough, got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "HD_SLACK_BOT_TOKEN=${HD_SLACK_BOT_TOKEN}") {
		t.Errorf("docker-compose.yaml should contain HD_SLACK_BOT_TOKEN passthrough, got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "HD_SLACK_SIGNING_SECRET=${HD_SLACK_SIGNING_SECRET}") {
		t.Errorf("docker-compose.yaml should contain HD_SLACK_SIGNING_SECRET passthrough, got:\n%s", yamlStr)
	}
}

func TestDockerComposeDevModePartialEnvVars(t *testing.T) {
	g := NewGenerator()
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-dev-partial"
	cfg.DeploymentMode = "docker"
	cfg.SecretsBackend = "dev"
	cfg.HasGithubPATIntegration = true
	cfg.GithubTokenPath = "$HD_GITHUB_TOKEN"
	// Slack fields left empty — should not appear

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

	// Should contain the one HD_ var
	if !strings.Contains(yamlStr, "HD_GITHUB_TOKEN=${HD_GITHUB_TOKEN}") {
		t.Errorf("docker-compose.yaml should contain HD_GITHUB_TOKEN, got:\n%s", yamlStr)
	}
	// Should NOT contain Slack vars (they were empty)
	if strings.Contains(yamlStr, "HD_SLACK") {
		t.Errorf("docker-compose.yaml should not contain HD_SLACK vars when empty, got:\n%s", yamlStr)
	}
}

func TestDockerComposeVaultModeNoDevEnvVars(t *testing.T) {
	g := NewGenerator()
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-vault-compose"
	cfg.DeploymentMode = "docker"
	cfg.SecretsBackend = "vault"
	cfg.HasGithubPATIntegration = true
	cfg.GithubTokenPath = "secrets/github/pat"

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

	// Should NOT contain any HD_* env var passthroughs in vault mode
	if strings.Contains(yamlStr, "HD_") {
		t.Errorf("docker-compose.yaml should not contain HD_ vars in vault mode, got:\n%s", yamlStr)
	}
}

func TestDockerComposeDevModeNoEnvVars(t *testing.T) {
	// Dev mode but no HD_* vars referenced — should still generate without errors
	g := NewGenerator()
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-dev-empty"
	cfg.DeploymentMode = "docker"
	cfg.SecretsBackend = "dev"
	cfg.HasGithubPATIntegration = true
	// Leave all secret fields empty

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

	// Should not contain any HD_ env var entries
	if strings.Contains(yamlStr, "HD_") {
		t.Errorf("docker-compose.yaml should not contain HD_ vars when none referenced, got:\n%s", yamlStr)
	}
	// Should still have the REPO env var
	if !strings.Contains(yamlStr, "REPO=") {
		t.Errorf("docker-compose.yaml should still have REPO env var, got:\n%s", yamlStr)
	}
}

func TestCollectDevEnvVars(t *testing.T) {
	t.Run("dev mode with HD_ vars", func(t *testing.T) {
		cfg := NewDefaultWizardConfig()
		cfg.SecretsBackend = "dev"
		cfg.GithubTokenPath = "$HD_GITHUB_TOKEN"
		cfg.GithubWebhookSecret = "$HD_GITHUB_WEBHOOK"
		cfg.SlackBotTokenPath = "$HD_SLACK_BOT"

		vars := cfg.CollectDevEnvVars()
		if len(vars) != 3 {
			t.Fatalf("expected 3 dev env vars, got %d: %v", len(vars), vars)
		}
		expected := map[string]bool{
			"HD_GITHUB_TOKEN":     true,
			"HD_GITHUB_WEBHOOK":  true,
			"HD_SLACK_BOT":      true,
		}
		for _, v := range vars {
			if !expected[v] {
				t.Errorf("unexpected var %q", v)
			}
		}
	})

	t.Run("vault mode returns nil", func(t *testing.T) {
		cfg := NewDefaultWizardConfig()
		cfg.SecretsBackend = "vault"
		cfg.GithubTokenPath = "secrets/github/pat"

		vars := cfg.CollectDevEnvVars()
		if vars != nil {
			t.Errorf("expected nil for vault mode, got %v", vars)
		}
	})

	t.Run("deduplication", func(t *testing.T) {
		cfg := NewDefaultWizardConfig()
		cfg.SecretsBackend = "dev"
		// Same var referenced in two fields
		cfg.GithubTokenPath = "$HD_GITHUB_TOKEN"
		cfg.GithubKeyPath = "$HD_GITHUB_TOKEN"

		vars := cfg.CollectDevEnvVars()
		if len(vars) != 1 {
			t.Fatalf("expected 1 deduplicated var, got %d: %v", len(vars), vars)
		}
		if vars[0] != "HD_GITHUB_TOKEN" {
			t.Errorf("expected HD_GITHUB_TOKEN, got %q", vars[0])
		}
	})

	t.Run("plain values ignored", func(t *testing.T) {
		cfg := NewDefaultWizardConfig()
		cfg.SecretsBackend = "dev"
		cfg.GithubTokenPath = "my-plain-token"
		cfg.GithubKeyPath = "$HD_GITHUB_KEY"

		vars := cfg.CollectDevEnvVars()
		if len(vars) != 1 {
			t.Fatalf("expected 1 var, got %d: %v", len(vars), vars)
		}
		if vars[0] != "HD_GITHUB_KEY" {
			t.Errorf("expected HD_GITHUB_KEY, got %q", vars[0])
		}
	})

	t.Run("nil config", func(t *testing.T) {
		var cfg *WizardConfig
		vars := cfg.CollectDevEnvVars()
		if vars != nil {
			t.Errorf("expected nil for nil config, got %v", vars)
		}
	})
}

// ===== Git init / git remote tests =====

func TestGitInitWhenCreatingRepo(t *testing.T) {
	// Use a temp dir as the config dir so git init runs in it
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "my-config")

	g := NewGenerator()
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-git-init"
	cfg.ConfigDir = configDir
	cfg.DeploymentMode = "docker"
	cfg.GithubCreateRepo = true
	cfg.GitRemoteURL = "git@github.com:myuser/hd-config.git"

	// Generate should run git init and git remote add
	err := g.Generate(cfg, configDir, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check that git init was run
	if !cfg.GitInit {
		t.Error("GitInit should be true after generating with GithubCreateRepo")
	}

	// Check that a .git directory was created
	gitDir := filepath.Join(configDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Errorf("git init should create .git directory at %s", gitDir)
	}

	// Check that the remote was added
	remoteOutput, err := exec.Command("git", "-C", configDir, "remote", "get-url", "origin").Output()
	if err != nil {
		t.Fatalf("git remote get-url origin failed: %v", err)
	}
	remoteURL := strings.TrimSpace(string(remoteOutput))
	if remoteURL != "git@github.com:myuser/hd-config.git" {
		t.Errorf("expected remote URL 'git@github.com:myuser/hd-config.git', got %q", remoteURL)
	}
}

func TestGitNotInitWhenNotCreatingRepo(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "my-config")

	g := NewGenerator()
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-no-git"
	cfg.ConfigDir = configDir
	cfg.DeploymentMode = "docker"
	cfg.GithubCreateRepo = false

	err := g.Generate(cfg, configDir, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check that git init was NOT run
	if cfg.GitInit {
		t.Error("GitInit should be false when GithubCreateRepo is false")
	}

	// Check that no .git directory was created
	gitDir := filepath.Join(configDir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		t.Error("git init should NOT be called when GithubCreateRepo is false")
	}
}

func TestGitInitWithoutRemoteURL(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "my-config")

	g := NewGenerator()
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-git-no-remote"
	cfg.ConfigDir = configDir
	cfg.DeploymentMode = "docker"
	cfg.GithubCreateRepo = true
	// No GitRemoteURL set

	err := g.Generate(cfg, configDir, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check that git init was run
	if !cfg.GitInit {
		t.Error("GitInit should be true after generating with GithubCreateRepo")
	}

	// Check that .git exists
	gitDir := filepath.Join(configDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Error("git init should create .git directory")
	}

	// Check that no remote was added
	_, err = exec.Command("git", "-C", configDir, "remote", "get-url", "origin").Output()
	if err == nil {
		t.Error("git remote should NOT be set when GitRemoteURL is empty")
	}
}

func TestGitRemoteAddIdempotent(t *testing.T) {
	// First run: generate with git remote add
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "my-config")

	g := NewGenerator()
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-idempotent-1"
	cfg.ConfigDir = configDir
	cfg.DeploymentMode = "docker"
	cfg.GithubCreateRepo = true
	cfg.GitRemoteURL = "git@github.com:myuser/hd-config.git"

	err := g.Generate(cfg, configDir, false)
	if err != nil {
		t.Fatalf("first Generate failed: %v", err)
	}

	// Verify remote was added
	remoteOutput, err := exec.Command("git", "-C", configDir, "remote", "get-url", "origin").Output()
	if err != nil {
		t.Fatalf("git remote get-url origin failed after first generate: %v", err)
	}
	firstURL := strings.TrimSpace(string(remoteOutput))
	if firstURL != "git@github.com:myuser/hd-config.git" {
		t.Errorf("expected remote URL 'git@github.com:myuser/hd-config.git', got %q", firstURL)
	}

	// Second run: generate again with a different remote URL
	cfg2 := NewDefaultWizardConfig()
	cfg2.ProjectName = "test-idempotent-2"
	cfg2.ConfigDir = configDir
	cfg2.DeploymentMode = "docker"
	cfg2.GithubCreateRepo = true
	cfg2.GitRemoteURL = "git@github.com:myuser/hd-config-updated.git"

	err = g.Generate(cfg2, configDir, false)
	if err != nil {
		t.Fatalf("second Generate failed (idempotent test): %v", err)
	}

	// Verify remote was updated
	remoteOutput, err = exec.Command("git", "-C", configDir, "remote", "get-url", "origin").Output()
	if err != nil {
		t.Fatalf("git remote get-url origin failed after second generate: %v", err)
	}
	secondURL := strings.TrimSpace(string(remoteOutput))
	if secondURL != "git@github.com:myuser/hd-config-updated.git" {
		t.Errorf("expected updated remote URL 'git@github.com:myuser/hd-config-updated.git', got %q", secondURL)
	}
}

func TestGitRemoteAddWhenRemoteDoesNotExist(t *testing.T) {
	// Ensure that when no origin remote exists, git remote add works
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "my-config")

	// First, create a git repo with no remote
	err := exec.Command("git", "init", configDir).Run()
	if err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	// Verify no remote exists
	_, err = exec.Command("git", "-C", configDir, "remote", "get-url", "origin").Output()
	if err == nil {
		t.Fatal("expected no origin remote to exist")
	}

	// Now generate with GithubCreateRepo
	g := NewGenerator()
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-no-remote-exists"
	cfg.ConfigDir = configDir
	cfg.DeploymentMode = "docker"
	cfg.GithubCreateRepo = true
	cfg.GitRemoteURL = "git@github.com:myuser/new-repo.git"

	err = g.Generate(cfg, configDir, false)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify remote was added
	remoteOutput, err := exec.Command("git", "-C", configDir, "remote", "get-url", "origin").Output()
	if err != nil {
		t.Fatalf("git remote get-url origin failed: %v", err)
	}
	remoteURL := strings.TrimSpace(string(remoteOutput))
	if remoteURL != "git@github.com:myuser/new-repo.git" {
		t.Errorf("expected remote URL 'git@github.com:myuser/new-repo.git', got %q", remoteURL)
	}
}

func TestDockerComposeRemoteURL(t *testing.T) {
	// Test that GitRemoteURL is used as REPO env var in docker-compose
	g := NewGenerator()
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-remote-url"
	cfg.DeploymentMode = "docker"
	cfg.GithubCreateRepo = true
	cfg.GitRemoteURL = "git@github.com:myuser/hd-config.git"

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

	// Should use the git remote URL as REPO
	if !strings.Contains(yamlStr, "REPO=git@github.com:myuser/hd-config.git") {
		t.Errorf("docker-compose.yaml should contain REPO with git remote URL, got:\n%s", yamlStr)
	}
}

func TestDockerComposeFallbackREPO(t *testing.T) {
	// Test that when GitRemoteURL is not set, REPO falls back to local path
	g := NewGenerator()
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-fallback-repo"
	cfg.DeploymentMode = "docker"
	cfg.GithubCreateRepo = true
	// No GitRemoteURL set

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

	// Should fall back to local path since no GitRemoteURL is set
	if !strings.Contains(yamlStr, "REPO=/etc/honeydipper/config") {
		t.Errorf("docker-compose.yaml should contain REPO with local path, got:\n%s", yamlStr)
	}
}

func TestDockerComposeLocalREPO(t *testing.T) {
	// Test that when GithubCreateRepo is false, REPO uses local path
	g := NewGenerator()
	cfg := NewDefaultWizardConfig()
	cfg.ProjectName = "test-local-repo"
	cfg.DeploymentMode = "docker"
	cfg.GithubCreateRepo = false

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

	// Should use local path as REPO
	if !strings.Contains(yamlStr, "REPO=/etc/honeydipper/config") {
		t.Errorf("docker-compose.yaml should contain REPO with local path, got:\n%s", yamlStr)
	}
}
