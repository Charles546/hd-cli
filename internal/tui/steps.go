// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If a separate written commercial agreement exists, you may use this file under those terms instead.

package tui

import (
	"fmt"
	"strings"

)

// UpdateStep processes a single step's input and updates the config.
// Called by the CLI layer externally; for embedded use call RunWizard instead.
func UpdateStep(m *WizardModel, step int, key string, value string) error {
	if m == nil || m.config == nil {
		return fmt.Errorf("wizard not initialized")
	}

	cfg := m.config

	// Normalize key
	key = strings.TrimSpace(strings.ToLower(key))

	switch step {
	case 2:
		if key == "project_name" {
			cfg.ProjectName = value
		}
	case 3:
		if key == "config_dir" {
			cfg.ConfigDir = value
		}
	case 4:
		if key == "deployment_mode" {
			cfg.DeploymentMode = value
		}
	case 5:
		switch key {
		case "essentials_repo_url":
			cfg.EssentialsRepoURL = value
		case "essentials_branch":
			cfg.EssentialsBranch = value
		}
	case 6:
		switch key {
		case "github_integration_type":
			cfg.GithubIntegrationType = value
		case "github_app_id":
			cfg.GithubAppID = value
		case "github_installation_id":
			cfg.GithubInstallationID = value
		case "github_key_path":
			cfg.GithubKeyPath = value
		case "github_token_path":
			cfg.GithubTokenPath = value
		case "github_webhook_secret_path":
			cfg.GithubWebhookSecret = value
		}
	case 7:
		switch key {
		case "slack_bot_token_path":
			cfg.SlackBotTokenPath = value
		case "slack_signing_secret_path":
			cfg.SlackSigningSecretPath = value
		case "slack_interaction_token_path":
			cfg.SlackInteractionTokenPath = value
		case "slack_slash_command_token_path":
			cfg.SlackSlashCommandTokenPath = value
		}
	case 8:
		switch key {
		case "ai_enabled":
			cfg.AIEnabled = value == "true" || value == "yes"
		case "ai_api_key_path":
			cfg.AIAPIKeyPath = value
		case "ai_base_url":
			cfg.AIBaseURL = value
		case "ai_model":
			cfg.AIModel = value
		case "ai_engine_name":
			cfg.AIEngineName = value
		}
	case 9:
		switch key {
		case "secrets_backend":
			cfg.SecretsBackend = value
		case "vault_address":
			cfg.VaultAddress = value
		case "vault_auth_method":
			cfg.VaultAuthMethod = value
		}
	case 10:
		switch key {
		case "redis_mode":
			cfg.RedisMode = value
		case "redis_connection_string":
			cfg.RedisConnString = value
		}
	case 11:
		switch key {
		case "docker_image_tag":
			cfg.DockerImageTag = value
		case "docker_enable_ui":
			cfg.DockerEnableUI = value == "true" || value == "yes"
		case "docker_api_port":
			_, _ = fmt.Sscanf(value, "%d", &cfg.DockerAPIPort)
		case "docker_webhook_port":
			_, _ = fmt.Sscanf(value, "%d", &cfg.DockerWebhookPort)
		}
	case 12:
		switch key {
		case "k8s_namespace":
			cfg.K8sNamespace = value
		case "k8s_repo_strategy":
			cfg.K8sRepoStrategy = value
		}
	case 13:
		switch key {
		case "source_clone_path":
			cfg.SourceClonePath = value
		case "source_branch":
			cfg.SourceBranch = value
		}
	case 14:
		switch key {
		case "github_create_repo":
			cfg.GithubCreateRepo = value == "true" || value == "yes"
		case "github_repo_name":
			cfg.GithubRepoName = value
		case "github_repo_visibility":
			cfg.GithubRepoVis = value
		}
	}

	return nil
}

// StepLabels returns a human-readable label for each wizard step.
func StepLabels() []string {
	return []string{
		"",
		"Welcome",
		"Project Name",
		"Config Directory",
		"Deployment Mode",
		"Config Repo",
		"GitHub Integration",
		"Slack Integration",
		"AI Agent",
		"Secrets Backend",
		"Redis",
		"Docker Config",
		"Kubernetes Config",
		"Source Config",
		"GitHub Repo",
		"Summary",
	}
}

// IsStepRequired returns true if the step's fields must be filled for a valid config.
func IsStepRequired(step int) bool {
	switch step {
	case 2, 3, 4, 5, 6, 9, 10:
		return true
	case 11:
		return true
	case 12, 13:
		return true
	default:
		return false
	}
}

// StepHelp returns contextual help text for a step.
func StepHelp(step int) string {
	helps := map[int]string{
		1:  "Welcome to Honeydipper! This wizard creates your initial config.",
		2:  "A short name for your project. Used as the directory name.",
		3:  "The path where config files will be written. Use absolute or relative path.",
		4:  "Docker: easiest setup. Source: build from source. K8s: use Kubernetes.",
		5:  "The essentials config repo provides baseline workflows and definitions.",
		6:  "GitHub integration enables push event triggers and workflow dispatch.",
		7:  "Slack integration enables bot notifications and slash commands.",
		8:  "AI agent provides natural language workflow triggers and completions.",
		9:  "Vault is recommended for production. Dev mode stores secrets in env vars.",
		10: "Local Redis runs in the same container. External uses a separate Redis server.",
		11: "Docker deployment settings. Only applies if Docker mode is selected.",
		12: "Kubernetes deployment settings. Only applies if K8s mode is selected.",
		13: "Source build settings. Only applies if source mode is selected.",
		14: "Optionally create a new GitHub repository for the generated config.",
		15: "Review your choices before generating the configuration files.",
	}
	if h, ok := helps[step]; ok {
		return h
	}
	return ""
}

// ValidateStepComplete checks if a step has all required data.
func ValidateStepComplete(m *WizardModel) error {
	if m == nil || m.config == nil {
		return fmt.Errorf("wizard not initialized")
	}

	cfg := m.config

	switch m.step {
	case 1:
		return nil
	case 2:
		if strings.TrimSpace(cfg.ProjectName) == "" {
			return fmt.Errorf("project name is required")
		}
	case 3:
		if strings.TrimSpace(cfg.ConfigDir) == "" {
			return fmt.Errorf("config directory is required")
		}
	case 4:
		if cfg.DeploymentMode == "" {
			return fmt.Errorf("deployment mode is required")
		}
		if cfg.DeploymentMode != "docker" && cfg.DeploymentMode != "source" && cfg.DeploymentMode != "kubernetes" {
			return fmt.Errorf("invalid deployment mode: %s", cfg.DeploymentMode)
		}
	case 5:
		if strings.TrimSpace(cfg.EssentialsRepoURL) == "" {
			return fmt.Errorf("essentials repo URL is required")
		}
	case 6:
		switch cfg.GithubIntegrationType {
	case "github_app":
		if strings.TrimSpace(cfg.GithubAppID) == "" {
			return fmt.Errorf("github App ID is required")
		}
		if strings.TrimSpace(cfg.GithubInstallationID) == "" {
			return fmt.Errorf("github Installation ID is required")
		}
	case "pat":
			if strings.TrimSpace(cfg.GithubTokenPath) == "" {
				return fmt.Errorf("token secret path is required")
			}
		}
	case 11:
		if cfg.DeploymentMode == "docker" {
			if strings.TrimSpace(cfg.DockerImageTag) == "" {
				return fmt.Errorf("docker image tag is required")
			}
		}
	case 12:
		if cfg.DeploymentMode == "kubernetes" {
			if strings.TrimSpace(cfg.K8sNamespace) == "" {
				return fmt.Errorf("kubernetes namespace is required")
			}
		}
	case 13:
		if cfg.DeploymentMode == "source" {
			if strings.TrimSpace(cfg.SourceClonePath) == "" {
				return fmt.Errorf("clone path is required")
			}
		}
	}

	return nil
}

// renderWelcome shows the welcome/intro screen with an overview of the wizard.
func renderWelcome(m *WizardModel) string {
	var b strings.Builder

	b.WriteString(WelcomeStyle.Render(
		"Welcome to the Honeydipper Configuration Wizard!\n\n" +
			"This wizard will guide you through setting up a new\n" +
			"Honeydipper configuration directory with sensible defaults.\n\n" +
			"You will be asked about:\n" +
			"  • Project settings and deployment mode\n" +
			"  • GitHub and Slack integrations\n" +
			"  • AI agent configuration\n" +
			"  • Secrets backend and Redis\n" +
			"  • Platform-specific options\n\n" +
			"At the end, all configuration files will be generated.\n",
	))
	b.WriteString("\n")
	b.WriteString(ConfirmStyle.Render("  Press Enter to begin, or q to quit."))
	b.WriteString("\n")

	return b.String()
}
