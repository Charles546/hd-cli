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
		case "essentials_clone_type":
			cfg.EssentialsCloneType = value
			// Also update the docker-compose field for consistency
			cfg.BootstrapCloneCredentialType = value
		case "essentials_clone_pat":
			cfg.EssentialsClonePAT = value
		case "essentials_clone_key":
			cfg.EssentialsCloneKey = value
		case "essentials_clone_key_pass_env":
			cfg.EssentialsCloneKeyPassEnv = value
		}
	case 6:
		switch key {
		case "secrets_backend":
			cfg.SecretsBackend = value
		case "vault_address":
			cfg.VaultAddress = value
		case "vault_auth_method":
			cfg.VaultAuthMethod = value
		}
	case 7:
		switch key {
		case "redis_mode":
			cfg.RedisMode = value
		case "redis_connection_string":
			cfg.RedisConnString = value
		}
	case 8:
		switch key {
		case "has_github_pat_integration":
			cfg.HasGithubPATIntegration = value == "true" || value == "yes"
		case "has_github_app_integration":
			cfg.HasGitHubAppIntegration = value == "true" || value == "yes"
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
	case 9:
		switch key {
		case "slack_bot_token_path":
			cfg.SlackBotTokenPath = value
		case "slack_signing_secret_path":
			cfg.SlackSigningSecretPath = value
		case "slack_interaction_token":
			cfg.SlackInteractionToken = value
		case "slack_slash_command_token":
			cfg.SlackSlashCommandToken = value
		}
	case 10:
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
		case "git_remote_url":
			cfg.GitRemoteURL = value
		case "use_local_copy":
			cfg.UseLocalCopy = value == "true" || value == "yes"
		case "config_repo_clone_auth":
			cfg.ConfigRepoCloneAuth = value
		case "config_repo_pat_env_var":
			cfg.ConfigRepoPATValue = value
		case "config_repo_gh_app_id":
			cfg.ConfigRepoGHAppID = value
		case "config_repo_gh_installation_id":
			cfg.ConfigRepoGHInstallID = value
		case "config_repo_gh_app_key":
			cfg.ConfigRepoGHAppKey = value
		case "config_repo_ssh_key":
			cfg.ConfigRepoSSHKey = value
		case "config_repo_ssh_file":
			cfg.ConfigRepoSSHFile = value
		case "config_repo_ssh_key_pass_env":
			cfg.ConfigRepoSSHKeyPassEnv = value
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
		"Secrets Backend",
		"Redis",
		"GitHub Integration",
		"Slack Integration",
		"AI Agent",
		"Docker Config",
		"Kubernetes Config",
		"Source Config",
		"GitHub Repo",
		"Clone Auth & Secure Exec",
		"Summary",
	}
}

// IsStepRequired returns true if the step's fields must be filled for a valid config.
func IsStepRequired(step int) bool {
	switch step {
	case 2, 3, 4, 5, 6, 7, 10:
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
		6:  "Vault is recommended for production. Dev mode stores secrets in env vars.",
		7:  "Local Redis runs in the same container. External uses a separate Redis server.",
		8:  "GitHub integration enables push event triggers and workflow dispatch.",
		9:  "Slack integration enables bot notifications and slash commands.",
		10: "AI agent provides natural language workflow triggers and completions.",
		11: "Docker deployment settings. Only applies if Docker mode is selected.",
		12: "Kubernetes deployment settings. Only applies if K8s mode is selected.",
		13: "Source build settings. Only applies if source mode is selected.",
		14: "Optionally create a new GitHub repository for the generated config.",
		15: "Configure clone authentication and secure execution settings.",
		16: "Review your choices before generating the configuration files.",
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
		if cfg.SecretsBackend == "" {
			return fmt.Errorf("secrets backend selection is required")
		}
	case 8:
		if cfg.HasGitHubAppIntegration {
			if strings.TrimSpace(cfg.GithubAppID) == "" {
				return fmt.Errorf("github App ID is required when GitHub App integration is enabled")
			}
			if strings.TrimSpace(cfg.GithubInstallationID) == "" {
				return fmt.Errorf("github Installation ID is required when GitHub App integration is enabled")
			}
			if strings.TrimSpace(cfg.GithubKeyPath) == "" {
				return fmt.Errorf("private key secret path is required when GitHub App integration is enabled")
			}
		}
		if cfg.HasGithubPATIntegration {
			if strings.TrimSpace(cfg.GithubTokenPath) == "" {
				return fmt.Errorf("token secret path is required when PAT integration is enabled")
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
	case 14:
		if cfg.GithubCreateRepo {
			if strings.TrimSpace(cfg.GitRemoteURL) == "" {
				return fmt.Errorf("git remote URL is required when creating a GitHub repo")
			}
			if !cfg.UseLocalCopy {
				auth := strings.TrimSpace(cfg.ConfigRepoCloneAuth)
				switch auth {
				case "none":
				case "pat":
					if strings.TrimSpace(cfg.ConfigRepoPATValue) == "" {
						return fmt.Errorf("PAT is required when clone auth is 'pat'")
					}
				case "github_app":
					if strings.TrimSpace(cfg.ConfigRepoGHAppID) == "" {
						return fmt.Errorf("GitHub App ID is required when clone auth is 'github_app'")
					}
					if strings.TrimSpace(cfg.ConfigRepoGHInstallID) == "" {
						return fmt.Errorf("installation ID is required when clone auth is 'github_app'")
					}
					if strings.TrimSpace(cfg.ConfigRepoGHAppKey) == "" {
						return fmt.Errorf("private key is required when clone auth is 'github_app'")
					}
				case "ssh":
					if !strings.HasPrefix(cfg.GitRemoteURL, "git@") {
						return fmt.Errorf("SSH remote URL must start with 'git@'")
					}
					if strings.TrimSpace(cfg.ConfigRepoSSHKey) == "" && strings.TrimSpace(cfg.ConfigRepoSSHFile) == "" {
						return fmt.Errorf("either SSH key content or SSH key file path is required when clone auth is 'ssh'")
					}
				default:
					return fmt.Errorf("invalid clone auth method %q (must be one of: none, pat, github_app, ssh)", auth)
				}
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
			"  • Secrets backend and Redis\n" +
			"  • GitHub and Slack integrations\n" +
			"  • AI agent configuration\n" +
			"  • Repo clone credentials\n" +
			"  • Platform-specific options\n\n" +
			"At the end, all configuration files will be generated.\n",
	))
	b.WriteString("\n")
	b.WriteString(ConfirmStyle.Render("  Press Enter to begin, or q to quit."))
	b.WriteString("\n")

	return b.String()
}
