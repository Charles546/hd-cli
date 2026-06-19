// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If a separate written commercial agreement exists, you may use this file under those terms instead.

// Package tui provides an interactive terminal wizard for the `hd init` command.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Charles546/hd-cli/internal/config"
)

// WizardModel is the top-level Bubble Tea model that orchestrates the init wizard.
type WizardModel struct {
	config *config.WizardConfig
	step   int
	total  int
	done   bool
	err    error
}

// stepCount is the total number of wizard steps (used for progress).
const stepCount = 15

// NewWizard creates a new wizard model starting at step 1.
func NewWizard(cfg *config.WizardConfig) *WizardModel {
	if cfg == nil {
		cfg = config.NewDefaultWizardConfig()
	}
	return &WizardModel{
		config: cfg,
		step:   1,
		total:  stepCount,
	}
}

// Init implements tea.Model.
func (m *WizardModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m *WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	if m.done {
		return m, nil
	}

	return m.handleStepInput(msg)
}

// View renders the current step of the wizard.
func (m *WizardModel) View() string {
	var b strings.Builder

	// Header
	b.WriteString(TitleStyle.Render("hd init — Honeydipper Project Wizard"))
	b.WriteString("\n")
	b.WriteString(ProgressBar(m.step, m.total))
	b.WriteString("\n\n")

	if m.done {
		b.WriteString(m.renderSummary())
		b.WriteString(ConfirmStyle.Render("\n  Press Enter to confirm and generate configs, or q to quit."))
		b.WriteString("\n")
		return b.String()
	}

	switch m.step {
	case 1:
		b.WriteString(renderWelcome(m))
	case 2:
		b.WriteString(renderStepTitle("Step 2: Project Name"))
		b.WriteString(renderInput("Project name", m.config.ProjectName, "Name for your config project"))
	case 3:
		b.WriteString(renderStepTitle("Step 3: Config Directory"))
		b.WriteString(renderInput("Config directory", m.config.ConfigDir, "Path where config files will be written"))
	case 4:
		b.WriteString(renderStepTitle("Step 4: Deployment Mode"))
		b.WriteString(renderRadio("Deployment mode", []string{"docker", "source", "kubernetes"}, m.config.DeploymentMode))
	case 5:
		b.WriteString(renderStepTitle("Step 5: Config Repo Setup"))
		b.WriteString(renderInput("Essentials repo URL", m.config.EssentialsRepoURL, "Honeydipper essentials config repo"))
		b.WriteString(renderInput("Branch", m.config.EssentialsBranch, "Branch to use"))
	case 6:
		b.WriteString(renderStepTitle("Step 6: GitHub Integration"))
		b.WriteString(renderRadio("Integration type", []string{"github_app", "pat"}, m.config.GithubIntegrationType))
		if m.config.GithubIntegrationType == "github_app" {
			b.WriteString(renderInput("App ID", m.config.GithubAppID, "GitHub App ID"))
			b.WriteString(renderInput("Installation ID", m.config.GithubInstallationID, "GitHub App Installation ID"))
			b.WriteString(renderInput("Private key secret path", m.config.GithubKeyPath, "Path to secret containing the private key"))
		} else {
			b.WriteString(renderInput("Token secret path", m.config.GithubTokenPath, "Path to secret containing the PAT"))
		}
		b.WriteString(renderInput("Webhook secret path", m.config.GithubWebhookSecret, "Path to webhook signing secret"))
	case 7:
		b.WriteString(renderStepTitle("Step 7: Slack Integration"))
		b.WriteString(renderInput("Bot token secret path", m.config.SlackBotTokenPath, "Path to Slack bot token"))
		b.WriteString(renderInput("Signing secret path", m.config.SlackSigningSecretPath, "Path to Slack signing secret"))
		b.WriteString(HelpStyle.Render("(Optional)"))
		b.WriteString(renderInput("Interaction token path", m.config.SlackInteractionTokenPath, "Slack interaction token (optional)"))
		b.WriteString(renderInput("Slash command token path", m.config.SlackSlashCommandTokenPath, "Slack slash command token (optional)"))
	case 8:
		b.WriteString(renderStepTitle("Step 8: AI Agent"))
		b.WriteString(renderRadio("Enable AI agent", []string{"yes", "no"}, boolToRadio(m.config.AIEnabled)))
		if m.config.AIEnabled {
			b.WriteString(renderInput("API key secret path", m.config.AIAPIKeyPath, "Path to OpenAI-compatible API key"))
			b.WriteString(renderInput("Base URL", m.config.AIBaseURL, "API base URL"))
			b.WriteString(renderInput("Model", m.config.AIModel, "LLM model name"))
			b.WriteString(renderInput("Engine name", m.config.AIEngineName, "Honeydipper engine name"))
		}
	case 9:
		b.WriteString(renderStepTitle("Step 9: Secrets Backend"))
		b.WriteString(renderRadio("Secrets backend", []string{"vault", "dev"}, m.config.SecretsBackend))
		if m.config.SecretsBackend == "vault" {
			b.WriteString(renderInput("Vault address", m.config.VaultAddress, "Vault server URL"))
			b.WriteString(renderInput("Auth method", m.config.VaultAuthMethod, "Vault auth method (e.g., token, approle)"))
		}
	case 10:
		b.WriteString(renderStepTitle("Step 10: Redis"))
		b.WriteString(renderRadio("Redis mode", []string{"local", "external"}, m.config.RedisMode))
		if m.config.RedisMode == "external" {
			b.WriteString(renderInput("Connection string", m.config.RedisConnString, "Redis connection string"))
		}
	case 11:
		b.WriteString(renderStepTitle("Step 11: Docker Configuration"))
		b.WriteString(renderInput("Image tag", m.config.DockerImageTag, "Honeydipper Docker image tag"))
		b.WriteString(renderRadio("Enable web UI", []string{"yes", "no"}, boolToRadio(m.config.DockerEnableUI)))
		b.WriteString(renderInput("API port", fmt.Sprintf("%d", m.config.DockerAPIPort), "API server port"))
		b.WriteString(renderInput("Webhook port", fmt.Sprintf("%d", m.config.DockerWebhookPort), "Webhook server port"))
	case 12:
		b.WriteString(renderStepTitle("Step 12: Kubernetes Configuration"))
		b.WriteString(renderInput("Namespace", m.config.K8sNamespace, "Kubernetes namespace"))
		b.WriteString(renderRadio("Config repo strategy", []string{"clone", "configmap", "git-sidecar"}, m.config.K8sRepoStrategy))
	case 13:
		b.WriteString(renderStepTitle("Step 13: Source Configuration"))
		b.WriteString(renderInput("Clone path", m.config.SourceClonePath, "Path where source will be cloned"))
		b.WriteString(renderInput("Branch", m.config.SourceBranch, "Honeydipper source branch"))
	case 14:
		b.WriteString(renderStepTitle("Step 14: GitHub Repo Creation"))
		b.WriteString(renderRadio("Create GitHub repo", []string{"yes", "no"}, boolToRadio(m.config.GithubCreateRepo)))
		if m.config.GithubCreateRepo {
			b.WriteString(renderInput("Repo name", m.config.GithubRepoName, "Repository name"))
			b.WriteString(renderRadio("Visibility", []string{"private", "public"}, m.config.GithubRepoVis))
		}
	case 15:
		b.WriteString(renderStepTitle("Step 15: Summary & Confirm"))
		b.WriteString(m.renderSummary())
		b.WriteString(ConfirmStyle.Render("\n  Press Enter to confirm and generate configs, or q to quit."))
	}

	b.WriteString("\n\n")
	b.WriteString(m.renderNavigation())
	return b.String()
}

func (m *WizardModel) renderNavigation() string {
	var hints []string
	if !m.done {
		hints = append(hints, "↑↓ navigate  enter=next  esc=back")
		if m.step > 1 {
			hints = append(hints, "backspace=prev step")
		}
	} else {
		hints = append(hints, "enter=confirm  q=quit")
	}
	return HelpStyle.Render(strings.Join(hints, "  "))
}

func (m *WizardModel) renderSummary() string {
	cfg := m.config
	var items []string

	items = append(items, fmt.Sprintf("  Project: %s", cfg.ProjectName))
	items = append(items, fmt.Sprintf("  Config dir: %s", cfg.ConfigDir))
	items = append(items, fmt.Sprintf("  Deployment: %s", cfg.DeploymentMode))
	items = append(items, fmt.Sprintf("  Essentials repo: %s (%s)", cfg.EssentialsRepoURL, cfg.EssentialsBranch))

	switch cfg.DeploymentMode {
	case "docker":
		items = append(items, fmt.Sprintf("  Docker image: %s", cfg.DockerImageTag))
		items = append(items, fmt.Sprintf("  UI enabled: %v", cfg.DockerEnableUI))
		items = append(items, fmt.Sprintf("  API port: %d", cfg.DockerAPIPort))
		items = append(items, fmt.Sprintf("  Webhook port: %d", cfg.DockerWebhookPort))
	case "kubernetes":
		items = append(items, fmt.Sprintf("  K8s namespace: %s", cfg.K8sNamespace))
		items = append(items, fmt.Sprintf("  Repo strategy: %s", cfg.K8sRepoStrategy))
	default:
		items = append(items, fmt.Sprintf("  Clone path: %s", cfg.SourceClonePath))
		items = append(items, fmt.Sprintf("  Branch: %s", cfg.SourceBranch))
	}

	items = append(items, fmt.Sprintf("  GitHub integration: %s", cfg.GithubIntegrationType))
	items = append(items, "  Slack integration: enabled")
	if cfg.AIEnabled {
		items = append(items, fmt.Sprintf("  AI agent: enabled (%s, %s)", cfg.AIModel, cfg.AIBaseURL))
	} else {
		items = append(items, "  AI agent: disabled")
	}
	items = append(items, fmt.Sprintf("  Secrets backend: %s", cfg.SecretsBackend))
	items = append(items, fmt.Sprintf("  Redis: %s", cfg.RedisMode))
	items = append(items, fmt.Sprintf("  GitHub repo creation: %v", cfg.GithubCreateRepo))

	return SummaryBoxStyle.Render(
		StepTitleStyle.Render("Configuration Summary") + "\n\n" + strings.Join(items, "\n"),
	)
}

// handleStepInput dispatches input handling based on the current step.
func (m *WizardModel) handleStepInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" {
				m.err = m.generateConfig()
				return m, tea.Quit
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if err := m.validateCurrentStep(); err != nil {
				m.err = err
				return m, nil
			}
			if m.step >= m.total {
				m.done = true
			} else {
				m.step++
			}
		case "esc":
			if m.step > 1 {
				m.step--
			}
		case "backspace":
			if m.step > 1 {
				m.step--
			}
		}
	}
	return m, nil
}

// validateCurrentStep checks the current step's required fields.
func (m *WizardModel) validateCurrentStep() error {
	switch m.step {
	case 2:
		if strings.TrimSpace(m.config.ProjectName) == "" {
			return fmt.Errorf("project name is required")
		}
	case 3:
		if strings.TrimSpace(m.config.ConfigDir) == "" {
			return fmt.Errorf("config directory is required")
		}
	case 6:
		if m.config.GithubIntegrationType == "github_app" {
			if strings.TrimSpace(m.config.GithubAppID) == "" {
				return fmt.Errorf("github App ID is required")
			}
			if strings.TrimSpace(m.config.GithubInstallationID) == "" {
				return fmt.Errorf("github Installation ID is required")
			}
			if strings.TrimSpace(m.config.GithubKeyPath) == "" {
				return fmt.Errorf("private key secret path is required")
			}
		} else {
			if strings.TrimSpace(m.config.GithubTokenPath) == "" {
				return fmt.Errorf("token secret path is required")
			}
		}
	case 11:
		if m.config.DeploymentMode != "docker" {
			return nil
		}
		if strings.TrimSpace(m.config.DockerImageTag) == "" {
			return fmt.Errorf("docker image tag is required")
		}
	}
	return nil
}

// generateConfig invokes the config generator with the wizard's config.
func (m *WizardModel) generateConfig() error {
	generator := config.NewGenerator()
	return generator.Generate(m.config, m.config.ConfigDir, false)
}

// RunWizard is the entry point that starts the interactive wizard.
// It takes over the terminal and returns the completed WizardConfig or an error.
func RunWizard() (*config.WizardConfig, error) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return nil, err
	}
	resultModel := result.(*WizardModel)
	if resultModel.err != nil {
		return nil, resultModel.err
	}
	return resultModel.config, nil
}

// Helper functions for rendering
func renderStepTitle(title string) string {
	return StepTitleStyle.Render(title) + "\n\n"
}

func renderInput(label, value, help string) string {
	var b strings.Builder
	b.WriteString(LabelStyle.Render(label + ":"))
	b.WriteString("\n")
	if value != "" {
		b.WriteString(InputStyle.Render(value))
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(mutedColor)).Render("(default: see help)"))
	}
	b.WriteString("\n")
	if help != "" {
		b.WriteString(HelpStyle.Render(help))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	return b.String()
}

func renderRadio(label string, options []string, selected string) string {
	var b strings.Builder
	b.WriteString(LabelStyle.Render(label + ":"))
	b.WriteString("\n")
	for _, opt := range options {
		if opt == selected {
			b.WriteString(SelectedItemStyle.Render("  ● "))
		} else {
			b.WriteString(UnselectedItemStyle.Render("  ○ "))
		}
		b.WriteString(opt)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	return b.String()
}

func boolToRadio(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
