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
		cfg.GithubRepoName = "myuser/hd-config"

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

		if !strings.Contains(yamlStr, "REPO=https://github.com/myuser/hd-config.git") {
			t.Errorf("docker-compose.yaml should contain REPO with GitHub URL, got:\n%s", yamlStr)
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

		if !strings.Contains(yamlStr, "token: \"secrets/slack/bot-token\"") {
			t.Errorf("integrations.yaml should contain custom bot token path, got:\n%s", yamlStr)
		}
		if !strings.Contains(yamlStr, "signatureSecret: \"secrets/slack/signing-secret\"") {
			t.Errorf("integrations.yaml should contain custom signing secret path, got:\n%s", yamlStr)
		}
		if !strings.Contains(yamlStr, "interact_token: \"secrets/slack/interaction\"") {
			t.Errorf("integrations.yaml should contain custom interaction token, got:\n%s", yamlStr)
		}
		if !strings.Contains(yamlStr, "slash_token: \"secrets/slack/slash\"") {
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

		if !strings.Contains(yamlStr, "token: \"custom/bot-token\"") {
			t.Errorf("integrations.yaml should contain custom bot token path")
		}
		if !strings.Contains(yamlStr, "LOOKUP[vault,/secrets/data/test-slack-partial/slack#signingSecret]") {
			t.Errorf("integrations.yaml should contain default LOOKUP for signing secret")
		}
		if !strings.Contains(yamlStr, "interact_token: \"custom/interaction\"") {
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

		if !strings.Contains(yamlStr, "pat: \"secrets/github/my-pat\"") {
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

		if !strings.Contains(yamlStr, "signatureSecret: \"secrets/github/webhook-secret\"") {
			t.Errorf("integrations.yaml should contain custom webhook secret path, got:\n%s", yamlStr)
		}
	})
}
