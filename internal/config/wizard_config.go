// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If a separate written commercial agreement exists, you may use this file under those terms instead.

package config

// WizardConfig holds all the user's choices during the init wizard.
// It is consumed by the config generator to render the Honeydipper config files.
type WizardConfig struct {
	// Project settings
	ProjectName    string `yaml:"project_name"`
	ConfigDir      string `yaml:"config_dir"`
	DeploymentMode string `yaml:"deployment_mode"` // docker, source, kubernetes

	// Essentials repo
	EssentialsRepoURL string `yaml:"essentials_repo_url"`
	EssentialsBranch  string `yaml:"essentials_branch"`

	// GitHub integration
	GithubIntegrationType string `yaml:"github_integration_type"` // github_app, pat
	GithubAppID           string `yaml:"github_app_id"`
	GithubInstallationID  string `yaml:"github_installation_id"`
	GithubKeyPath         string `yaml:"github_key_path"`
	GithubTokenPath       string `yaml:"github_token_path"`
	GithubWebhookSecret   string `yaml:"github_webhook_secret_path"`

	// Slack integration
	SlackBotTokenPath          string `yaml:"slack_bot_token_path"`
	SlackSigningSecretPath     string `yaml:"slack_signing_secret_path"`
	SlackInteractionTokenPath  string `yaml:"slack_interaction_token_path"`  // optional
	SlackSlashCommandTokenPath string `yaml:"slack_slash_command_token_path"` // optional

	// AI agent
	AIEnabled    bool   `yaml:"ai_enabled"`
	AIAPIKeyPath string `yaml:"ai_api_key_path"`
	AIBaseURL    string `yaml:"ai_base_url"`
	AIModel      string `yaml:"ai_model"`
	AIEngineName string `yaml:"ai_engine_name"`

	// Secrets backend
	SecretsBackend  string `yaml:"secrets_backend"` // vault, dev
	VaultAddress    string `yaml:"vault_address"`
	VaultAuthMethod string `yaml:"vault_auth_method"`

	// Redis
	RedisMode       string `yaml:"redis_mode"` // local, external
	RedisConnString string `yaml:"redis_connection_string,omitempty"`

	// Docker-specific
	DockerImageTag    string `yaml:"docker_image_tag,omitempty"`
	DockerEnableUI    bool   `yaml:"docker_enable_ui,omitempty"`
	DockerAPIPort     int    `yaml:"docker_api_port,omitempty"`
	DockerWebhookPort int    `yaml:"docker_webhook_port,omitempty"`

	// Kubernetes-specific
	K8sNamespace    string `yaml:"k8s_namespace,omitempty"`
	K8sRepoStrategy string `yaml:"k8s_repo_strategy,omitempty"`

	// Source-specific
	SourceClonePath string `yaml:"source_clone_path,omitempty"`
	SourceBranch    string `yaml:"source_branch,omitempty"`

	// GitHub repo creation
	GithubCreateRepo bool   `yaml:"github_create_repo,omitempty"`
	GithubRepoName   string `yaml:"github_repo_name,omitempty"`
	GithubRepoVis    string `yaml:"github_repo_visibility,omitempty"` // private, public
}

// NewDefaultWizardConfig returns a WizardConfig pre-filled with sensible defaults.
func NewDefaultWizardConfig() *WizardConfig {
	return &WizardConfig{
		ProjectName:           "hd-config",
		ConfigDir:             "./hd-config",
		DeploymentMode:        "docker",
		EssentialsRepoURL:     "https://github.com/honeydipper/honeydipper-config-essentials.git",
		EssentialsBranch:      "v4-rc",
		GithubIntegrationType: "pat",
		SecretsBackend:        "vault",
		RedisMode:             "local",
		DockerImageTag:        "v4-latest",
		DockerEnableUI:        false,
		DockerAPIPort:         9000,
		DockerWebhookPort:     8080,
		K8sNamespace:          "honeydipper",
		K8sRepoStrategy:       "clone",
		SourceBranch:          "v4",
		AIModel:               "gpt-4o",
		AIBaseURL:             "https://api.openai.com/v1",
		AIEngineName:          "default",
	}
}
