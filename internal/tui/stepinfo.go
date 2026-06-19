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
)

// fieldDescriptor describes a single text input field within a step.
type fieldDescriptor struct {
	label       string
	placeholder string
	help        string
	getValue    func(*config.WizardConfig) string
	setValue    func(*config.WizardConfig, string)
}

// stepInfo describes the structure and behavior of a wizard step.
type stepInfo struct {
	title        string
	stepType     stepType
	fields       []fieldDescriptor
	radioLabel   string
	radioOptions []string
	radioGetter  func(*config.WizardConfig) string
	radioSetter  func(*config.WizardConfig, string)
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
					help:        "The path where config files will be written. Use absolute or relative path.",
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
			},
		}
	case 6:
		return &stepInfo{
			title:    "Step 6: GitHub Integration",
			stepType: stepTypeRadio,
			radioLabel: "Integration type",
			radioOptions: []string{"github_app", "pat"},
			radioGetter: func(c *config.WizardConfig) string { return c.GithubIntegrationType },
			radioSetter: func(c *config.WizardConfig, v string) { c.GithubIntegrationType = v },
			fields: []fieldDescriptor{
				{
					label:       "App ID",
					placeholder: "GitHub App ID",
					help:        "GitHub App ID (only for github_app integration)",
					getValue:    func(c *config.WizardConfig) string { return c.GithubAppID },
					setValue:    func(c *config.WizardConfig, v string) { c.GithubAppID = v },
				},
				{
					label:       "Installation ID",
					placeholder: "GitHub App Installation ID",
					help:        "GitHub App Installation ID (only for github_app integration)",
					getValue:    func(c *config.WizardConfig) string { return c.GithubInstallationID },
					setValue:    func(c *config.WizardConfig, v string) { c.GithubInstallationID = v },
				},
				{
					label:       "Private key secret path",
					placeholder: "Path to secret containing the private key",
					help:        "Path to the secret containing the GitHub App private key",
					getValue:    func(c *config.WizardConfig) string { return c.GithubKeyPath },
					setValue:    func(c *config.WizardConfig, v string) { c.GithubKeyPath = v },
				},
				{
					label:       "Token secret path",
					placeholder: "Path to secret containing the PAT",
					help:        "Path to the secret containing the Personal Access Token",
					getValue:    func(c *config.WizardConfig) string { return c.GithubTokenPath },
					setValue:    func(c *config.WizardConfig, v string) { c.GithubTokenPath = v },
				},
				{
					label:       "Webhook secret path",
					placeholder: "Path to webhook signing secret",
					help:        "Path to the webhook signing secret",
					getValue:    func(c *config.WizardConfig) string { return c.GithubWebhookSecret },
					setValue:    func(c *config.WizardConfig, v string) { c.GithubWebhookSecret = v },
				},
			},
		}
	case 7:
		return &stepInfo{
			title:    "Step 7: Slack Integration",
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
					label:       "Interaction token path",
					placeholder: "Slack interaction token (optional)",
					help:        "(Optional) Path to Slack interaction token",
					getValue:    func(c *config.WizardConfig) string { return c.SlackInteractionTokenPath },
					setValue:    func(c *config.WizardConfig, v string) { c.SlackInteractionTokenPath = v },
				},
				{
					label:       "Slash command token path",
					placeholder: "Slack slash command token (optional)",
					help:        "(Optional) Path to Slack slash command token",
					getValue:    func(c *config.WizardConfig) string { return c.SlackSlashCommandTokenPath },
					setValue:    func(c *config.WizardConfig, v string) { c.SlackSlashCommandTokenPath = v },
				},
			},
		}
	case 8:
		return &stepInfo{
			title:    "Step 8: AI Agent",
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
	case 9:
		return &stepInfo{
			title:    "Step 9: Secrets Backend",
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
	case 10:
		return &stepInfo{
			title:    "Step 10: Redis",
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
