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
}

func TestGenerateNilConfig(t *testing.T) {
	g := NewGenerator()
	err := g.Generate(nil, t.TempDir(), false)
	if err == nil {
		t.Error("expected error for nil config")
	}
}
