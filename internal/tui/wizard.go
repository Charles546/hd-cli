// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If a separate written commercial agreement exists, you may use this file under those terms instead.

// Package tui provides an interactive terminal wizard for the `hd init` command.
package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"

	"github.com/Charles546/hd-cli/internal/config"
)

// inputMode represents the interaction mode for the current step.
type inputMode int

const (
	modeNavigate        inputMode = iota // Navigating between fields/steps (welcome)
	modeTextInput                        // Actively typing in a text input
	modeRadioSelect                      // Selecting a radio option
	modeCheckboxSelect                   // Toggling checkboxes
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
	radioIndex    int    // currently highlighted radio option index
	radioSavedVal string // saved config value when entering a radio step, restored on Esc

	// Checkbox selection state
	checkboxIndex int // currently highlighted checkbox option index

	// Error display
	validationErr string

	// saveMsg displays a temporary confirmation after saving answers.
	saveMsg string

	// pendingDefault tracks whether the current text input holds a default value
	// that should be cleared on the next character or backspace keypress.
	pendingDefault bool

	// quit is set to true when the user quits via ctrl+q or ctrl+c.
	// RunWizard checks this to distinguish quit from completion.
	quit bool
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
	m.saveMsg = ""
	m.mode = modeNavigate
	m.textInputs = nil
	m.textInput = textinput.Model{}
	m.pendingDefault = false

	stepInfo := getStepInfo(s)
	if stepInfo == nil {
		return nil
	}

	switch stepInfo.stepType {
	case stepTypeRadio:
		m.mode = modeRadioSelect
		m.radioIndex = m.findRadioIndex(stepInfo)
		// Save the current config value so we can revert on Esc
		m.radioSavedVal = stepInfo.radioGetter(m.config)
		// Build text inputs for conditional fields based on current config
		m.buildRadioTextInputs(stepInfo)
	case stepTypeCheckbox:
		m.mode = modeCheckboxSelect
		m.checkboxIndex = 0
		// Build text inputs for conditional fields based on current config
		m.buildCheckboxTextInputs(stepInfo)
	case stepTypeMultiField:
		m.mode = modeTextInput
		m.buildMultiFieldInputs(stepInfo)
	case stepTypeSingleField:
		m.mode = modeTextInput
		// Fix 1: For step 3 (Config Directory), default to ./<project-name>
		placeholder := stepInfo.fields[0].placeholder
		defaultValue := stepInfo.fields[0].getValue(m.config)
		if s == 3 && m.config.ProjectName != "" {
			defaultValue = "./" + m.config.ProjectName
		}
		ti := textinput.New()
		ti.Placeholder = placeholder
		ti.Width = 60
		ti.Prompt = ""
		ti.SetValue(defaultValue)
		m.textInputs = []textinput.Model{ti}
		m.textInput = ti
		m.textInput.Focus()
		if defaultValue != "" {
			m.pendingDefault = true
		}
	}

	return nil
}

// buildMultiFieldInputs creates text inputs for a multi-field step,
// filtering out conditional fields whose conditions are not met.
func (m *WizardModel) buildMultiFieldInputs(stepInfo *stepInfo) {
	var visibleFields []fieldDescriptor
	for _, f := range stepInfo.fields {
		if f.condition == nil || f.condition(m.config) {
			visibleFields = append(visibleFields, f)
		}
	}
	m.textInputs = make([]textinput.Model, len(visibleFields))
	for i, field := range visibleFields {
		ti := textinput.New()
		ti.Placeholder = field.placeholder
		ti.Width = 60
		ti.Prompt = ""
		ti.SetValue(field.getValue(m.config))
		m.textInputs[i] = ti
	}
	if len(m.textInputs) > 0 {
		m.textInput = m.textInputs[0]
		m.textInput.Focus()
		if m.textInputs[0].Value() != "" {
			m.pendingDefault = true
		}
	}
}

// buildCheckboxTextInputs creates text inputs for conditional fields
// in a checkbox step, based on current config values.
func (m *WizardModel) buildCheckboxTextInputs(stepInfo *stepInfo) {
	var visibleFields []fieldDescriptor
	for _, f := range stepInfo.fields {
		if f.condition == nil || f.condition(m.config) {
			visibleFields = append(visibleFields, f)
		}
	}
	m.textInputs = make([]textinput.Model, len(visibleFields))
	for i, field := range visibleFields {
		ti := textinput.New()
		ti.Placeholder = field.placeholder
		ti.Width = 60
		ti.Prompt = ""
		ti.SetValue(field.getValue(m.config))
		m.textInputs[i] = ti
	}
}

// buildRadioTextInputs creates text inputs for conditional fields
// in a radio step, based on current config values.
func (m *WizardModel) buildRadioTextInputs(stepInfo *stepInfo) {
	var visibleFields []fieldDescriptor
	for _, f := range stepInfo.fields {
		if f.condition == nil || f.condition(m.config) {
			visibleFields = append(visibleFields, f)
		}
	}
	m.textInputs = make([]textinput.Model, len(visibleFields))
	for i, field := range visibleFields {
		ti := textinput.New()
		ti.Placeholder = field.placeholder
		ti.Width = 60
		ti.Prompt = ""
		ti.SetValue(field.getValue(m.config))
		m.textInputs[i] = ti
	}
	if len(m.textInputs) > 0 {
		m.textInput = m.textInputs[0]
	}
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
		case "ctrl+c":
			// Ctrl+C always quits, even in text input mode
			m.quit = true
			return m, tea.Quit
		case "ctrl+q":
			// Ctrl+Q always quits, even in text input mode
			m.quit = true
			return m, tea.Quit
		case "ctrl+s":
			// Save answers at any step (including text input mode)
			if !m.done {
				// Commit current text input value before saving,
				// otherwise the typed value is lost (still in textinput.Model).
				stepInfo := getStepInfo(m.step)
				if stepInfo != nil {
					if m.mode == modeTextInput {
						m.saveCurrentFieldValue(stepInfo)
					} else if m.mode == modeCheckboxSelect && m.currentField > 0 {
						m.saveCheckboxFieldValue(stepInfo)
					}
				}
				if err := m.saveAnswersFile(); err != nil {
					m.saveMsg = fmt.Sprintf("✗ Failed to save: %v", err)
				} else {
					m.saveMsg = fmt.Sprintf("✓ Answers saved to ./%s-answers.yaml", m.config.ProjectName)
				}
				return m, nil
			}
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
		case "s":
			// Refinement 2: Save answers to YAML file, then generate config (same as Enter)
			if err := m.saveAnswersFile(); err != nil {
				m.validationErr = fmt.Sprintf("failed to save answers: %v", err)
				return m, nil
			}
			m.err = m.generateConfig()
			return m, tea.Quit
		case "ctrl+q":
			// Quit without generating config
			m.quit = true
			return m, tea.Quit
		case "esc", "backspace":
			m.done = false
			m.step = prevStep(stepCount, m.config)
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
	case modeCheckboxSelect:
		return m.handleCheckboxSelect(msg, stepInfo)
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
				m.err = m.generateConfig()
				return m, tea.Quit
			}
			return m, m.initStep(nextStep(m.step, m.config))

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
				return m, m.initStep(prevStep(m.step, m.config))
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
					return m, m.initStep(prevStep(m.step, m.config))
				}
				return m, nil
			}
			// Otherwise let textinput handle it
		}
	}

	// Handle pendingDefault: clear the default value on first keystroke
	if m.pendingDefault {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.Type {
			case tea.KeyRunes:
				// Typing a character: clear the default value first
				m.textInput.SetValue("")
				m.pendingDefault = false
			case tea.KeyBackspace:
				// Backspace: clear the entire default value
				m.textInput.SetValue("")
				m.pendingDefault = false
			case tea.KeyLeft, tea.KeyRight:
				// Arrow keys: exit select-all mode, move cursor to end
				m.pendingDefault = false
				m.textInput.SetCursor(len(m.textInput.Value()))
			}
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
				// Tentatively apply the selection and rebuild text inputs
				stepInfo.radioSetter(m.config, stepInfo.radioOptions[m.radioIndex])
				m.buildRadioTextInputs(stepInfo)
			}
			return m, nil
		case "down", "j":
			if m.radioIndex < len(stepInfo.radioOptions)-1 {
				m.radioIndex++
				// Tentatively apply the selection and rebuild text inputs
				stepInfo.radioSetter(m.config, stepInfo.radioOptions[m.radioIndex])
				m.buildRadioTextInputs(stepInfo)
			}
			return m, nil
		case "enter", "tab":
			// Commit the selection
			stepInfo.radioSetter(m.config, stepInfo.radioOptions[m.radioIndex])
			// Rebuild text inputs to reflect the committed selection
			m.buildRadioTextInputs(stepInfo)
			m.validationErr = ""
			m.saveMsg = ""
			// If there are visible conditional fields, switch to text input mode
			if len(m.textInputs) > 0 {
				m.mode = modeTextInput
				m.currentField = 0
				m.textInput = m.textInputs[0]
				m.textInput.Focus()
				m.textInputs[0] = m.textInput
				return m, nil
			}
			// No conditional fields, advance to next step
			if m.step >= m.total {
				m.err = m.generateConfig()
				return m, tea.Quit
			}
			return m, m.initStep(nextStep(m.step, m.config))
		case "esc", "backspace":
			// Revert the radio config value to what it was when we entered this step
			stepInfo.radioSetter(m.config, m.radioSavedVal)
			if m.step > 1 {
				return m, m.initStep(prevStep(m.step, m.config))
			}
			return m, nil
		}
	}
	return m, nil
}

// handleCheckboxSelect handles checkbox toggle navigation and text input for conditional fields.
func (m *WizardModel) handleCheckboxSelect(msg tea.Msg, stepInfo *stepInfo) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.currentField == 0 {
				// Navigate within checkboxes
				if m.checkboxIndex > 0 {
					m.checkboxIndex--
				}
				return m, nil
			}
			// When in text fields, let the text input handle 'k' (fall through)
		case "down", "j":
			if m.currentField == 0 {
				if m.checkboxIndex < len(stepInfo.checkboxes)-1 {
					m.checkboxIndex++
				} else if len(m.textInputs) > 0 {
					// Move from last checkbox to first text field
					m.currentField = 1
					m.textInput = m.textInputs[0]
					m.textInput.Focus()
					m.textInputs[0] = m.textInput
				}
				return m, nil
			}
			// When in text fields, let the text input handle 'j' (fall through)
		case " ":
			// Toggle the current checkbox (only when on checkbox row)
			if m.currentField == 0 && m.checkboxIndex < len(stepInfo.checkboxes) {
				opt := stepInfo.checkboxes[m.checkboxIndex]
				current := opt.getValue(m.config)
				opt.setValue(m.config, !current)
				// Rebuild text inputs since conditions may have changed
				m.buildCheckboxTextInputs(stepInfo)
			}
			return m, nil
		case "enter":
			if m.currentField == 0 {
				// On checkboxes: toggle current, then move to next checkbox or text fields
				if m.checkboxIndex < len(stepInfo.checkboxes) {
					opt := stepInfo.checkboxes[m.checkboxIndex]
					current := opt.getValue(m.config)
					opt.setValue(m.config, !current)
					// Rebuild text inputs since conditions may have changed
					m.buildCheckboxTextInputs(stepInfo)
				}
				// Move to next checkbox or to text fields
				if m.checkboxIndex < len(stepInfo.checkboxes)-1 {
					m.checkboxIndex++
				} else if len(m.textInputs) > 0 {
					m.currentField = 1
					m.textInput = m.textInputs[0]
					m.textInput.Focus()
					m.textInputs[0] = m.textInput
				} else {
					// No text fields, advance to next step
					if err := m.validateCurrentStep(); err != nil {
						m.validationErr = err.Error()
						return m, nil
					}
					if m.step >= m.total {
						m.err = m.generateConfig()
						return m, tea.Quit
					}
					return m, m.initStep(nextStep(m.step, m.config))
				}
			} else {
				// In text fields: move to next or advance
				m.saveCheckboxFieldValue(stepInfo)
				if m.currentField < len(m.textInputs) {
					m.textInput.Blur()
					m.textInputs[m.currentField-1] = m.textInput
					m.currentField++
					m.textInput = m.textInputs[m.currentField-1]
					m.textInput.Focus()
					m.textInputs[m.currentField-1] = m.textInput
				} else {
					// Last text field: validate and advance
					if err := m.validateCurrentStep(); err != nil {
						m.validationErr = err.Error()
						return m, nil
					}
					if m.step >= m.total {
						m.err = m.generateConfig()
						return m, tea.Quit
					}
					return m, m.initStep(nextStep(m.step, m.config))
				}
			}
			return m, nil
		case "tab":
			if m.currentField == 0 {
				// On checkboxes: move to next checkbox or to text fields (no toggle)
				if m.checkboxIndex < len(stepInfo.checkboxes)-1 {
					m.checkboxIndex++
				} else if len(m.textInputs) > 0 {
					m.currentField = 1
					m.textInput = m.textInputs[0]
					m.textInput.Focus()
					m.textInputs[0] = m.textInput
				} else {
					// No text fields, advance to next step
					if err := m.validateCurrentStep(); err != nil {
						m.validationErr = err.Error()
						return m, nil
					}
					if m.step >= m.total {
						m.err = m.generateConfig()
						return m, tea.Quit
					}
					return m, m.initStep(nextStep(m.step, m.config))
				}
			} else {
				// In text fields: move to next field or advance step
				m.saveCheckboxFieldValue(stepInfo)
				if m.currentField < len(m.textInputs) {
					m.textInput.Blur()
					m.textInputs[m.currentField-1] = m.textInput
					m.currentField++
					m.textInput = m.textInputs[m.currentField-1]
					m.textInput.Focus()
					m.textInputs[m.currentField-1] = m.textInput
				} else {
					// Last text field: validate and advance
					if err := m.validateCurrentStep(); err != nil {
						m.validationErr = err.Error()
						return m, nil
					}
					if m.step >= m.total {
						m.err = m.generateConfig()
						return m, tea.Quit
					}
					return m, m.initStep(nextStep(m.step, m.config))
				}
			}
			return m, nil
		case "esc":
			if m.currentField > 0 {
				// In text fields: go back to previous field
				m.textInput.Blur()
				m.textInputs[m.currentField-1] = m.textInput
				m.currentField--
				if m.currentField > 0 {
					m.textInput = m.textInputs[m.currentField-1]
					m.textInput.Focus()
					m.textInputs[m.currentField-1] = m.textInput
				}
				return m, nil
			}
			// On checkboxes: go back a step
			if m.step > 1 {
				return m, m.initStep(prevStep(m.step, m.config))
			}
			return m, nil
		case "backspace":
			if m.currentField > 0 {
				// In text fields: only navigate back if text is empty
				if m.textInput.Value() == "" {
					m.textInput.Blur()
					m.textInputs[m.currentField-1] = m.textInput
					m.currentField--
					if m.currentField > 0 {
						m.textInput = m.textInputs[m.currentField-1]
						m.textInput.Focus()
						m.textInputs[m.currentField-1] = m.textInput
					}
					return m, nil
				}
				// Otherwise let textinput handle backspace (delete character)
				// Exit switch and delegate to text input below
				break
			}
			// On checkboxes: go back a step
			if m.step > 1 {
				return m, m.initStep(prevStep(m.step, m.config))
			}
			return m, nil
		}
	}

	// Delegate to text input when in text field mode
	if m.currentField > 0 {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		m.textInputs[m.currentField-1] = m.textInput
		return m, cmd
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
				// Fix 1: On the last step (summary), generate config directly
				// instead of going through an intermediate done/confirmation screen.
				m.err = m.generateConfig()
				return m, tea.Quit
			}
			return m, m.initStep(nextStep(m.step, m.config))
		case "s":
			// Fix 1: On the last step, 's' saves answers then generates directly.
			if m.step >= m.total {
				if err := m.saveAnswersFile(); err != nil {
					m.validationErr = fmt.Sprintf("failed to save answers: %v", err)
					return m, nil
				}
				m.err = m.generateConfig()
				return m, tea.Quit
			}
		case "esc", "backspace":
			if m.step > 1 {
				return m, m.initStep(prevStep(m.step, m.config))
			}
		}
	}
	return m, nil
}

// saveCurrentFieldValue saves the current text input's value to the config.
func (m *WizardModel) saveCurrentFieldValue(stepInfo *stepInfo) {
	// For multi-field and radio steps, we need to map the visible field index back to the original field
	if stepInfo.stepType == stepTypeMultiField || stepInfo.stepType == stepTypeRadio {
		visibleIdx := 0
		for _, f := range stepInfo.fields {
			if f.condition == nil || f.condition(m.config) {
				if visibleIdx == m.currentField {
					val := m.textInput.Value()
					f.setValue(m.config, val)
					return
				}
				visibleIdx++
			}
		}
	} else if m.currentField < len(stepInfo.fields) {
		val := m.textInput.Value()
		stepInfo.fields[m.currentField].setValue(m.config, val)
	}
}

// saveCheckboxFieldValue saves the current text input's value to the config
// for checkbox step conditional fields.
func (m *WizardModel) saveCheckboxFieldValue(stepInfo *stepInfo) {
	if m.currentField < 1 || m.currentField > len(m.textInputs) {
		return
	}
	// Find the corresponding field descriptor
	fieldIdx := m.currentField - 1
	visibleIdx := 0
	for _, f := range stepInfo.fields {
		if f.condition == nil || f.condition(m.config) {
			if visibleIdx == fieldIdx {
				val := m.textInput.Value()
				f.setValue(m.config, val)
				return
			}
			visibleIdx++
		}
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
		// Refinement 1: When done, show the confirmation prompt (summary already shown in step 15)
		b.WriteString(ConfirmStyle.Render("  Press Enter to confirm and generate configs."))
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render("  s=save answers & generate • esc=back • ctrl+q=quit"))
		b.WriteString("\n")
		return b.String()
	}

	// Render step content
	b.WriteString(m.renderStepContent())

	// Validation error
	if m.validationErr != "" {
		b.WriteString(ErrorStyle.Render("  ✗ " + m.validationErr))
		b.WriteString("\n")
	}

	// Save confirmation message
	if m.saveMsg != "" {
		b.WriteString(SuccessStyle.Render("  " + m.saveMsg))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.renderNavigation())
	return b.String()
}


// renderRadioSelection renders the radio options with the selected marker.
// Used by both modeRadioSelect and modeTextInput (for radio steps with conditional fields).
func (m *WizardModel) renderRadioSelection(stepInfo *stepInfo) string {
	var b strings.Builder
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
		// Refinement 1: Show the summary directly when entering step 15
		b.WriteString(renderStepTitle("Step 15: Summary & Confirm"))
		b.WriteString(m.renderSummary())
		b.WriteString("\n")
		b.WriteString(ConfirmStyle.Render("  Press Enter to generate configs."))
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render("  s=save answers & generate • esc=back • ctrl+q=quit"))
		return b.String()
	}

	b.WriteString(renderStepTitle(stepInfo.title))

	switch m.mode {
	case modeTextInput:
		// For radio steps that have switched to text input mode, render the
		// radio selection above the conditional text fields so the user can
		// still see which option they selected.
		if stepInfo.stepType == stepTypeRadio {
			b.WriteString(m.renderRadioSelection(stepInfo))
			b.WriteString("\n")
			// Render visible conditional text fields
			visibleIdx := 0
			for _, field := range stepInfo.fields {
				if field.condition != nil && !field.condition(m.config) {
					continue
				}
				if visibleIdx < len(m.textInputs) {
					b.WriteString(LabelStyle.Render(field.label + ":"))
					b.WriteString("\n")
					ti := m.textInputs[visibleIdx]
					if visibleIdx == m.currentField {
						if m.pendingDefault {
							b.WriteString(SelectAllStyle.Render(ti.View()))
						} else {
							b.WriteString(FocusedInputStyle.Render(ti.View()))
						}
					} else {
						b.WriteString(InputStyle.Render(ti.View()))
					}
					b.WriteString("\n")
					if field.help != "" {
						b.WriteString(HelpStyle.Render(field.help))
						b.WriteString("\n")
					}
					b.WriteString("\n")
				}
				visibleIdx++
			}
		} else if stepInfo.stepType == stepTypeMultiField {
			// For multi-field steps, render only visible fields
			visibleIdx := 0
			for _, field := range stepInfo.fields {
				if field.condition != nil && !field.condition(m.config) {
					continue
				}
				// Determine the label, placeholder, and help for this field
				fieldLabel := field.label
				fieldPlaceholder := field.placeholder
				fieldHelp := field.help
				// Override labels for Slack secret fields based on secrets backend
				if m.step == 9 {
					secretKey := slackSecretFieldKey(field.label)
					if secretKey != "" {
						fieldLabel, fieldPlaceholder, fieldHelp = slackSecretLabels(m.config, secretKey)
					}
				}
				b.WriteString(LabelStyle.Render(fieldLabel + ":"))
				b.WriteString("\n")
				if visibleIdx < len(m.textInputs) {
					ti := m.textInputs[visibleIdx]
					if visibleIdx == m.currentField {
						if m.pendingDefault {
							b.WriteString(SelectAllStyle.Render(ti.View()))
						} else {
							b.WriteString(FocusedInputStyle.Render(ti.View()))
						}
					} else {
						val := ti.Value()
						if val != "" {
							b.WriteString(InputStyle.Render(val))
						} else {
							b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(mutedColor)).Render(fieldPlaceholder))
						}
					}
				}
				b.WriteString("\n")
				if fieldHelp != "" {
					b.WriteString(HelpStyle.Render(fieldHelp))
					b.WriteString("\n")
				}
				b.WriteString("\n")
				visibleIdx++
			}
		} else {
			for i, field := range stepInfo.fields {
				b.WriteString(LabelStyle.Render(field.label + ":"))
				b.WriteString("\n")
				if i < len(m.textInputs) {
					ti := m.textInputs[i]
					if i == m.currentField {
						if m.pendingDefault {
							b.WriteString(SelectAllStyle.Render(ti.View()))
						} else {
							b.WriteString(FocusedInputStyle.Render(ti.View()))
						}
					} else {
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
		}

	case modeRadioSelect:
		b.WriteString(m.renderRadioSelection(stepInfo))
		b.WriteString("\n")
		// Render conditional text fields that match current conditions
		visibleIdx := 0
		for _, field := range stepInfo.fields {
			if field.condition != nil && !field.condition(m.config) {
				continue
			}
			if visibleIdx < len(m.textInputs) {
				b.WriteString(LabelStyle.Render(field.label + ":"))
				b.WriteString("\n")
				ti := m.textInputs[visibleIdx]
				b.WriteString(FocusedInputStyle.Render(ti.View()))
				b.WriteString("\n")
				if field.help != "" {
					b.WriteString(HelpStyle.Render(field.help))
					b.WriteString("\n")
				}
				b.WriteString("\n")
			}
			visibleIdx++
		}

	case modeCheckboxSelect:
		b.WriteString(LabelStyle.Render(stepInfo.checkboxLabel + ":"))
		b.WriteString("\n")
		for i, cb := range stepInfo.checkboxes {
			checked := cb.getValue(m.config)
			checkboxChar := "☐"
			if checked {
				checkboxChar = "☑"
			}
			if i == m.checkboxIndex && m.currentField == 0 {
				b.WriteString(SelectedItemStyle.Render("  " + checkboxChar + " "))
				b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(accentColor)).Render(cb.label))
			} else {
				b.WriteString(UnselectedItemStyle.Render("  " + checkboxChar + " "))
				b.WriteString(UnselectedItemStyle.Render(cb.label))
			}
			b.WriteString("\n")
			if cb.help != "" {
				b.WriteString(HelpStyle.Render("      " + cb.help))
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
		// Render conditional text fields
		if len(m.textInputs) > 0 {
			b.WriteString(LabelStyle.Render("Secret paths:"))
			b.WriteString("\n")
			// Map visible fields to labels
			visibleIdx := 0
			for _, field := range stepInfo.fields {
				if field.condition != nil && !field.condition(m.config) {
					continue
				}
				// Determine the label, placeholder, and help for this field
				fieldLabel := field.label
				fieldPlaceholder := field.placeholder
				fieldHelp := field.help
				// Override labels for GitHub secret fields based on secrets backend
				if m.step == 8 {
					secretKey := ghSecretFieldKey(field.label)
					if secretKey != "" {
						fieldLabel, fieldPlaceholder, fieldHelp = ghSecretLabels(m.config, secretKey)
					}
				}
				b.WriteString(LabelStyle.Render("  " + fieldLabel + ":"))
				b.WriteString("\n")
				if visibleIdx < len(m.textInputs) {
					ti := m.textInputs[visibleIdx]
					fieldInputIdx := visibleIdx + 1 // +1 because 0 is checkboxes
					if fieldInputIdx == m.currentField {
						b.WriteString(FocusedInputStyle.Render(ti.View()))
					} else {
						val := ti.Value()
						if val != "" {
							b.WriteString(InputStyle.Render(val))
						} else {
							b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(mutedColor)).Render(fieldPlaceholder))
						}
					}
				}
				b.WriteString("\n")
				if fieldHelp != "" {
					b.WriteString(HelpStyle.Render("  " + fieldHelp))
					b.WriteString("\n")
				}
				b.WriteString("\n")
				visibleIdx++
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
		case modeCheckboxSelect:
			hints = append(hints, "↑↓ navigate")
			hints = append(hints, "space=toggle")
			hints = append(hints, "enter=next")
			if m.step > 1 {
				hints = append(hints, "esc=back")
			}
		case modeNavigate:
			hints = append(hints, "enter=next")
			if m.step > 1 {
				hints = append(hints, "esc=back")
			}
		}
		hints = append(hints, "ctrl+s=save")
		hints = append(hints, "ctrl+q=quit")
	} else {
		hints = append(hints, "enter=confirm")
		hints = append(hints, "esc=back")
		hints = append(hints, "ctrl+q=quit")
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
	items = append(items, fmt.Sprintf("  Clone credentials: %s", cfg.EssentialsCloneType))
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

	// Show both integration types
	var ghTypes []string
	if cfg.HasGitHubAppIntegration {
		ghTypes = append(ghTypes, "github_app")
	}
	if cfg.HasGithubPATIntegration {
		ghTypes = append(ghTypes, "pat")
	}
	ghIntegration := "none"
	if len(ghTypes) > 0 {
		ghIntegration = strings.Join(ghTypes, "+")
	}
	items = append(items, fmt.Sprintf("  GitHub integration: %s", ghIntegration))
	items = append(items, "  Slack integration: enabled")
	// Show secret format hint based on secrets backend
	if isDevMode(m.config) {
		items = append(items, "  Secret format: plain values / $HD_* refs")
	} else {
		items = append(items, "  Secret format: LOOKUP[vault,...] paths")
	}
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
		if m.config.SecretsBackend == "" {
			return fmt.Errorf("secrets backend selection is required")
		}
	case 8:
		if m.config.HasGitHubAppIntegration {
			if strings.TrimSpace(m.config.GithubAppID) == "" {
				return fmt.Errorf("github App ID is required when GitHub App integration is enabled")
			}
			if strings.TrimSpace(m.config.GithubInstallationID) == "" {
				return fmt.Errorf("GitHub Installation ID is required when GitHub App integration is enabled")
			}
			if strings.TrimSpace(m.config.GithubKeyPath) == "" {
				return fmt.Errorf("private key value is required when GitHub App integration is enabled")
			}
		}
		if m.config.HasGithubPATIntegration {
			if strings.TrimSpace(m.config.GithubTokenPath) == "" {
				return fmt.Errorf("token value is required when PAT integration is enabled")
			}
		}
		// Validate dev mode secret fields: reject non-HD_ env var references
		if isDevMode(m.config) {
			if err := validateDevSecretValue(m.config.GithubKeyPath); err != "" {
				return fmt.Errorf("private key: %s", err)
			}
			if err := validateDevSecretValue(m.config.GithubTokenPath); err != "" {
				return fmt.Errorf("token: %s", err)
			}
			if err := validateDevSecretValue(m.config.GithubWebhookSecret); err != "" {
				return fmt.Errorf("webhook secret: %s", err)
			}
		}
	case 9:
		// Validate dev mode secret fields for Slack
		if isDevMode(m.config) {
			if err := validateDevSecretValue(m.config.SlackBotTokenPath); err != "" {
				return fmt.Errorf("bot token: %s", err)
			}
			if err := validateDevSecretValue(m.config.SlackSigningSecretPath); err != "" {
				return fmt.Errorf("signing secret: %s", err)
			}
			if err := validateDevSecretValue(m.config.SlackInteractionToken); err != "" {
				return fmt.Errorf("interaction token: %s", err)
			}
			if err := validateDevSecretValue(m.config.SlackSlashCommandToken); err != "" {
				return fmt.Errorf("slash command token: %s", err)
			}
		}
	case 11:
		if m.config.DeploymentMode != "docker" {
			return nil
		}
		if strings.TrimSpace(m.config.DockerImageTag) == "" {
			return fmt.Errorf("docker image tag is required")
		}
	case 12:
		if m.config.DeploymentMode != "kubernetes" {
			return nil
		}
		if strings.TrimSpace(m.config.K8sNamespace) == "" {
			return fmt.Errorf("kubernetes namespace is required")
		}
		if strings.TrimSpace(m.config.K8sRepoStrategy) == "" {
			return fmt.Errorf("config repo strategy is required")
		}
	case 13:
		if m.config.DeploymentMode != "source" {
			return nil
		}
		if strings.TrimSpace(m.config.SourceClonePath) == "" {
			return fmt.Errorf("clone path is required")
		}
		if strings.TrimSpace(m.config.SourceBranch) == "" {
			return fmt.Errorf("source branch is required")
		}
	case 14:
		if m.config.GithubCreateRepo {
			if strings.TrimSpace(m.config.GitRemoteURL) == "" {
				return fmt.Errorf("git remote URL is required when creating a GitHub repo")
			}
		}
	}
	return nil
}

// generateConfig invokes the config generator with the wizard's config.
func (m *WizardModel) generateConfig() error {
	// Compute absolute path for ConfigDir if not already set
	if m.config.ConfigDirAbs == "" {
		if abs, err := filepath.Abs(m.config.ConfigDir); err == nil {
			m.config.ConfigDirAbs = abs
		} else {
			m.config.ConfigDirAbs = m.config.ConfigDir
		}
	}
	generator := config.NewGenerator()
	return generator.Generate(m.config, m.config.ConfigDir, false)
}

// saveAnswersFile saves the wizard answers to a YAML file for reuse with --config flag.
func (m *WizardModel) saveAnswersFile() error {
	answersPath := "./" + m.config.ProjectName + "-answers.yaml"
	data, err := yaml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("failed to marshal answers: %w", err)
	}
	if err := os.WriteFile(answersPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write answers file: %w", err)
	}
	return nil
}

// RunWizard is the entry point that starts the interactive wizard.
// It takes over the terminal and returns the completed WizardConfig or an error.
// Returns nil, nil if the user quit without completing.
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
	// If the user quit (via q or ctrl+c), return nil config
	if resultModel.quit {
		return nil, nil
	}
	return resultModel.config, nil
}

// RunWizardWithConfig starts the interactive wizard with a pre-populated config.
// Used by hd init --config <file> (without --non-interactive) to let the user
// review and modify the loaded answers before generating.
// Returns nil, nil if the user quit without completing.
func RunWizardWithConfig(cfg *config.WizardConfig) (*config.WizardConfig, error) {
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
	// If the user quit (via q or ctrl+c), return nil config
	if resultModel.quit {
		return nil, nil
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
