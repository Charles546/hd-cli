// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If a separate written commercial agreement exists, you may use this file under those terms instead.

// Package tui provides an interactive terminal wizard for the `hd init` command.
package tui

import (
	"fmt"
	"path/filepath"
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
	radioIndex  int // currently highlighted radio option index

	// Checkbox selection state
	checkboxIndex int // currently highlighted checkbox option index

	// Error display
	validationErr string

	// pendingDefault tracks whether the current text input holds a default value
	// that should be cleared on the next character or backspace keypress.
	pendingDefault bool
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
	m.pendingDefault = false

	stepInfo := getStepInfo(s)
	if stepInfo == nil {
		return nil
	}

	switch stepInfo.stepType {
	case stepTypeRadio:
		m.mode = modeRadioSelect
		m.radioIndex = m.findRadioIndex(stepInfo)
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
			} else {
				// Navigate within text fields
				m.textInput.Blur()
				m.textInputs[m.currentField-1] = m.textInput
				m.currentField--
				if m.currentField > 0 {
					m.textInput = m.textInputs[m.currentField-1]
					m.textInput.Focus()
					m.textInputs[m.currentField-1] = m.textInput
				}
			}
			return m, nil
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
			} else if m.currentField < len(m.textInputs) {
				m.textInput.Blur()
				m.textInputs[m.currentField-1] = m.textInput
				m.currentField++
				if m.currentField <= len(m.textInputs) {
					m.textInput = m.textInputs[m.currentField-1]
					m.textInput.Focus()
					m.textInputs[m.currentField-1] = m.textInput
				}
			}
			return m, nil
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
						m.done = true
					} else {
						return m, m.initStep(m.step + 1)
					}
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
						m.done = true
					} else {
						return m, m.initStep(m.step + 1)
					}
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
						m.done = true
					} else {
						return m, m.initStep(m.step + 1)
					}
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
						m.done = true
					} else {
						return m, m.initStep(m.step + 1)
					}
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
				return m, m.initStep(m.step - 1)
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
				return m, m.initStep(m.step - 1)
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
	// For multi-field steps, we need to map the visible field index back to the original field
	if stepInfo.stepType == stepTypeMultiField {
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
		// For multi-field steps, render only visible fields
		if stepInfo.stepType == stepTypeMultiField {
			visibleIdx := 0
			for _, field := range stepInfo.fields {
				if field.condition != nil && !field.condition(m.config) {
					continue
				}
				b.WriteString(LabelStyle.Render(field.label + ":"))
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
				b.WriteString(LabelStyle.Render("  " + field.label + ":"))
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
							b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(mutedColor)).Render(field.placeholder))
						}
					}
				}
				b.WriteString("\n")
				if field.help != "" {
					b.WriteString(HelpStyle.Render("  " + field.help))
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
				return fmt.Errorf("github Installation ID is required when GitHub App integration is enabled")
			}
			if strings.TrimSpace(m.config.GithubKeyPath) == "" {
				return fmt.Errorf("private key secret path is required when GitHub App integration is enabled")
			}
		}
		if m.config.HasGithubPATIntegration {
			if strings.TrimSpace(m.config.GithubTokenPath) == "" {
				return fmt.Errorf("token secret path is required when PAT integration is enabled")
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
