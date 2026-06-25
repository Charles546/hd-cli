// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If a separate written commercial agreement exists, you may use this file under those terms instead.

package tui

import (
	"fmt"
	"strings"

	"github.com/Charles546/hd-cli/internal/config"
)

// stepType categorizes how a step is interacted with.
type stepType int

const (
	stepTypeNavigate   stepType = iota // No inputs, just press enter (welcome)
	stepTypeSingleField               // One text input field
	stepTypeMultiField                // Multiple text input fields
	stepTypeRadio                     // Radio selection (possibly with conditional fields)
	stepTypeCheckbox                  // Multiple checkbox toggles (yes/no for each option)
)

// fieldDescriptor describes a single text input field within a step.
type fieldDescriptor struct {
	label       string
	placeholder string
	help        string
	getValue    func(*config.WizardConfig) string
	setValue    func(*config.WizardConfig, string)
	// condition, if set, determines whether this field is shown.
	// When nil, the field is always shown.
	condition func(*config.WizardConfig) bool
	// placeholderFunc, if set, returns dynamic placeholder text based on config.
	// It takes precedence over the static placeholder field.
	placeholderFunc func(*config.WizardConfig) string
}

// checkboxOption describes a single checkbox toggle within a step.
type checkboxOption struct {
	label    string
	help     string
	getValue func(*config.WizardConfig) bool
	setValue func(*config.WizardConfig, bool)
	// condition, if set, determines whether this checkbox is shown.
	// When nil, the checkbox is always shown.
	condition func(*config.WizardConfig) bool
}

// stepInfo describes the structure and behavior of a wizard step.
type stepInfo struct {
	title          string
	stepType       stepType
	fields         []fieldDescriptor
	checkboxes     []checkboxOption
	checkboxLabel  string
	radioLabel     string
	radioOptions   []string
	radioGetter    func(*config.WizardConfig) string
	radioSetter    func(*config.WizardConfig, string)
}

// isDevMode returns true if the wizard is configured for dev secrets backend.
func isDevMode(cfg *config.WizardConfig) bool {
	return cfg != nil && cfg.SecretsBackend == "dev"
}

// ghSecretLabels returns the label, placeholder, and help text for a GitHub
// secret field depending on the secrets backend.
func ghSecretLabels(cfg *config.WizardConfig, field string) (label, placeholder, help string) {
	if isDevMode(cfg) {
		switch field {
		case "key":
			return "Private key value", "Plain text or $HD_ENV_VAR", "GitHub App private key value or env var reference (only for github_app integration)"
		case "token":
			return "Token value", "Plain text or $HD_ENV_VAR", "Personal Access Token value or env var reference (only for PAT integration)"
		case "webhook":
			return "Webhook secret value", "Plain text or $HD_ENV_VAR", "Webhook signing secret value or env var reference (recommended for github_app integration)"
		}
	}
	switch field {
	case "key":
		return "Private key secret path", "Path to secret containing the private key", "Path to the secret containing the GitHub App private key (only for github_app integration)"
	case "token":
		return "Token secret path", "Path to secret containing the PAT", "Path to the secret containing the Personal Access Token (only for PAT integration)"
	case "webhook":
		return "Webhook secret path", "Path to webhook signing secret", "Path to the webhook signing secret (recommended for github_app integration)"
	}
	return "", "", ""
}

// slackSecretLabels returns the label, placeholder, and help text for a Slack
// secret field depending on the secrets backend.
func slackSecretLabels(cfg *config.WizardConfig, field string) (label, placeholder, help string) {
	if isDevMode(cfg) {
		switch field {
		case "bot_token":
			return "Bot token value", "Plain text or $HD_ENV_VAR", "Slack bot token value or env var reference"
		case "signing":
			return "Signing secret value", "Plain text or $HD_ENV_VAR", "Slack signing secret value or env var reference"
		case "interaction":
			return "Interaction token value", "Plain text or $HD_ENV_VAR (optional)", "(Optional) Slack interaction token value or env var reference"
		case "slash":
			return "Slash command token value", "Plain text or $HD_ENV_VAR (optional)", "(Optional) Slack slash command token value or env var reference"
		}
	}
	switch field {
	case "bot_token":
		return "Bot token secret path", "Path to Slack bot token", "Path to the Slack bot token secret"
	case "signing":
		return "Signing secret path", "Path to Slack signing secret", "Path to the Slack signing secret"
	case "interaction":
		return "Interaction token", "Slack interaction token (optional)", "(Optional) Slack interaction token"
	case "slash":
		return "Slash command token", "Slack slash command token (optional)", "(Optional) Slack slash command token"
	}
	return "", "", ""
}


// secureExecPlaceholder returns a placeholder string that adapts based on
// whether a secure-exec driver is enabled. When SecureExecDriver is set to
// a non-"none" value, the placeholder shows an hd-lookup path example.
// Otherwise it falls back to the given default placeholder.
func secureExecPlaceholder(cfg *config.WizardConfig, secureExample, defaultPlaceholder string) string {
	if cfg != nil && cfg.SecureExecDriver != "" && cfg.SecureExecDriver != "none" {
		return secureExample
	}
	return defaultPlaceholder
}
// getStepInfo returns the stepInfo for a given step number.
func getStepInfo(step int) *stepInfo {
	switch step {
	case 1:
		return &stepInfo{
			title:    "Step 1: Welcome",
			stepType: stepTypeNavigate,
		}
	case 2:
		return &stepInfo{
			title:    "Step 2: Project Name",
			stepType: stepTypeSingleField,
			fields: []fieldDescriptor{
				{
					label:       "Project name",
					placeholder: "Name for your config project",
					help:        "A short name for your project. Used as the directory name.",
					getValue:    func(c *config.WizardConfig) string { return c.ProjectName },
					setValue:    func(c *config.WizardConfig, v string) { c.ProjectName = v },
				},
			},
		}
	case 3:
		return &stepInfo{
			title:    "Step 3: Config Directory",
			stepType: stepTypeSingleField,
			fields: []fieldDescriptor{
				{
					label:       "Config directory",
					placeholder: "Path where config files will be written",
					help:        "The path where config files will be written. Use absolute or relative path. Defaults to the project name.",
					getValue:    func(c *config.WizardConfig) string { return c.ConfigDir },
					setValue:    func(c *config.WizardConfig, v string) { c.ConfigDir = v },
				},
			},
		}
	case 4:
		return &stepInfo{
			title:        "Step 4: Deployment Mode",
			stepType:     stepTypeRadio,
			radioLabel:   "Deployment mode",
			radioOptions: []string{"docker", "source", "kubernetes"},
			radioGetter:  func(c *config.WizardConfig) string { return c.DeploymentMode },
			radioSetter:  func(c *config.WizardConfig, v string) { c.DeploymentMode = v },
		}
	case 5:
		return &stepInfo{
			title:    "Step 5: Config Repo Setup",
			stepType: stepTypeMultiField,
			fields: []fieldDescriptor{
				{
					label:       "Essentials repo URL",
					placeholder: "Honeydipper essentials config repo",
					help:        "The essentials config repo provides baseline workflows and definitions.",
					getValue:    func(c *config.WizardConfig) string { return c.EssentialsRepoURL },
					setValue:    func(c *config.WizardConfig, v string) { c.EssentialsRepoURL = v },
				},
				{
					label:       "Branch",
					placeholder: "Branch to use",
					help:        "Branch of the essentials repo to use.",
					getValue:    func(c *config.WizardConfig) string { return c.EssentialsBranch },
					setValue:    func(c *config.WizardConfig, v string) { c.EssentialsBranch = v },
				},
				{
					label:       "Clone credential type",
					placeholder: "none, pat, ssh, or github_app",
					help:        "Credential type for cloning private repos. Use 'none' for public repos.",
					getValue:    func(c *config.WizardConfig) string { return c.EssentialsCloneType },
					setValue:    func(c *config.WizardConfig, v string) { c.EssentialsCloneType = v },
				},
				{
					label:       "Token env var name",
					placeholder: "Environment variable containing the token/password",
					help:        "Name of env var that holds the token (for pat type). Will be passed to Docker container.",
					getValue:    func(c *config.WizardConfig) string { return c.EssentialsClonePAT },
					setValue:    func(c *config.WizardConfig, v string) { c.EssentialsClonePAT = v },
					condition:   func(c *config.WizardConfig) bool { return c.EssentialsCloneType == "pat" },
				},
				{
					label:       "Username",
					placeholder: "Git username (optional, defaults to x-access-token)",
					help:        "Username for git authentication (only used with pat type).",
					getValue:    func(c *config.WizardConfig) string { return c.EssentialsCloneKey },
					setValue:    func(c *config.WizardConfig, v string) { c.EssentialsCloneKey = v },
					condition:   func(c *config.WizardConfig) bool { return c.EssentialsCloneType == "pat" },
				},
			},
		}
	case 6:
		return &stepInfo{
			title:        "Step 6: Secrets Backend",
			stepType:     stepTypeRadio,
			radioLabel:   "Secrets backend",
			radioOptions: []string{"vault", "dev"},
			radioGetter:  func(c *config.WizardConfig) string { return c.SecretsBackend },
			radioSetter:  func(c *config.WizardConfig, v string) { c.SecretsBackend = v },
			fields: []fieldDescriptor{
				{
					label:       "Vault address",
					placeholder: "Vault server URL",
					help:        "Vault server URL (only for vault backend)",
					getValue:    func(c *config.WizardConfig) string { return c.VaultAddress },
					setValue:    func(c *config.WizardConfig, v string) { c.VaultAddress = v },
					condition:   func(c *config.WizardConfig) bool { return c.SecretsBackend == "vault" },
				},
				{
					label:       "Auth method",
					placeholder: "Vault auth method (e.g., token, approle)",
					help:        "Vault authentication method",
					getValue:    func(c *config.WizardConfig) string { return c.VaultAuthMethod },
					setValue:    func(c *config.WizardConfig, v string) { c.VaultAuthMethod = v },
					condition:   func(c *config.WizardConfig) bool { return c.SecretsBackend == "vault" },
				},
			},
		}
	case 7:
		return &stepInfo{
			title:        "Step 7: Redis",
			stepType:     stepTypeRadio,
			radioLabel:   "Redis mode",
			radioOptions: []string{"local", "external"},
			radioGetter:  func(c *config.WizardConfig) string { return c.RedisMode },
			radioSetter:  func(c *config.WizardConfig, v string) { c.RedisMode = v },
			fields: []fieldDescriptor{
				{
					label:       "Connection string",
					placeholder: "Redis connection string",
					help:        "Redis connection string (only for external mode)",
					getValue:    func(c *config.WizardConfig) string { return c.RedisConnString },
					setValue:    func(c *config.WizardConfig, v string) { c.RedisConnString = v },
					condition:   func(c *config.WizardConfig) bool { return c.RedisMode == "external" },
				},
			},
		}
	case 8:
		return &stepInfo{
			title:       "Step 8: GitHub Integration",
			stepType:    stepTypeCheckbox,
			checkboxLabel: "GitHub integration type",
			checkboxes: []checkboxOption{
				{
					label:    "GitHub App integration",
					help:     "Use GitHub App for authentication (can be combined with PAT)",
					getValue: func(c *config.WizardConfig) bool { return c.HasGitHubAppIntegration },
					setValue: func(c *config.WizardConfig, v bool) { c.HasGitHubAppIntegration = v },
				},
				{
					label:    "PAT integration",
					help:     "Use Personal Access Token for authentication (can be combined with GitHub App)",
					getValue: func(c *config.WizardConfig) bool { return c.HasGithubPATIntegration },
					setValue: func(c *config.WizardConfig, v bool) { c.HasGithubPATIntegration = v },
				},
			},
			fields: []fieldDescriptor{
				{
					label:       "App ID",
					placeholder: "GitHub App ID",
					help:        "GitHub App ID (only for github_app integration)",
					getValue:    func(c *config.WizardConfig) string { return c.GithubAppID },
					setValue:    func(c *config.WizardConfig, v string) { c.GithubAppID = v },
					condition:   func(c *config.WizardConfig) bool { return c.HasGitHubAppIntegration },
				},
				{
					label:       "Installation ID",
					placeholder: "GitHub App Installation ID",
					help:        "GitHub App Installation ID (only for github_app integration)",
					getValue:    func(c *config.WizardConfig) string { return c.GithubInstallationID },
					setValue:    func(c *config.WizardConfig, v string) { c.GithubInstallationID = v },
					condition:   func(c *config.WizardConfig) bool { return c.HasGitHubAppIntegration },
				},
				{
					label:       "Private key secret path",
					placeholder: "Path to secret containing the private key",
					help:        "Path to the secret containing the GitHub App private key (only for github_app integration)",
					getValue:    func(c *config.WizardConfig) string { return c.GithubKeyPath },
					setValue:    func(c *config.WizardConfig, v string) { c.GithubKeyPath = v },
					condition:   func(c *config.WizardConfig) bool { return c.HasGitHubAppIntegration },
				},
				{
					label:       "Token secret path",
					placeholder: "Path to secret containing the PAT",
					help:        "Path to the secret containing the Personal Access Token (only for PAT integration)",
					getValue:    func(c *config.WizardConfig) string { return c.GithubTokenPath },
					setValue:    func(c *config.WizardConfig, v string) { c.GithubTokenPath = v },
					condition:   func(c *config.WizardConfig) bool { return c.HasGithubPATIntegration },
				},
				{
					label:       "Webhook secret path",
					placeholder: "Path to webhook signing secret",
					help:        "Path to the webhook signing secret (recommended for github_app integration)",
					getValue:    func(c *config.WizardConfig) string { return c.GithubWebhookSecret },
					setValue:    func(c *config.WizardConfig, v string) { c.GithubWebhookSecret = v },
					condition:   func(c *config.WizardConfig) bool { return c.HasGitHubAppIntegration },
				},
			},
		}
	case 9:
		return &stepInfo{
			title:    "Step 9: Slack Integration",
			stepType: stepTypeMultiField,
			fields: []fieldDescriptor{
				{
					label:       "Bot token secret path",
					placeholder: "Path to Slack bot token",
					help:        "Path to the Slack bot token secret",
					getValue:    func(c *config.WizardConfig) string { return c.SlackBotTokenPath },
					setValue:    func(c *config.WizardConfig, v string) { c.SlackBotTokenPath = v },
				},
				{
					label:       "Signing secret path",
					placeholder: "Path to Slack signing secret",
					help:        "Path to the Slack signing secret",
					getValue:    func(c *config.WizardConfig) string { return c.SlackSigningSecretPath },
					setValue:    func(c *config.WizardConfig, v string) { c.SlackSigningSecretPath = v },
				},
				{
					label:       "Interaction token",
					placeholder: "Slack interaction token (optional)",
					help:        "(Optional) Slack interaction token",
					getValue:    func(c *config.WizardConfig) string { return c.SlackInteractionToken },
					setValue:    func(c *config.WizardConfig, v string) { c.SlackInteractionToken = v },
				},
				{
					label:       "Slash command token",
					placeholder: "Slack slash command token (optional)",
					help:        "(Optional) Slack slash command token",
					getValue:    func(c *config.WizardConfig) string { return c.SlackSlashCommandToken },
					setValue:    func(c *config.WizardConfig, v string) { c.SlackSlashCommandToken = v },
				},
			},
		}
	case 10:
		return &stepInfo{
			title:        "Step 10: AI Agent",
			stepType:     stepTypeRadio,
			radioLabel:   "Enable AI agent",
			radioOptions: []string{"yes", "no"},
			radioGetter:  func(c *config.WizardConfig) string { return boolToRadio(c.AIEnabled) },
			radioSetter:  func(c *config.WizardConfig, v string) { c.AIEnabled = radioToBool(v) },
			fields: []fieldDescriptor{
				{
					label:       "API key secret path",
					placeholder: "Path to OpenAI-compatible API key",
					help:        "Path to the API key secret",
					getValue:    func(c *config.WizardConfig) string { return c.AIAPIKeyPath },
					setValue:    func(c *config.WizardConfig, v string) { c.AIAPIKeyPath = v },
					condition:   func(c *config.WizardConfig) bool { return c.AIEnabled },
				},
				{
					label:       "Base URL",
					placeholder: "API base URL",
					help:        "OpenAI-compatible API base URL",
					getValue:    func(c *config.WizardConfig) string { return c.AIBaseURL },
					setValue:    func(c *config.WizardConfig, v string) { c.AIBaseURL = v },
					condition:   func(c *config.WizardConfig) bool { return c.AIEnabled },
				},
				{
					label:       "Model",
					placeholder: "LLM model name",
					help:        "LLM model name (e.g., gpt-4o)",
					getValue:    func(c *config.WizardConfig) string { return c.AIModel },
					setValue:    func(c *config.WizardConfig, v string) { c.AIModel = v },
					condition:   func(c *config.WizardConfig) bool { return c.AIEnabled },
				},
				{
					label:       "Engine name",
					placeholder: "Honeydipper engine name",
					help:        "Honeydipper engine name for the AI agent",
					getValue:    func(c *config.WizardConfig) string { return c.AIEngineName },
					setValue:    func(c *config.WizardConfig, v string) { c.AIEngineName = v },
					condition:   func(c *config.WizardConfig) bool { return c.AIEnabled },
				},
			},
		}
	case 11:
		return &stepInfo{
			title:    "Step 11: Docker Configuration",
			stepType: stepTypeMultiField,
			fields: []fieldDescriptor{
				{
					label:       "Image tag",
					placeholder: "Honeydipper Docker image tag",
					help:        "Docker image tag for Honeydipper",
					getValue:    func(c *config.WizardConfig) string { return c.DockerImageTag },
					setValue:    func(c *config.WizardConfig, v string) { c.DockerImageTag = v },
				},
				{
					label:       "Enable web UI",
					placeholder: "yes or no",
					help:        "Enable the web UI (yes/no)",
					getValue:    func(c *config.WizardConfig) string { return boolToRadio(c.DockerEnableUI) },
					setValue:    func(c *config.WizardConfig, v string) { c.DockerEnableUI = radioToBool(v) },
				},
				{
					label:       "API port",
					placeholder: "API server port",
					help:        "API server port number",
					getValue:    func(c *config.WizardConfig) string { return fmt.Sprintf("%d", c.DockerAPIPort) },
					setValue:    func(c *config.WizardConfig, v string) { c.DockerAPIPort = parseInt(v) },
				},
				{
					label:       "Webhook port",
					placeholder: "Webhook server port",
					help:        "Webhook server port number",
					getValue:    func(c *config.WizardConfig) string { return fmt.Sprintf("%d", c.DockerWebhookPort) },
					setValue:    func(c *config.WizardConfig, v string) { c.DockerWebhookPort = parseInt(v) },
				},
			},
		}
	case 12:
		return &stepInfo{
			title:    "Step 12: Kubernetes Configuration",
			stepType: stepTypeMultiField,
			fields: []fieldDescriptor{
				{
					label:       "Namespace",
					placeholder: "Kubernetes namespace",
					help:        "Kubernetes namespace for Honeydipper",
					getValue:    func(c *config.WizardConfig) string { return c.K8sNamespace },
					setValue:    func(c *config.WizardConfig, v string) { c.K8sNamespace = v },
				},
				{
					label:       "Config repo strategy",
					placeholder: "clone, configmap, or git-sidecar",
					help:        "Strategy for providing config repo in K8s",
					getValue:    func(c *config.WizardConfig) string { return c.K8sRepoStrategy },
					setValue:    func(c *config.WizardConfig, v string) { c.K8sRepoStrategy = v },
				},
			},
		}
	case 13:
		return &stepInfo{
			title:    "Step 13: Source Configuration",
			stepType: stepTypeMultiField,
			fields: []fieldDescriptor{
				{
					label:       "Clone path",
					placeholder: "Path where source will be cloned",
					help:        "Path where Honeydipper source will be cloned",
					getValue:    func(c *config.WizardConfig) string { return c.SourceClonePath },
					setValue:    func(c *config.WizardConfig, v string) { c.SourceClonePath = v },
				},
				{
					label:       "Branch",
					placeholder: "Honeydipper source branch",
					help:        "Honeydipper source branch to use",
					getValue:    func(c *config.WizardConfig) string { return c.SourceBranch },
					setValue:    func(c *config.WizardConfig, v string) { c.SourceBranch = v },
				},
			},
		}
	case 14:
		return &stepInfo{
			title:       "Step 14: GitHub Repo Creation",
			stepType:    stepTypeCheckbox,
			checkboxLabel: "Create GitHub repo",
			checkboxes: []checkboxOption{
				{
					label:    "Create GitHub repo",
					help:     "Create a new GitHub repository for the generated config",
					getValue: func(c *config.WizardConfig) bool { return c.GithubCreateRepo },
					setValue: func(c *config.WizardConfig, v bool) { c.GithubCreateRepo = v },
				},
				{
					label:    "Use local copy instead of clone",
					help:     "Skip git clone and use the local config directory as the REPO (only when creating repo)",
					getValue: func(c *config.WizardConfig) bool { return c.UseLocalCopy },
					setValue: func(c *config.WizardConfig, v bool) { c.UseLocalCopy = v },
					condition: func(c *config.WizardConfig) bool { return c.GithubCreateRepo },
				},
			},
			fields: []fieldDescriptor{
				{
					label:       "Git remote URL",
					placeholder: "git@github.com:user/repo.git or https://github.com/user/repo.git",
					help:        "Git remote URL for the config repository (required, used as REPO env var in docker-compose)",
					getValue:    func(c *config.WizardConfig) string { return c.GitRemoteURL },
					setValue:    func(c *config.WizardConfig, v string) { c.GitRemoteURL = v },
					condition:   func(c *config.WizardConfig) bool { return c.GithubCreateRepo },
				},
				{
					label:       "Secure-exec driver",
					placeholder: "none, hd-driver-vault, or gcloud-secret",
					help:        "Secure execution driver: none (disabled), hd-driver-vault (Vault secrets), or gcloud-secret (Google Cloud Secret Manager)",
					getValue:    func(c *config.WizardConfig) string { return c.SecureExecDriver },
					setValue:    func(c *config.WizardConfig, v string) { c.SecureExecDriver = v },
					condition:   func(c *config.WizardConfig) bool { return c.GithubCreateRepo && !c.UseLocalCopy },
				},
				{
					label:       "Vault server address",
					placeholder: "https://vault.example.com:8200",
					help:        "VAULT_ADDR - Vault server address for the secure-exec driver",
					getValue:    func(c *config.WizardConfig) string { return c.SecureExecVaultAddr },
					setValue:    func(c *config.WizardConfig, v string) { c.SecureExecVaultAddr = v },
					condition:   func(c *config.WizardConfig) bool { return c.GithubCreateRepo && !c.UseLocalCopy && c.SecureExecDriver == "hd-driver-vault" },
				},
			},
		}
	case 15:
		return &stepInfo{
			title:    "Step 15: Clone Auth",
			stepType: stepTypeMultiField,
			fields: []fieldDescriptor{
				{
					label:       "Clone auth method",
					placeholder: "none, pat, github_app, or ssh",
					help:        "Authentication method for cloning the config repository (required when creating repo without local copy)",
					getValue:    func(c *config.WizardConfig) string { return c.ConfigRepoCloneAuth },
					setValue:    func(c *config.WizardConfig, v string) { c.ConfigRepoCloneAuth = v },
					condition:   func(c *config.WizardConfig) bool { return c.GithubCreateRepo && !c.UseLocalCopy },
				},
				{
					label:       "PAT",
					placeholderFunc: func(cfg *config.WizardConfig) string {
						return secureExecPlaceholder(cfg, "hd-lookup:/secrets/data/project/pat", "ghp_xxxxx or $MY_PAT")
					},
					help:        "Personal Access Token value, or $ENV_VAR reference to an environment variable holding the PAT",
					getValue:    func(c *config.WizardConfig) string { return c.ConfigRepoPATValue },
					setValue:    func(c *config.WizardConfig, v string) { c.ConfigRepoPATValue = v },
					condition:   func(c *config.WizardConfig) bool { return c.GithubCreateRepo && !c.UseLocalCopy && c.ConfigRepoCloneAuth == "pat" },
				},
				{
					label:       "GitHub App ID",
					placeholderFunc: func(cfg *config.WizardConfig) string {
						return secureExecPlaceholder(cfg, "hd-lookup:/secrets/data/project/gh_app_id", "12345")
					},
					help:        "GitHub App ID for authentication",
					getValue:    func(c *config.WizardConfig) string { return c.ConfigRepoGHAppID },
					setValue:    func(c *config.WizardConfig, v string) { c.ConfigRepoGHAppID = v },
					condition:   func(c *config.WizardConfig) bool { return c.GithubCreateRepo && !c.UseLocalCopy && c.ConfigRepoCloneAuth == "github_app" },
				},
				{
					label:       "Installation ID",
					placeholderFunc: func(cfg *config.WizardConfig) string {
						return secureExecPlaceholder(cfg, "hd-lookup:/secrets/data/project/gh_install_id", "67890")
					},
					help:        "GitHub App Installation ID",
					getValue:    func(c *config.WizardConfig) string { return c.ConfigRepoGHInstallID },
					setValue:    func(c *config.WizardConfig, v string) { c.ConfigRepoGHInstallID = v },
					condition:   func(c *config.WizardConfig) bool { return c.GithubCreateRepo && !c.UseLocalCopy && c.ConfigRepoCloneAuth == "github_app" },
				},
				{
					label:       "Private key",
					placeholderFunc: func(cfg *config.WizardConfig) string {
						return secureExecPlaceholder(cfg, "hd-lookup:/secrets/data/project/gh_private_key", "-----BEGIN RSA PRIVATE KEY-----\n...")
					},
					help:        "GitHub App private key content",
					getValue:    func(c *config.WizardConfig) string { return c.ConfigRepoGHAppKey },
					setValue:    func(c *config.WizardConfig, v string) { c.ConfigRepoGHAppKey = v },
					condition:   func(c *config.WizardConfig) bool { return c.GithubCreateRepo && !c.UseLocalCopy && c.ConfigRepoCloneAuth == "github_app" && (c.SecureExecDriver == "none" || c.SecureExecDriver == "") },
				},
				{
					label:       "SSH key content",
					placeholderFunc: func(cfg *config.WizardConfig) string {
						return secureExecPlaceholder(cfg, "hd-lookup:/secrets/data/project/ssh_key", "-----BEGIN OPENSSH PRIVATE KEY-----\n...")
					},
					help:        "SSH private key content (used as DIPPER_SSH_KEY)",
					getValue:    func(c *config.WizardConfig) string { return c.ConfigRepoSSHKey },
					setValue:    func(c *config.WizardConfig, v string) { c.ConfigRepoSSHKey = v },
					condition:   func(c *config.WizardConfig) bool { return c.GithubCreateRepo && !c.UseLocalCopy && c.ConfigRepoCloneAuth == "ssh" && (c.SecureExecDriver == "none" || c.SecureExecDriver == "") },
				},
				{
					label:       "SSH key file path",
					placeholder: "/home/user/.ssh/id_rsa",
					help:        "Path to SSH private key file (used as DIPPER_SSH_FILE, alternative to inline key)",
					getValue:    func(c *config.WizardConfig) string { return c.ConfigRepoSSHFile },
					setValue:    func(c *config.WizardConfig, v string) { c.ConfigRepoSSHFile = v },
					condition:   func(c *config.WizardConfig) bool { return c.GithubCreateRepo && !c.UseLocalCopy && c.ConfigRepoCloneAuth == "ssh" },
				},
				{
					label:       "SSH key passphrase env var",
					placeholder: "e.g. SSH_KEY_PASS",
					help:        "Environment variable name for SSH key passphrase (optional)",
					getValue:    func(c *config.WizardConfig) string { return c.ConfigRepoSSHKeyPassEnv },
					setValue:    func(c *config.WizardConfig, v string) { c.ConfigRepoSSHKeyPassEnv = v },
					condition:   func(c *config.WizardConfig) bool { return c.GithubCreateRepo && !c.UseLocalCopy && c.ConfigRepoCloneAuth == "ssh" },
				},
				// Secure-exec secret path fields (shown when a secure-exec driver is active)
				{
					label:       "PAT secret path (hd-lookup: prepended automatically)",
					placeholder: "/secrets/data/project/pat",
					help:        "Secret path for the PAT when using secure-exec driver (hd-lookup: prefix will be added automatically)",
					getValue:    func(c *config.WizardConfig) string { return c.ConfigRepoPATPath },
					setValue:    func(c *config.WizardConfig, v string) { c.ConfigRepoPATPath = v },
					condition:   func(c *config.WizardConfig) bool { return c.GithubCreateRepo && !c.UseLocalCopy && c.ConfigRepoCloneAuth == "pat" && c.SecureExecDriver != "none" && c.SecureExecDriver != "" },
				},
				{
					label:       "GitHub App key secret path (hd-lookup: prepended automatically)",
					placeholder: "/secrets/data/project/gh_app_key",
					help:        "Secret path for the GitHub App private key when using secure-exec driver (hd-lookup: prefix will be added automatically)",
					getValue:    func(c *config.WizardConfig) string { return c.ConfigRepoGHAppKeyPath },
					setValue:    func(c *config.WizardConfig, v string) { c.ConfigRepoGHAppKeyPath = v },
					condition:   func(c *config.WizardConfig) bool { return c.GithubCreateRepo && !c.UseLocalCopy && c.ConfigRepoCloneAuth == "github_app" && c.SecureExecDriver != "none" && c.SecureExecDriver != "" },
				},
				{
					label:       "SSH key secret path (hd-lookup: prepended automatically)",
					placeholder: "/secrets/data/project/ssh_key",
					help:        "Secret path for the SSH key when using secure-exec driver (hd-lookup: prefix will be added automatically)",
					getValue:    func(c *config.WizardConfig) string { return c.ConfigRepoSSHKeyPath },
					setValue:    func(c *config.WizardConfig, v string) { c.ConfigRepoSSHKeyPath = v },
					condition:   func(c *config.WizardConfig) bool { return c.GithubCreateRepo && !c.UseLocalCopy && c.ConfigRepoCloneAuth == "ssh" && c.SecureExecDriver != "none" && c.SecureExecDriver != "" },
				},
			},
		}
	case 16:
		return &stepInfo{
			title:    "Step 16: Summary & Confirm",
			stepType: stepTypeNavigate,
		}
	default:
		return nil
	}
}

// isStepRequired returns whether the given step should be shown based on the
// current configuration.  Steps 11 (Docker), 12 (Kubernetes), and 13 (Source)
// are deployment-mode-specific and are only required when the matching mode is selected.
func isStepRequired(step int, cfg *config.WizardConfig) bool {
	switch step {
	case 11: // Docker configuration
		return cfg.DeploymentMode == "docker"
	case 12: // Kubernetes configuration
		return cfg.DeploymentMode == "kubernetes"
	case 13: // Source configuration
		return cfg.DeploymentMode == "source"
	default:
		return true
	}
}

// nextStep returns the next required step after the given step,
// or the current step if there is no next required step.
func nextStep(step int, cfg *config.WizardConfig) int {
	for s := step + 1; s <= stepCount; s++ {
		if isStepRequired(s, cfg) {
			return s
		}
	}
	return step
}

// prevStep returns the previous required step before the given step,
// or the current step if there is no previous required step.
func prevStep(step int, cfg *config.WizardConfig) int {
	for s := step - 1; s >= 1; s-- {
		if isStepRequired(s, cfg) {
			return s
		}
	}
	return step
}

// ghSecretFieldKey maps a GitHub integration field label to a secret key
// recognized by ghSecretLabels.  Returns "" for non-secret fields.
func ghSecretFieldKey(label string) string {
	switch label {
	case "Private key secret path", "Private key value":
		return "key"
	case "Token secret path", "Token value":
		return "token"
	case "Webhook secret path", "Webhook secret value":
		return "webhook"
	default:
		return ""
	}
}

// slackSecretFieldKey maps a Slack integration field label to a secret key
// recognized by slackSecretLabels.  Returns "" for non-secret fields.
func slackSecretFieldKey(label string) string {
	switch label {
	case "Bot token secret path", "Bot token value":
		return "bot_token"
	case "Signing secret path", "Signing secret value":
		return "signing"
	case "Interaction token", "Interaction token value":
		return "interaction"
	case "Slash command token", "Slash command token value":
		return "slash"
	default:
		return ""
	}
}

// isDevSecretField returns true if the given step and field label is a secret
// field in dev mode (i.e. a field whose value should be validated as either
// a plain value or an $HD_* env var reference).
func isDevSecretField(step int, label string) bool {
	switch step {
	case 8:
		return ghSecretFieldKey(label) != ""
	case 9:
		return slackSecretFieldKey(label) != ""
	default:
		return false
	}
}

// validateDevSecretValue validates a secret field value in dev mode.
// If the value starts with "$", it must be "$HD_SOMETHING".
// Plain values (no "$" prefix) are accepted as-is.
// Returns an error message if validation fails, or "" if valid.
func validateDevSecretValue(value string) string {
	v := strings.TrimSpace(value)
	if v == "" {
		return ""
	}
	if !strings.HasPrefix(v, "$") {
		// Plain value — accepted as-is in dev mode
		return ""
	}
	// Starts with "$" — must be $HD_SOMETHING
	if !strings.HasPrefix(v, "$HD_") {
		return fmt.Sprintf("env var reference %q must use $HD_ prefix (only HD_* env vars are supported for interpolation)", v)
	}
	// Strip the "$" prefix and store just the variable name
	return ""
}

// stripEnvVarPrefix strips the leading "$" from an env var reference.
// e.g. "$HD_GITHUB_TOKEN" → "HD_GITHUB_TOKEN"
// If the value doesn't start with "$", it is returned unchanged.
func stripEnvVarPrefix(value string) string {
	if strings.HasPrefix(value, "$") {
		return value[1:]
	}
	return value
}
