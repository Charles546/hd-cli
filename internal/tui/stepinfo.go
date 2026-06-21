// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If a separate written commercial agreement exists, you may use this file under those terms instead.

package tui

import (
	"strconv"

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
}

// checkboxOption describes a single checkbox toggle within a step.
type checkboxOption struct {
	label    string
	help     string
	getValue func(*config.WizardConfig) bool
	setValue func(*config.WizardConfig, bool)
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
			return "Private key value", "Plain text or $ENV_VAR", "GitHub App private key value or env var name (only for github_app integration)"
		case "token":
			return "Token value", "Plain text or $ENV_VAR", "Personal Access Token value or env var name (only for PAT integration)"
		case "webhook":
			return "Webhook secret value", "Plain text or $ENV_VAR", "Webhook signing secret value or env var name (recommended for github_app integration)"
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
			return "Bot token value", "Plain text or $ENV_VAR", "Slack bot token value or env var name"
		case "signing":
			return "Signing secret value", "Plain text or $ENV_VAR", "Slack signing secret value or env var name"
		case "interaction":
			return "Interaction token value", "Plain text or $ENV_VAR (optional)", "(Optional) Slack interaction token value or env var name"
		case "slash":
			return "Slash command token value", "Plain text or $ENV_VAR (optional)", "(Optional) Slack slash command token value or env var name"
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
			title:    "Step 4: Deployment Mode",
			stepType: stepTypeRadio,
			radioLabel: "Deployment mode",
			radioOptions: []string{"docker", "source", "kubernetes"},
			radioGetter: func(c *config.WizardConfig) string { return c.DeploymentMode },
			radioSetter: func(c *config.WizardConfig, v string) { c.DeploymentMode = v },
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
			title:    "Step 6: Secrets Backend",
			stepType: stepTypeRadio,
			radioLabel: "Secrets backend",
			radioOptions: []string{"vault", "dev"},
			radioGetter: func(c *config.WizardConfig) string { return c.SecretsBackend },
			radioSetter: func(c *config.WizardConfig, v string) { c.SecretsBackend = v },
			fields: []fieldDescriptor{
				{
					label:       "Vault address",
					placeholder: "Vault server URL",
					help:        "Vault server URL (only for vault backend)",
					getValue:    func(c *config.WizardConfig) string { return c.VaultAddress },
					setValue:    func(c *config.WizardConfig, v string) { c.VaultAddress = v },
				},
				{
					label:       "Auth method",
					placeholder: "Vault auth method (e.g., token, approle)",
					help:        "Vault authentication method",
					getValue:    func(c *config.WizardConfig) string { return c.VaultAuthMethod },
					setValue:    func(c *config.WizardConfig, v string) { c.VaultAuthMethod = v },
				},
			},
		}
	case 7:
		return &stepInfo{
			title:    "Step 7: Redis",
			stepType: stepTypeRadio,
			radioLabel: "Redis mode",
			radioOptions: []string{"local", "external"},
			radioGetter: func(c *config.WizardConfig) string { return c.RedisMode },
			radioSetter: func(c *config.WizardConfig, v string) { c.RedisMode = v },
			fields: []fieldDescriptor{
				{
					label:       "Connection string",
					placeholder: "Redis connection string",
					help:        "Redis connection string (only for external mode)",
					getValue:    func(c *config.WizardConfig) string { return c.RedisConnString },
					setValue:    func(c *config.WizardConfig, v string) { c.RedisConnString = v },
				},
			},
		}
	case 8:
		return &stepInfo{
			title:    "Step 8: GitHub Integration",
			stepType: stepTypeCheckbox,
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
			title:    "Step 10: AI Agent",
			stepType: stepTypeRadio,
			radioLabel: "Enable AI agent",
			radioOptions: []string{"yes", "no"},
			radioGetter: func(c *config.WizardConfig) string { return boolToRadio(c.AIEnabled) },
			radioSetter: func(c *config.WizardConfig, v string) { c.AIEnabled = radioToBool(v) },
			fields: []fieldDescriptor{
				{
					label:       "API key secret path",
					placeholder: "Path to OpenAI-compatible API key",
					help:        "Path to the API key secret",
					getValue:    func(c *config.WizardConfig) string { return c.AIAPIKeyPath },
					setValue:    func(c *config.WizardConfig, v string) { c.AIAPIKeyPath = v },
				},
				{
					label:       "Base URL",
					placeholder: "API base URL",
					help:        "OpenAI-compatible API base URL",
					getValue:    func(c *config.WizardConfig) string { return c.AIBaseURL },
					setValue:    func(c *config.WizardConfig, v string) { c.AIBaseURL = v },
				},
				{
					label:       "Model",
					placeholder: "LLM model name",
					help:        "LLM model name (e.g., gpt-4o)",
					getValue:    func(c *config.WizardConfig) string { return c.AIModel },
					setValue:    func(c *config.WizardConfig, v string) { c.AIModel = v },
				},
				{
					label:       "Engine name",
					placeholder: "Honeydipper engine name",
					help:        "Honeydipper engine name for the AI agent",
					getValue:    func(c *config.WizardConfig) string { return c.AIEngineName },
					setValue:    func(c *config.WizardConfig, v string) { c.AIEngineName = v },
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
					getValue:    func(c *config.WizardConfig) string { return strconv.Itoa(c.DockerAPIPort) },
					setValue:    func(c *config.WizardConfig, v string) { c.DockerAPIPort = parseInt(v) },
				},
				{
					label:       "Webhook port",
					placeholder: "Webhook server port",
					help:        "Webhook server port number",
					getValue:    func(c *config.WizardConfig) string { return strconv.Itoa(c.DockerWebhookPort) },
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
			title:    "Step 14: GitHub Repo Creation",
			stepType: stepTypeRadio,
			radioLabel: "Create GitHub repo",
			radioOptions: []string{"yes", "no"},
			radioGetter: func(c *config.WizardConfig) string { return boolToRadio(c.GithubCreateRepo) },
			radioSetter: func(c *config.WizardConfig, v string) { c.GithubCreateRepo = radioToBool(v) },
			fields: []fieldDescriptor{
				{
					label:       "Repo name",
					placeholder: "Repository name",
					help:        "Name for the new GitHub repository",
					getValue:    func(c *config.WizardConfig) string { return c.GithubRepoName },
					setValue:    func(c *config.WizardConfig, v string) { c.GithubRepoName = v },
				},
				{
					label:       "Visibility",
					placeholder: "private or public",
					help:        "Repository visibility",
					getValue:    func(c *config.WizardConfig) string { return c.GithubRepoVis },
					setValue:    func(c *config.WizardConfig, v string) { c.GithubRepoVis = v },
				},
			},
		}
	case 15:
		return &stepInfo{
			title:    "Step 15: Summary & Confirm",
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
