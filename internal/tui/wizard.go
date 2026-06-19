// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If a separate written commercial agreement exists, you may use this file under those terms instead.

// Package tui provides an interactive terminal wizard for the `hd init` command.
package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Charles546/hd-cli/internal/config"
)

// inputMode represents the interaction mode for the current step.
type inputMode int

const (
	modeNavigate inputMode = iota // Navigating between fields/steps
	modeTextInput                 // Actively typing in a text input
	modeRadioSelect               // Selecting a radio option
)

// WizardModel is the top-level Bubble Tea model that orchestrates the init wizard.
type WizardModel struct {
	config    *config.WizardConfig
	step      int
	total     int
	done      bool
	err       error

	// Interactive state
	mode        inputMode
	currentField int // index of the currently focused field within a step

	// Text input state
	textInput   textinput.Model
	textInputs  []textinput.Model // all text inputs for the current step

	// Radio selection state
	radioIndex  int // currently highlighted radio option index

	// Error display
	validationErr string
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
	return m.initStep(1)
}

// initStep prepares the model state for the given step.
func (m *WizardModel) initStep(s int) tea.Cmd {
	m.step = s
	m.currentField = 0
	m.validationErr = ""
	m.mode = modeNavigate
	m.textInputs = nil
	m.textInput = textinput.Model{}

	stepInfo := getStepInfo(s)
	if stepInfo == nil {
		return nil
	}

	switch stepInfo.stepType {
	case stepTypeRadio:
		m.mode = modeRadioSelect
		// Set radioIndex to match current config value
		m.radioIndex = m.findRadioIndex(stepInfo)
	case stepTypeMultiField:
		m.textInputs = make([]textinput.Model, len(stepInfo.fields))
		for i, field := range stepInfo.fields {
			ti := textinput.New()
			ti.Placeholder = field.placeholder
			ti.Width = 60
			ti.Prompt = ""
			// Pre-fill with existing value
			ti.SetValue(field.getValue(m.config))
			m.textInputs[i] = ti
		}
		// Focus first field
		m.mode = modeTextInput
		m.textInput = m.textInputs[0]
		m.textInput.Focus()
	case stepTypeSingleField:
		ti := textinput.New()
		ti.Placeholder = stepInfo.fields[0].placeholder
		ti.Width = 60
		ti.Prompt = ""
		ti.SetValue(stepInfo.fields[0].getValue(m.config))
		m.textInputs = []textinput.Model{ti}
		m.textInput = ti
		m.mode = modeTextInput
		m.textInput.Focus()
	}

	return nil
}

// findRadioIndex finds the index of the currently selected radio option.
func (m *WizardModel) findRadioIndex(info *stepInfo) int {
	val := info.radioGetter(m.config)
	for i, opt := range info.radioOptions {
		if opt == val {
			return i
		}
	}
	return 0
}

// Update implements tea.Model.
func (m *WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.mode == modeTextInput {
				// Let textinput handle q when typing
				break
			}
			return m, tea.Quit
		}
	}

	if m.done {
		return m.handleDone(msg)
	}

	return m.handleStepInput(msg)
}

// handleDone handles input on the summary/confirmation screen.
func (m *WizardModel) handleDone(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.err = m.generateConfig()
			return m, tea.Quit
		case "esc", "backspace":
			m.done = false
			m.step = stepCount - 1
			return m, m.initStep(m.step)
		}
	}
	return m, nil
}

// handleStepInput dispatches input handling based on the current step.
func (m *WizardModel) handleStepInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	stepInfo := getStepInfo(m.step)
	if stepInfo == nil {
		return m, nil
	}

	switch m.mode {
	case modeTextInput:
		return m.handleTextInput(msg, stepInfo)
	case modeRadioSelect:
		return m.handleRadioSelect(msg, stepInfo)
	case modeNavigate:
		return m.handleNavigate(msg, stepInfo)
	}

	return m, nil
}

// handleTextInput delegates to the active text input and handles field/step navigation.
func (m *WizardModel) handleTextInput(msg tea.Msg, stepInfo *stepInfo) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "tab":
			// Save current field value
			m.saveCurrentFieldValue(stepInfo)

			if m.currentField < len(m.textInputs)-1 {
				// Move to next field within the step
				m.textInput.Blur()
				m.textInputs[m.currentField] = m.textInput
				m.currentField++
				m.textInput = m.textInputs[m.currentField]
				m.textInput.Focus()
				m.textInputs[m.currentField] = m.textInput
				return m, nil
			}

			// Last field: validate and advance
			if err := m.validateCurrentStep(); err != nil {
				m.validationErr = err.Error()
				return m, nil
			}

			if m.step >= m.total {
				m.done = true
			} else {
				return m, m.initStep(m.step + 1)
			}
			return m, nil

		case "esc":
			if m.currentField > 0 {
				// Move back to previous field
				m.textInput.Blur()
				m.textInputs[m.currentField] = m.textInput
				m.currentField--
				m.textInput = m.textInputs[m.currentField]
				m.textInput.Focus()
				m.textInputs[m.currentField] = m.textInput
				return m, nil
			}
			// First field: go back to previous step
			if m.step > 1 {
				return m, m.initStep(m.step - 1)
			}
			return m, nil

		case "backspace":
			// If text input is empty, go back a step
			if m.textInput.Value() == "" {
				if m.currentField > 0 {
					m.textInput.Blur()
					m.textInputs[m.currentField] = m.textInput
					m.currentField--
					m.textInput = m.textInputs[m.currentField]
					m.textInput.Focus()
					m.textInputs[m.currentField] = m.textInput
					return m, nil
				}
				if m.step > 1 {
					return m, m.initStep(m.step - 1)
				}
				return m, nil
			}
			// Otherwise let textinput handle it
		}
	}

	// Delegate to text input
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	m.textInputs[m.currentField] = m.textInput
	return m, cmd
}

// handleRadioSelect handles arrow key navigation and selection for radio steps.
func (m *WizardModel) handleRadioSelect(msg tea.Msg, stepInfo *stepInfo) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.radioIndex > 0 {
				m.radioIndex--
			}
			return m, nil
		case "down", "j":
			if m.radioIndex < len(stepInfo.radioOptions)-1 {
				m.radioIndex++
			}
			return m, nil
		case "enter", "tab":
			// Commit the selection
			stepInfo.radioSetter(m.config, stepInfo.radioOptions[m.radioIndex])
			m.validationErr = ""
			if m.step >= m.total {
				m.done = true
			} else {
				return m, m.initStep(m.step + 1)
			}
			return m, nil
		case "esc", "backspace":
			if m.step > 1 {
				return m, m.initStep(m.step - 1)
			}
			return m, nil
		}
	}
	return m, nil
}

// handleNavigate handles navigation for steps without inputs (e.g. welcome).
func (m *WizardModel) handleNavigate(msg tea.Msg, stepInfo *stepInfo) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "tab":
			if m.step >= m.total {
				m.done = true
			} else {
				return m, m.initStep(m.step + 1)
			}
		case "esc", "backspace":
			if m.step > 1 {
				return m, m.initStep(m.step - 1)
			}
		}
	}
	return m, nil
}

// saveCurrentFieldValue saves the current text input's value to the config.
func (m *WizardModel) saveCurrentFieldValue(stepInfo *stepInfo) {
	if m.currentField < len(stepInfo.fields) {
		val := m.textInput.Value()
		stepInfo.fields[m.currentField].setValue(m.config, val)
	}
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
		b.WriteString(m.renderNavigation())
		return b.String()
	}

	// Render step content
	b.WriteString(m.renderStepContent())

	// Validation error
	if m.validationErr != "" {
		b.WriteString(ErrorStyle.Render("  ✗ " + m.validationErr))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.renderNavigation())
	return b.String()
}

// renderStepContent renders the interactive content for the current step.
func (m *WizardModel) renderStepContent() string {
	stepInfo := getStepInfo(m.step)
	if stepInfo == nil {
		return ""
	}

	var b strings.Builder

	switch m.step {
	case 1:
		b.WriteString(renderWelcome(m))
		return b.String()
	case 15:
		b.WriteString(renderStepTitle("Step 15: Summary & Confirm"))
		b.WriteString(m.renderSummary())
		b.WriteString(ConfirmStyle.Render("\n  Press Enter to confirm and generate configs, or q to quit."))
		return b.String()
	}

	b.WriteString(renderStepTitle(stepInfo.title))

	switch m.mode {
	case modeTextInput:
		for i, field := range stepInfo.fields {
			b.WriteString(LabelStyle.Render(field.label + ":"))
			b.WriteString("\n")

			if i < len(m.textInputs) {
				// Render the text input
				ti := m.textInputs[i]
				if i == m.currentField {
					// Active field: show with focused style
					b.WriteString(FocusedInputStyle.Render(ti.View()))
				} else {
					// Inactive field: show value or placeholder
					val := ti.Value()
					if val != "" {
						b.WriteString(InputStyle.Render(val))
					} else {
						b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(mutedColor)).Render(field.placeholder))
					}
				}
			}
			b.WriteString("\n")

			if field.help != "" {
				b.WriteString(HelpStyle.Render(field.help))
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}

	case modeRadioSelect:
		b.WriteString(LabelStyle.Render(stepInfo.radioLabel + ":"))
		b.WriteString("\n")
		for i, opt := range stepInfo.radioOptions {
			if i == m.radioIndex {
				b.WriteString(SelectedItemStyle.Render("  ● "))
				b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(accentColor)).Render(opt))
			} else {
				b.WriteString(UnselectedItemStyle.Render("  ○ "))
				b.WriteString(UnselectedItemStyle.Render(opt))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
		if stepInfo.fields != nil {
			// Render conditional fields after radio selection
			for i, field := range stepInfo.fields {
				if i < len(m.textInputs) {
					b.WriteString(LabelStyle.Render(field.label + ":"))
					b.WriteString("\n")
					ti := m.textInputs[i]
					b.WriteString(FocusedInputStyle.Render(ti.View()))
					b.WriteString("\n")
					if field.help != "" {
						b.WriteString(HelpStyle.Render(field.help))
						b.WriteString("\n")
					}
					b.WriteString("\n")
				}
			}
		}
	}

	return b.String()
}

func (m *WizardModel) renderNavigation() string {
	var hints []string
	if !m.done {
		switch m.mode {
		case modeTextInput:
			hints = append(hints, "type to enter text")
			if m.currentField < len(m.textInputs)-1 {
				hints = append(hints, "tab=next field")
			} else {
				hints = append(hints, "enter=next")
			}
			if m.currentField > 0 {
				hints = append(hints, "esc=prev field")
			} else {
				hints = append(hints, "esc=back")
			}
		case modeRadioSelect:
			hints = append(hints, "↑↓ select")
			hints = append(hints, "enter=confirm")
			if m.step > 1 {
				hints = append(hints, "esc=back")
			}
		case modeNavigate:
			hints = append(hints, "enter=next")
			if m.step > 1 {
				hints = append(hints, "esc=back")
			}
		}
		hints = append(hints, "q=quit")
	} else {
		hints = append(hints, "enter=confirm")
		hints = append(hints, "esc=back")
		hints = append(hints, "q=quit")
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
	items = append(items, fmt.Sprintf("  Secrets backend: %s", cfg.SecretsBackend))
	items = append(items, fmt.Sprintf("  Redis: %s", cfg.RedisMode))

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
	items = append(items, fmt.Sprintf("  GitHub repo creation: %v", cfg.GithubCreateRepo))

	return SummaryBoxStyle.Render(
		StepTitleStyle.Render("Configuration Summary") + "\n\n" + strings.Join(items, "\n"),
	)
}

// validateCurrentStep checks the current step's required fields.
func (m *WizardModel) validateCurrentStep() error {
	// Save current field first
	stepInfo := getStepInfo(m.step)
	if stepInfo != nil && m.mode == modeTextInput {
		m.saveCurrentFieldValue(stepInfo)
	}

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
		// Validate secrets backend selection
		if m.config.SecretsBackend == "" {
			return fmt.Errorf("secrets backend selection is required")
		}
	case 8:
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


func boolToRadio(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// radioToBool converts a radio string to a boolean.
func radioToBool(s string) bool {
	return s == "yes"
}

// parseInt parses a string to int, returning 0 on failure.
func parseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func renderStepTitle(title string) string {
	return StepTitleStyle.Render(title) + "\n\n"
}
