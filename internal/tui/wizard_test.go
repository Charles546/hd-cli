// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If a separate written commercial agreement exists, you may use this file under those terms instead.

package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"

	"github.com/Charles546/hd-cli/internal/config"
)

func TestNewWizard(t *testing.T) {
	m := NewWizard(nil)
	if m == nil {
		t.Fatal("NewWizard returned nil")
	}
	if m.step != 1 {
		t.Errorf("Initial step = %d, want 1", m.step)
	}
	if m.total != stepCount {
		t.Errorf("Total steps = %d, want %d", m.total, stepCount)
	}
	if m.config == nil {
		t.Fatal("WizardModel config is nil")
	}
}

func TestNewWizardWithConfig(t *testing.T) {
	cfg := &config.WizardConfig{
		ProjectName: "custom-project",
	}
	m := NewWizard(cfg)
	if m.config.ProjectName != "custom-project" {
		t.Errorf("ProjectName = %q, want custom-project", m.config.ProjectName)
	}
}

func TestUpdateStep(t *testing.T) {
	m := NewWizard(nil)

	tests := []struct {
		step  int
		key   string
		value string
		check func(*config.WizardConfig) error
	}{
		{
			step: 2, key: "project_name", value: "test-project",
			check: func(c *config.WizardConfig) error {
				if c.ProjectName != "test-project" {
					return fmt.Errorf("ProjectName = %q", c.ProjectName)
				}
				return nil
			},
		},
		{
			step: 4, key: "deployment_mode", value: "kubernetes",
			check: func(c *config.WizardConfig) error {
				if c.DeploymentMode != "kubernetes" {
					return fmt.Errorf("DeploymentMode = %q", c.DeploymentMode)
				}
				return nil
			},
		},
		{
			step: 8, key: "ai_enabled", value: "true",
			check: func(c *config.WizardConfig) error {
				if !c.AIEnabled {
					return fmt.Errorf("AIEnabled = false")
				}
				return nil
			},
		},
		{
			step: 9, key: "secrets_backend", value: "dev",
			check: func(c *config.WizardConfig) error {
				if c.SecretsBackend != "dev" {
					return fmt.Errorf("SecretsBackend = %q", c.SecretsBackend)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		if err := UpdateStep(m, tt.step, tt.key, tt.value); err != nil {
			t.Errorf("UpdateStep(%d, %q, %q) error: %v", tt.step, tt.key, tt.value, err)
		}
		if err := tt.check(m.config); err != nil {
			t.Errorf("check failed for step %d: %v", tt.step, err)
		}
	}
}

func TestUpdateStepNilWizard(t *testing.T) {
	err := UpdateStep(nil, 1, "key", "value")
	if err == nil {
		t.Error("expected error for nil wizard")
	}
}

func TestStepLabels(t *testing.T) {
	labels := StepLabels()
	if len(labels) != stepCount+1 {
		t.Errorf("StepLabels count = %d, want %d", len(labels), stepCount+1)
	}
	if labels[1] != "Welcome" {
		t.Errorf("labels[1] = %q, want Welcome", labels[1])
	}
}

func TestIsStepRequired(t *testing.T) {
	requiredSteps := []int{2, 3, 4, 5, 6, 9, 10, 11, 12, 13}
	for _, step := range requiredSteps {
		if !IsStepRequired(step) {
			t.Errorf("step %d should be required", step)
		}
	}
	optionalSteps := []int{1, 7, 8, 14, 15}
	for _, step := range optionalSteps {
		if IsStepRequired(step) {
			t.Errorf("step %d should not be required", step)
		}
	}
}

func TestStepHelp(t *testing.T) {
	for i := 1; i <= stepCount; i++ {
		h := StepHelp(i)
		if h == "" {
			t.Errorf("StepHelp(%d) returned empty string", i)
		}
	}
}

func TestValidateStepComplete(t *testing.T) {
	tests := []struct {
		name    string
		step    int
		setup   func(*config.WizardConfig)
		wantErr bool
	}{
		{
			name:    "step 1 (welcome) always valid",
			step:    1,
			wantErr: false,
		},
		{
			name:    "step 2 with project name",
			step:    2,
			setup:   func(c *config.WizardConfig) { c.ProjectName = "test" },
			wantErr: false,
		},
		{
			name:    "step 2 without project name",
			step:    2,
			setup:   func(c *config.WizardConfig) { c.ProjectName = "" },
			wantErr: true,
		},
		{
			name:    "step 4 with valid mode",
			step:    4,
			setup:   func(c *config.WizardConfig) { c.DeploymentMode = "docker" },
			wantErr: false,
		},
		{
			name:    "step 4 with invalid mode",
			step:    4,
			setup:   func(c *config.WizardConfig) { c.DeploymentMode = "invalid" },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.NewDefaultWizardConfig()
			if tt.setup != nil {
				tt.setup(cfg)
			}
			m := NewWizard(cfg)
			m.step = tt.step
			err := ValidateStepComplete(m)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStepComplete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProgressBar(t *testing.T) {
	bar := ProgressBar(5, 15)
	if bar == "" {
		t.Error("ProgressBar returned empty string")
	}
	bar = ProgressBar(0, 0)
	if bar != "" {
		t.Error("ProgressBar(0,0) should return empty string")
	}
}

func TestBoolToRadio(t *testing.T) {
	if boolToRadio(true) != "yes" {
		t.Error("boolToRadio(true) should return 'yes'")
	}
	if boolToRadio(false) != "no" {
		t.Error("boolToRadio(false) should return 'no'")
	}
}

// --- New interactive wizard tests ---

func TestWizardModelInitStep(t *testing.T) {
	m := NewWizard(nil)

	// Step 1 should be navigate mode
	m = runInitStep(m, 1)
	if m.mode != modeNavigate {
		t.Errorf("Step 1 mode = %d, want modeNavigate(%d)", m.mode, modeNavigate)
	}

	// Step 2 should be single field text input
	m = runInitStep(m, 2)
	if m.mode != modeTextInput {
		t.Errorf("Step 2 mode = %d, want modeTextInput(%d)", m.mode, modeTextInput)
	}
	if len(m.textInputs) != 1 {
		t.Errorf("Step 2 textInputs count = %d, want 1", len(m.textInputs))
	}
	if !m.textInput.Focused() {
		t.Error("Step 2 text input should be focused")
	}

	// Step 4 should be radio select
	m = runInitStep(m, 4)
	if m.mode != modeRadioSelect {
		t.Errorf("Step 4 mode = %d, want modeRadioSelect(%d)", m.mode, modeRadioSelect)
	}

	// Step 5 should be multi-field text input
	m = runInitStep(m, 5)
	if m.mode != modeTextInput {
		t.Errorf("Step 5 mode = %d, want modeTextInput(%d)", m.mode, modeTextInput)
	}
	if len(m.textInputs) != 2 {
		t.Errorf("Step 5 textInputs count = %d, want 2", len(m.textInputs))
	}
}

func TestTextInputCapture(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	cfg.ProjectName = "" // Clear default for clean test
	m := NewWizard(cfg)
	m = runInitStep(m, 2) // Project Name step

	// Simulate typing "my-project"
	for _, ch := range "my-project" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}

	// The text input should contain what we typed
	if m.textInput.Value() != "my-project" {
		t.Errorf("textInput.Value() = %q, want %q", m.textInput.Value(), "my-project")
	}
}

func TestTextInputBackspace(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	cfg.ProjectName = "" // Clear default for clean test
	m := NewWizard(cfg)
	m = runInitStep(m, 2)

	// Type "abc" then backspace
	for _, ch := range "abc" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyBackspace})

	if m.textInput.Value() != "ab" {
		t.Errorf("textInput.Value() after backspace = %q, want %q", m.textInput.Value(), "ab")
	}
}

func TestTextInputEnterCommitsValue(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	cfg.ProjectName = ""
	m := NewWizard(cfg)
	m = runInitStep(m, 2)

	// Type a project name
	for _, ch := range "test-proj" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}

	// Press Enter to commit
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	// Config should be updated
	if m.config.ProjectName != "test-proj" {
		t.Errorf("ProjectName = %q, want %q", m.config.ProjectName, "test-proj")
	}
	// Should have moved to step 3
	if m.step != 3 {
		t.Errorf("step = %d, want 3", m.step)
	}
}

func TestMultiFieldTabNavigation(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	cfg.EssentialsRepoURL = "" // Clear defaults for clean test
	cfg.EssentialsBranch = ""
	m := NewWizard(cfg)
	m = runInitStep(m, 5) // Config Repo Setup - 2 fields

	// Type in first field
	for _, ch := range "https://example.com" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}

	// Press Tab to move to next field
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	// Should now be on field 1
	if m.currentField != 1 {
		t.Errorf("currentField = %d, want 1", m.currentField)
	}
	// First field value should be saved
	if m.textInputs[0].Value() != "https://example.com" {
		t.Errorf("field 0 value = %q, want %q", m.textInputs[0].Value(), "https://example.com")
	}
	// Second field should be focused
	if !m.textInputs[1].Focused() {
		t.Error("second field should be focused after tab")
	}
}

func TestMultiFieldEscBackNavigation(t *testing.T) {
	m := NewWizard(nil)
	m = runInitStep(m, 5) // 2 fields

	// Move to second field
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})
	if m.currentField != 1 {
		t.Fatalf("currentField = %d, want 1", m.currentField)
	}

	// Press Esc to go back to first field
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})

	if m.currentField != 0 {
		t.Errorf("currentField = %d, want 0", m.currentField)
	}
	if !m.textInputs[0].Focused() {
		t.Error("first field should be focused after esc")
	}
}

func TestMultiFieldEscOnFirstFieldGoesBack(t *testing.T) {
	m := NewWizard(nil)
	m = runInitStep(m, 5)

	// Press Esc on first field - should go to previous step
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})

	if m.step != 4 {
		t.Errorf("step = %d, want 4 (previous step)", m.step)
	}
}

func TestRadioSelectionDownArrow(t *testing.T) {
	m := NewWizard(nil)
	m = runInitStep(m, 4) // Deployment Mode

	// Default should be "docker" (index 0)
	if m.radioIndex != 0 {
		t.Fatalf("radioIndex = %d, want 0", m.radioIndex)
	}

	// Press down arrow
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})

	if m.radioIndex != 1 {
		t.Errorf("radioIndex = %d, want 1", m.radioIndex)
	}

	// Press down again
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})

	if m.radioIndex != 2 {
		t.Errorf("radioIndex = %d, want 2", m.radioIndex)
	}

	// Press down again (should not go past end)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})

	if m.radioIndex != 2 {
		t.Errorf("radioIndex = %d, want 2 (clamped)", m.radioIndex)
	}
}

func TestRadioSelectionUpArrow(t *testing.T) {
	m := NewWizard(nil)
	m = runInitStep(m, 4)

	// Move down twice
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.radioIndex != 2 {
		t.Fatalf("radioIndex = %d, want 2", m.radioIndex)
	}

	// Press up arrow
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyUp})

	if m.radioIndex != 1 {
		t.Errorf("radioIndex = %d, want 1", m.radioIndex)
	}

	// Press up to beginning
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyUp})

	if m.radioIndex != 0 {
		t.Errorf("radioIndex = %d, want 0", m.radioIndex)
	}

	// Press up again (should not go below 0)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyUp})

	if m.radioIndex != 0 {
		t.Errorf("radioIndex = %d, want 0 (clamped)", m.radioIndex)
	}
}

func TestRadioSelectionEnterCommits(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	cfg.DeploymentMode = "docker"
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// Move to "source" (index 1)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})

	// Press Enter to commit
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	// Config should be updated
	if m.config.DeploymentMode != "source" {
		t.Errorf("DeploymentMode = %q, want %q", m.config.DeploymentMode, "source")
	}
	// Should advance to next step
	if m.step != 5 {
		t.Errorf("step = %d, want 5", m.step)
	}
}

func TestRadioSelectionEscGoesBack(t *testing.T) {
	m := NewWizard(nil)
	m = runInitStep(m, 4)

	// Press Esc to go back
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})

	if m.step != 3 {
		t.Errorf("step = %d, want 3", m.step)
	}
}

func TestRadioSelectionBackspaceGoesBack(t *testing.T) {
	m := NewWizard(nil)
	m = runInitStep(m, 4)

	// Press Backspace to go back
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyBackspace})

	if m.step != 3 {
		t.Errorf("step = %d, want 3", m.step)
	}
}

func TestNavigationEnterNext(t *testing.T) {
	m := NewWizard(nil)
	m = runInitStep(m, 1) // Welcome step

	// Press Enter to go to next step
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.step != 2 {
		t.Errorf("step = %d, want 2", m.step)
	}
}

func TestNavigationBackFromStep1(t *testing.T) {
	m := NewWizard(nil)
	m = runInitStep(m, 1)

	// Press Esc on step 1 - should stay on step 1
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})

	if m.step != 1 {
		t.Errorf("step = %d, want 1 (should not go below 1)", m.step)
	}
}

func TestQuitWithCtrlC(t *testing.T) {
	m := NewWizard(nil)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if cmd == nil {
		t.Error("expected quit command for ctrl+c")
	}
}

func TestTextInputDefaultValue(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 2) // Project Name

	// The text input should be pre-filled with the default value
	if m.textInput.Value() != "hd-config" {
		t.Errorf("textInput.Value() = %q, want %q", m.textInput.Value(), "hd-config")
	}
}

func TestMultiFieldDefaultValue(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 5) // Config Repo Setup

	if m.textInputs[0].Value() != "https://github.com/honeydipper/honeydipper-config-essentials.git" {
		t.Errorf("field 0 = %q, want essentials URL", m.textInputs[0].Value())
	}
	if m.textInputs[1].Value() != "v4-rc" {
		t.Errorf("field 1 = %q, want %q", m.textInputs[1].Value(), "v4-rc")
	}
}

func TestRadioDefaultSelection(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4) // Deployment Mode

	// Default is "docker" which is index 0
	if m.radioIndex != 0 {
		t.Errorf("radioIndex = %d, want 0 (docker)", m.radioIndex)
	}
}

func TestViewContainsStepTitle(t *testing.T) {
	m := NewWizard(nil)
	m = runInitStep(m, 2)

	view := m.View()
	if !strings.Contains(view, "Project Name") {
		t.Errorf("View() should contain step title 'Project Name', got: %s", view)
	}
}

func TestViewContainsTextInput(t *testing.T) {
	m := NewWizard(nil)
	m = runInitStep(m, 2)

	view := m.View()
	// The view should contain the text input rendering
	if !strings.Contains(view, "Project name") {
		t.Errorf("View() should contain 'Project name' label, got: %s", view)
	}
}

func TestViewContainsRadioOptions(t *testing.T) {
	m := NewWizard(nil)
	m = runInitStep(m, 4)

	view := m.View()
	if !strings.Contains(view, "docker") {
		t.Errorf("View() should contain 'docker' option, got: %s", view)
	}
	if !strings.Contains(view, "source") {
		t.Errorf("View() should contain 'source' option, got: %s", view)
	}
	if !strings.Contains(view, "kubernetes") {
		t.Errorf("View() should contain 'kubernetes' option, got: %s", view)
	}
}

func TestViewContainsNavigationHints(t *testing.T) {
	m := NewWizard(nil)
	m = runInitStep(m, 2)

	view := m.View()
	if !strings.Contains(view, "q=quit") {
		t.Errorf("View() should contain 'q=quit' hint, got: %s", view)
	}
}

func TestViewSummaryScreen(t *testing.T) {
	m := NewWizard(nil)
	m.done = true

	view := m.View()
	if !strings.Contains(view, "Configuration Summary") {
		t.Errorf("View() should contain 'Configuration Summary', got: %s", view)
	}
	if !strings.Contains(view, "Project:") {
		t.Errorf("View() should contain 'Project:' in summary, got: %s", view)
	}
}

func TestDoneScreenEnterConfirms(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m.done = true

	// Press Enter - should trigger generateConfig and return a quit command
	m, cmd := updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Error("expected quit command after enter on done screen")
	}
	// The model should still be in done state
	if !m.done {
		t.Error("model should still be in done state")
	}
}

func TestDoneScreenEscGoesBack(t *testing.T) {
	m := NewWizard(nil)
	m.done = true

	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})

	if m.done {
		t.Error("should no longer be in done state after esc")
	}
}

func TestTextInputEmptyBackspaceGoesBack(t *testing.T) {
	m := NewWizard(nil)
	m = runInitStep(m, 2)

	// Ensure text input is empty
	m.textInput.SetValue("")

	// Press Backspace on empty input - should go to previous step
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyBackspace})

	if m.step != 1 {
		t.Errorf("step = %d, want 1 (back from step 2)", m.step)
	}
}

func TestMultiFieldLastFieldEnterAdvances(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 5) // 2 fields

	// Move to last field
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})
	if m.currentField != 1 {
		t.Fatalf("currentField = %d, want 1", m.currentField)
	}

	// Press Enter on last field - should advance to next step
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.step != 6 {
		t.Errorf("step = %d, want 6", m.step)
	}
}

func TestGetStepInfoAllSteps(t *testing.T) {
	for i := 1; i <= stepCount; i++ {
		info := getStepInfo(i)
		if info == nil {
			t.Errorf("getStepInfo(%d) returned nil", i)
		}
	}
}

func TestGetStepInfoInvalidStep(t *testing.T) {
	info := getStepInfo(0)
	if info != nil {
		t.Error("getStepInfo(0) should return nil")
	}
	info = getStepInfo(99)
	if info != nil {
		t.Error("getStepInfo(99) should return nil")
	}
}

func TestRadioToBool(t *testing.T) {
	if !radioToBool("yes") {
		t.Error("radioToBool('yes') should return true")
	}
	if radioToBool("no") {
		t.Error("radioToBool('no') should return false")
	}
	if radioToBool("maybe") {
		t.Error("radioToBool('maybe') should return false")
	}
}

func TestParseInt(t *testing.T) {
	if parseInt("42") != 42 {
		t.Errorf("parseInt('42') = %d, want 42", parseInt("42"))
	}
	if parseInt("abc") != 0 {
		t.Errorf("parseInt('abc') = %d, want 0", parseInt("abc"))
	}
	if parseInt("") != 0 {
		t.Errorf("parseInt('') = %d, want 0", parseInt(""))
	}
}

func TestRenderStepTitle(t *testing.T) {
	result := renderStepTitle("Test Title")
	if !strings.Contains(result, "Test Title") {
		t.Errorf("renderStepTitle() = %q, want to contain 'Test Title'", result)
	}
}

func TestWizardModelInitMethod(t *testing.T) {
	m := NewWizard(nil)
	cmd := m.Init()
	// Init should return a cmd (from initStep)
	_ = cmd
	if m.step != 1 {
		t.Errorf("after Init(), step = %d, want 1", m.step)
	}
}

// runInitStep is a helper that calls initStep and returns the updated model.
// Since initStep uses value receiver internally but modifies the model,
// we need to call it properly.
func runInitStep(m *WizardModel, step int) *WizardModel {
	m.initStep(step)
	return m
}


// updateWizard is a helper that calls Update and type-asserts the result.
func updateWizard(m *WizardModel, msg tea.Msg) (*WizardModel, tea.Cmd) {
	model, cmd := m.Update(msg)
	return model.(*WizardModel), cmd
}

// updateWizardCmd is a helper that calls Update and returns both with type assertion.
func updateWizardCmd(m *WizardModel, msg tea.Msg) (*WizardModel, tea.Cmd) {
	model, cmd := m.Update(msg)
	return model.(*WizardModel), cmd
}

// Ensure textinput.Model is used (compile-time check)
var _ textinput.Model
