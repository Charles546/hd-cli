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
	ConfigDirAbs   string `yaml:"-"`               // Absolute path of ConfigDir, computed during generation
	DeploymentMode string `yaml:"deployment_mode"` // docker, source, kubernetes

	// Essentials repo
	EssentialsRepoURL string `yaml:"essentials_repo_url"`
	EssentialsBranch  string `yaml:"essentials_branch"`
	EssentialsPath    string `yaml:"essentials_path,omitempty"`

	// Bootstrap repo clone credentials (used in init.yaml template)
	EssentialsCloneType      string `yaml:"essentials_clone_type"`       // none, pat, ssh, github_app
	EssentialsClonePAT       string `yaml:"essentials_clone_pat"`        // env var name containing the PAT
	EssentialsCloneKey       string `yaml:"essentials_clone_key"`       // SSH key file path
	EssentialsCloneKeyPassEnv string `yaml:"essentials_clone_key_pass_env"` // env var for SSH key passphrase

	// Bootstrap clone credential type for docker-compose.yaml (kept for backward compat)
	BootstrapCloneCredentialType string `yaml:"bootstrap_clone_credential_type"` // none, github_app, pat

	// GitHub integration
	HasGithubPATIntegration   bool   `yaml:"has_github_pat_integration"`
	HasGitHubAppIntegration   bool   `yaml:"has_github_app_integration"`
	GithubAppID               string `yaml:"github_app_id"`
	GithubInstallationID      string `yaml:"github_installation_id"`
	GithubKeyPath             string `yaml:"github_key_path"`
	GithubTokenPath           string `yaml:"github_token_path"`
	GithubWebhookSecret       string `yaml:"github_webhook_secret_path"`

	// Slack integration
	SlackBotTokenPath         string `yaml:"slack_bot_token_path"`
	SlackSigningSecretPath    string `yaml:"slack_signing_secret_path"`
	SlackInteractionToken     string `yaml:"slack_interaction_token"`  // renamed from SlackInteractionTokenPath
	SlackSlashCommandToken    string `yaml:"slack_slash_command_token"` // renamed from SlackSlashCommandTokenPath

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

	// Dev mode env vars tracks which HD_* variables are referenced during the wizard.
	// Used by the docker-compose template to pass them into the container.
	DevEnvVars []string `yaml:"dev_env_vars,omitempty"`
}

// NewDefaultWizardConfig returns a WizardConfig pre-filled with sensible defaults.
func NewDefaultWizardConfig() *WizardConfig {
	return &WizardConfig{
		ProjectName:                  "hd-config",
		ConfigDir:                    "./hd-config",
		DeploymentMode:               "docker",
		EssentialsRepoURL:            "https://github.com/honeydipper/honeydipper-config-essentials.git",
		EssentialsBranch:             "v4-rc",
		EssentialsCloneType:          "none",
		BootstrapCloneCredentialType: "none",
		HasGithubPATIntegration:      true,
		SecretsBackend:               "vault",
		RedisMode:                    "local",
		DockerImageTag:               "v4-latest",
		DockerEnableUI:               false,
		DockerAPIPort:                9000,
		DockerWebhookPort:            8080,
		K8sNamespace:                 "honeydipper",
		K8sRepoStrategy:              "clone",
		SourceBranch:                 "v4",
		AIModel:                      "gpt-4o",
		AIBaseURL:                    "https://api.openai.com/v1",
		AIEngineName:                 "default",
	}
}

// CollectDevEnvVars scans the wizard config for all $HD_* env var references
// used in dev mode and returns a deduplicated list of variable names (with HD_ prefix).
func (c *WizardConfig) CollectDevEnvVars() []string {
	if c == nil || c.SecretsBackend != "dev" {
		return nil
	}

	seen := make(map[string]bool)
	var vars []string

	fields := []string{
		c.GithubTokenPath,
		c.GithubKeyPath,
		c.GithubWebhookSecret,
		c.SlackBotTokenPath,
		c.SlackSigningSecretPath,
		c.SlackInteractionToken,
		c.SlackSlashCommandToken,
	}

	for _, v := range fields {
		if len(v) > 4 && v[:4] == "$HD_" {
			name := v[1:] // strip "$"
			if !seen[name] {
				seen[name] = true
				vars = append(vars, name)
			}
		}
	}

	return vars
}
