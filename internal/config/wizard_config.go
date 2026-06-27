// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If a separate written commercial agreement exists, you may use this file under those terms instead.

package config

import (
	"fmt"
	"strings"
)

// WizardConfig holds all the user's choices during the init wizard.
// It is consumed by the config generator to render the Honeydipper config files.
type WizardConfig struct {
	// Project settings
	ProjectName    string `yaml:"project_name"`
	ConfigDir      string `yaml:"config_dir"`
	ConfigDirAbs   string `yaml:"-"`                 // Absolute path of ConfigDir, computed during generation
	DeploymentMode string `yaml:"deployment_mode"` // docker, source, kubernetes

	// Essentials repo
	EssentialsRepoURL string `yaml:"essentials_repo_url"`
	EssentialsBranch  string `yaml:"essentials_branch"`
	EssentialsPath    string `yaml:"essentials_path,omitempty"`

	// Bootstrap repo clone credentials (used in init.yaml template)
	EssentialsCloneType       string `yaml:"essentials_clone_type"`          // none, pat, ssh, github_app
	EssentialsClonePAT        string `yaml:"essentials_clone_pat"`           // env var name containing the PAT
	EssentialsCloneKey        string `yaml:"essentials_clone_key"`           // SSH key file path
	EssentialsCloneKeyPassEnv string `yaml:"essentials_clone_key_pass_env"`  // env var for SSH key passphrase

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
	SlackBotTokenPath      string `yaml:"slack_bot_token_path"`
	SlackSigningSecretPath string `yaml:"slack_signing_secret_path"`
	SlackInteractionToken  string `yaml:"slack_interaction_token"`  // renamed from SlackInteractionTokenPath
	SlackSlashCommandToken string `yaml:"slack_slash_command_token"` // renamed from SlackSlashCommandTokenPath

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
	GitRemoteURL     string `yaml:"git_remote_url,omitempty"`        // git remote URL for the config repo
	UseLocalCopy     bool   `yaml:"use_local_copy,omitempty"`        // use local copy instead of cloning
	GitInit          bool   `yaml:"git_init,omitempty"`               // whether to git init the config dir

	// Config repo clone authentication (used when GithubCreateRepo && !UseLocalCopy)
	ConfigRepoCloneAuth     string            `yaml:"config_repo_clone_auth"`                // none, pat, github_app, ssh
	ConfigRepoPATValue      string            `yaml:"config_repo_pat_value"`                 // PAT value or $ENV_VAR reference
	ConfigRepoGHAppID       string            `yaml:"config_repo_gh_app_id"`
	ConfigRepoGHInstallID   string            `yaml:"config_repo_gh_installation_id"`
	ConfigRepoGHAppKey      string            `yaml:"config_repo_gh_app_key"`
	ConfigRepoSSHKey        string            `yaml:"config_repo_ssh_key"`                   // inline SSH key content
	ConfigRepoSSHFile       string            `yaml:"config_repo_ssh_file"`                  // path to SSH key file
	ConfigRepoSSHKeyPassEnv string            `yaml:"config_repo_ssh_key_pass_env"`          // env var name for key passphrase
	ConfigRepoPATPath       string            `yaml:"config_repo_pat_path"`                  // hd-lookup path for PAT (secure-exec mode)
	ConfigRepoGHAppKeyPath  string            `yaml:"config_repo_gh_app_key_path"`           // hd-lookup path for GH App private key (secure-exec mode)
	ConfigRepoSSHKeyPath    string            `yaml:"config_repo_ssh_key_path"`              // hd-lookup path for SSH key (secure-exec mode)
	ConfigRepoCloneEnvVars  map[string]string `yaml:"config_repo_clone_env_vars,omitempty"`  // derived, for template

	// Secure execution settings
	SecureExecDriver     string `yaml:"secure_exec_driver,omitempty"`      // none, hd-driver-vault, gcloud-secret
	SecureExecVaultAddr  string `yaml:"secure_exec_vault_addr,omitempty"`   // VAULT_ADDR for vault driver
	SecureExecVaultToken string `yaml:"secure_exec_vault_token,omitempty"` // VAULT_TOKEN for vault driver (optional)

	// Secure exec env vars map for docker-compose template (derived from the fields above)
	SecureExecEnvVars map[string]string `yaml:"secure_exec_env_vars,omitempty"` // derived, for template

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
		ConfigRepoCloneAuth:          "none",
		SecureExecDriver:             "none",
		AIModel:                      "gpt-4o",
		AIBaseURL:                    "https://api.openai.com/v1",
		AIEngineName:                 "default",
	}
}

// BuildConfigRepoCloneEnvVars returns a map of env var names to their placeholder
// values for the selected config repo clone authentication method.
// hd-lookup prefix handling: For PAT values starting with "$HD_", the value is
// treated as an hd-lookup reference and passed through directly.
func (c *WizardConfig) BuildConfigRepoCloneEnvVars() map[string]string {
	if c == nil {
		return nil
	}
	switch c.ConfigRepoCloneAuth {
	case "pat":
		// When secure-exec is enabled, use the secret path instead of raw value
		if strings.TrimSpace(c.ConfigRepoPATPath) != "" {
			return map[string]string{
				"DIPPER_PASS_ENV": c.ConfigRepoPATPath,
			}
		}
		patVal := strings.TrimSpace(c.ConfigRepoPATValue)
		if strings.HasPrefix(patVal, "$") {
			// Check if this is an hd-lookup reference ($HD_*)
			if strings.HasPrefix(patVal, "$HD_") {
				// hd-lookup reference: pass through the value as-is
				// The daemon will resolve it at runtime
				envVarName := patVal[1:] // strip "$"
				return map[string]string{
					"HD_LOOKUP_PREFIX": envVarName,
					envVarName:          patVal,
				}
			}
			// Regular env var reference: $MY_PAT -> DIPPER_PASS_ENV=MY_PAT
			envVarName := patVal[1:]
			return map[string]string{
				"DIPPER_PASS_ENV": envVarName,
				envVarName:        fmt.Sprintf("${%s}", envVarName),
			}
		}
		// Raw PAT value: set DIPPER_PASS_ENV=DIPPER_GITHUB_PAT and DIPPER_GITHUB_PAT=<value>
		return map[string]string{
			"DIPPER_PASS_ENV":   "DIPPER_GITHUB_PAT",
			"DIPPER_GITHUB_PAT": patVal,
		}
	case "github_app":
		m := map[string]string{
			"GH_APP_TOKEN_SOURCE": "github",
			"GH_APP_ID":           c.ConfigRepoGHAppID,
			"GH_INSTALLATION_ID":  c.ConfigRepoGHInstallID,
		}
		if strings.TrimSpace(c.ConfigRepoGHAppKeyPath) != "" {
			m["GH_APP_KEY"] = c.ConfigRepoGHAppKeyPath
		} else {
			m["GH_APP_KEY"] = c.ConfigRepoGHAppKey
		}
		return m
	case "ssh":
		m := map[string]string{}
		if strings.TrimSpace(c.ConfigRepoSSHKeyPath) != "" {
			m["DIPPER_SSH_KEY"] = c.ConfigRepoSSHKeyPath
		} else if c.ConfigRepoSSHKey != "" {
			m["DIPPER_SSH_KEY"] = c.ConfigRepoSSHKey
		}
		if c.ConfigRepoSSHFile != "" {
			m["DIPPER_SSH_FILE"] = c.ConfigRepoSSHFile
		}
		if c.ConfigRepoSSHKeyPassEnv != "" {
			m["DIPPER_SSH_KEY_PASS_ENV"] = c.ConfigRepoSSHKeyPassEnv
		}
		if len(m) > 0 {
			return m
		}
		return nil
	default:
		return nil
	}
}

// BuildSecureExecEnvVars returns a map of environment variables for secure execution.
// When a secure-exec driver is selected, these vars configure the daemon's security policy.
// The output uses the exact env var names required by the daemon:
//   - HD_SECURE_LOADER: path to the secure-exec driver
//   - VAULT_ADDR: vault server address (for hd-driver-vault)
//   - VAULT_ROLE_ID: docker-secret-file://role_id (hardcoded for vault driver)
//   - VAULT_SECRET_ID: docker-secret-file://secret_id (hardcoded for vault driver)
//   - VAULT_TOKEN: vault token (optional, only when SecureExecVaultToken is set)
func (c *WizardConfig) BuildSecureExecEnvVars() map[string]string {
	if c == nil || c.SecureExecDriver == "" || c.SecureExecDriver == "none" {
		return nil
	}
	m := map[string]string{}
	switch c.SecureExecDriver {
	case "hd-driver-vault":
		m["HD_SECURE_LOADER"] = "./hd-driver-vault"
		if c.SecureExecVaultAddr != "" {
			m["VAULT_ADDR"] = c.SecureExecVaultAddr
		}
		m["VAULT_ROLE_ID"] = "docker-secret-file://role_id"
		m["VAULT_SECRET_ID"] = "docker-secret-file://secret_id"
		if strings.TrimSpace(c.SecureExecVaultToken) != "" {
			m["VAULT_TOKEN"] = c.SecureExecVaultToken
		}
	case "gcloud-secret":
		m["HD_SECURE_LOADER"] = "./gcloud-secret"
	}
	if len(m) == 0 {
		return nil
	}
	return m
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
