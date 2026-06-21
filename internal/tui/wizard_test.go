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
			step: 6, key: "secrets_backend", value: "dev",
			check: func(c *config.WizardConfig) error {
				if c.SecretsBackend != "dev" {
					return fmt.Errorf("SecretsBackend = %q", c.SecretsBackend)
				}
				return nil
			},
		},
		{
			step: 10, key: "ai_enabled", value: "true",
			check: func(c *config.WizardConfig) error {
				if !c.AIEnabled {
					return fmt.Errorf("AIEnabled = false")
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
	// Verify new order
	if labels[6] != "Secrets Backend" {
		t.Errorf("labels[6] = %q, want Secrets Backend", labels[6])
	}
	if labels[7] != "Redis" {
		t.Errorf("labels[7] = %q, want Redis", labels[7])
	}
	if labels[8] != "GitHub Integration" {
		t.Errorf("labels[8] = %q, want GitHub Integration", labels[8])
	}
	if labels[9] != "Slack Integration" {
		t.Errorf("labels[9] = %q, want Slack Integration", labels[9])
	}
	if labels[10] != "AI Agent" {
		t.Errorf("labels[10] = %q, want AI Agent", labels[10])
	}
}

func TestIsStepRequired(t *testing.T) {
	requiredSteps := []int{2, 3, 4, 5, 6, 7, 10, 11, 12, 13}
	for _, step := range requiredSteps {
		if !IsStepRequired(step) {
			t.Errorf("step %d should be required", step)
		}
	}
	optionalSteps := []int{1, 8, 9, 14, 15}
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
		{
			name:    "step 6 with secrets backend selected",
			step:    6,
			setup:   func(c *config.WizardConfig) { c.SecretsBackend = "vault" },
			wantErr: false,
		},
		{
			name:    "step 6 without secrets backend",
			step:    6,
			setup:   func(c *config.WizardConfig) { c.SecretsBackend = "" },
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
	if len(m.textInputs) != 3 {
		t.Errorf("Step 5 textInputs count = %d, want 3", len(m.textInputs))
	}

	// Step 6 should be radio select (Secrets Backend)
	m = runInitStep(m, 6)
	if m.mode != modeRadioSelect {
		t.Errorf("Step 6 mode = %d, want modeRadioSelect(%d)", m.mode, modeRadioSelect)
	}

	// Step 8 should be checkbox select (GitHub Integration)
	m = runInitStep(m, 8)
	if m.mode != modeCheckboxSelect {
		t.Errorf("Step 8 mode = %d, want modeCheckboxSelect(%d)", m.mode, modeCheckboxSelect)
	}

	// Step 9 should be multi-field text input (Slack Integration)
	m = runInitStep(m, 9)
	if m.mode != modeTextInput {
		t.Errorf("Step 9 mode = %d, want modeTextInput(%d)", m.mode, modeTextInput)
	}
	if len(m.textInputs) != 4 {
		t.Errorf("Step 9 textInputs count = %d, want 4", len(m.textInputs))
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
	cfg.EssentialsCloneType = ""
	cfg.EssentialsClonePAT = ""
	cfg.EssentialsCloneKey = ""
	m := NewWizard(cfg)
	m = runInitStep(m, 5) // Config Repo Setup - 5 fields

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
	m = runInitStep(m, 5) // 5 fields

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
	if !strings.Contains(view, "ctrl+q=quit") {
		t.Errorf("View() should contain 'ctrl+q=quit' hint, got: %s", view)
	}
}

func TestCtrlQInTextInputModeQuits(t *testing.T) {
	// Ctrl+Q should quit even in text input mode
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 2) // Project Name step (text input mode)

	// Press Ctrl+Q - should quit
	_, cmd := updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyCtrlQ})

	if cmd == nil {
		t.Error("ctrl+q in text input mode should produce a quit command")
	}
}

func TestCtrlQOnDoneScreenQuits(t *testing.T) {
	// Ctrl+Q on the done screen should quit without generating config
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m.done = true

	// Press ctrl+q - should quit without generating
	_, cmd := updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyCtrlQ})

	if cmd == nil {
		t.Error("ctrl+q on done screen should produce a quit command")
	}
	// err should be nil (no config generation)
	if m.err != nil {
		t.Errorf("err = %v, want nil (no config generation on quit)", m.err)
	}
}

func TestViewSummaryScreen(t *testing.T) {
	// With refinement 1, the summary is shown at step 15, not on the done screen.
	// The done screen shows a confirmation prompt instead.
	m := NewWizard(nil)
	m.done = true

	view := m.View()
	// Done screen should show the confirmation prompt
	if !strings.Contains(view, "Press Enter to confirm and generate configs") {
		t.Errorf("View() should contain confirmation prompt, got: %s", view)
	}
	// Done screen should NOT contain the summary (it's shown at step 15)
	if strings.Contains(view, "Configuration Summary") {
		t.Errorf("Done screen should NOT contain 'Configuration Summary' (shown at step 15)")
	}
}

func TestStep15ShowsSummary(t *testing.T) {
	// Refinement 1: Step 15 should show the summary directly
	m := NewWizard(nil)
	m = runInitStep(m, 15)

	view := m.View()
	if !strings.Contains(view, "Configuration Summary") {
		t.Errorf("Step 15 View() should contain 'Configuration Summary', got: %s", view)
	}
	if !strings.Contains(view, "Project:") {
		t.Errorf("Step 15 View() should contain 'Project:' in summary, got: %s", view)
	}
	// Verify new order in summary
	if !strings.Contains(view, "Secrets backend:") {
		t.Errorf("Step 15 View() should contain 'Secrets backend:' in summary, got: %s", view)
	}
	if !strings.Contains(view, "Redis:") {
		t.Errorf("Step 15 View() should contain 'Redis:' in summary, got: %s", view)
	}
	// Step 15 should also show the generate prompt
	if !strings.Contains(view, "Press Enter to generate configs") {
		t.Errorf("Step 15 View() should contain 'Press Enter to generate configs', got: %s", view)
	}
	// Step 15 should show the save hint
	if !strings.Contains(view, "s=save answers") {
		t.Errorf("Step 15 View() should contain save hint, got: %s", view)
	}
}

func TestStep15EnterGeneratesDirectly(t *testing.T) {
	// Fix 1: Pressing Enter on step 15 should generate config directly
	// without going through the done/confirmation screen.
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 15) // Summary step (navigate mode)

	// Press Enter - should trigger generateConfig and return quit
	m, cmd := updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Error("expected quit command after enter on step 15")
	}
	// Should NOT be in done state (direct generation, no intermediate done)
	if m.done {
		t.Error("should not be in done state after enter on step 15")
	}
}

func TestStep15SaveAndGenerateDirectly(t *testing.T) {
	// Fix 1: Pressing 's' on step 15 should save answers and generate directly.
	cfg := config.NewDefaultWizardConfig()
	cfg.ProjectName = "save-gen-test"
	m := NewWizard(cfg)
	m = runInitStep(m, 15)

	// Press 's' - should save and generate
	m, cmd := updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})

	if cmd == nil {
		t.Error("expected quit command after 's' on step 15")
	}
	if m.done {
		t.Error("should not be in done state after 's' on step 15")
	}
}

func TestStep15EscGoesBack(t *testing.T) {
	// Pressing Esc on step 15 should go back to previous step
	m := NewWizard(nil)
	m = runInitStep(m, 15)

	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})

	if m.step == 15 {
		t.Errorf("step = %d, should have gone back from step 15", m.step)
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

// TestStepOrder verifies the correct order of wizard steps
func TestStepOrder(t *testing.T) {
	expectedOrder := []struct {
		step  int
		title string
	}{
		{1, "Welcome"},
		{2, "Project Name"},
		{3, "Config Directory"},
		{4, "Deployment Mode"},
		{5, "Config Repo Setup"},
		{6, "Secrets Backend"},  // Moved up
		{7, "Redis"},             // Moved up
		{8, "GitHub Integration"}, // Moved down
		{9, "Slack Integration"},  // Moved down
		{10, "AI Agent"},          // Moved down
		{11, "Docker Configuration"},
		{12, "Kubernetes Configuration"},
		{13, "Source Configuration"},
		{14, "GitHub Repo Creation"},
		{15, "Summary & Confirm"},
	}

	for _, expected := range expectedOrder {
		info := getStepInfo(expected.step)
		if info == nil {
			t.Fatalf("Step %d not found", expected.step)
		}
		if !strings.Contains(info.title, expected.title) {
			t.Errorf("Step %d title = %q, want to contain %q", expected.step, info.title, expected.title)
		}
	}
}

func TestRunWizardWithConfig(t *testing.T) {
	cfg := &config.WizardConfig{
		ProjectName:    "preloaded-project",
		DeploymentMode: "kubernetes",
		SecretsBackend: "dev",
	}
	// Just verify the wizard model is created correctly with pre-populated config
	m := NewWizard(cfg)
	if m.config.ProjectName != "preloaded-project" {
		t.Errorf("ProjectName = %q, want preloaded-project", m.config.ProjectName)
	}
	if m.config.DeploymentMode != "kubernetes" {
		t.Errorf("DeploymentMode = %q, want kubernetes", m.config.DeploymentMode)
	}
}

func TestCtrlSSaveAtAnyStep(t *testing.T) {
	// Refinement 3: Ctrl+S should save answers at any non-text-input step
	cfg := config.NewDefaultWizardConfig()
	cfg.ProjectName = "save-test"
	m := NewWizard(cfg)
	m = runInitStep(m, 1) // Welcome step (navigate mode)

	// Press Ctrl+S
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlS})

	// Should show save confirmation message
	view := m.View()
	if !strings.Contains(view, "Answers saved to") {
		t.Errorf("View() should contain save confirmation, got: %s", view)
	}
	// Should still be on the same step
	if m.step != 1 {
		t.Errorf("step = %d, want 1 (should not advance after save)", m.step)
	}
}

func TestCtrlSSaveOnRadioStep(t *testing.T) {
	// Refinement 3: Ctrl+S should save on radio select steps
	cfg := config.NewDefaultWizardConfig()
	cfg.ProjectName = "radio-save-test"
	m := NewWizard(cfg)
	m = runInitStep(m, 4) // Deployment Mode (radio)

	// Press Ctrl+S
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlS})

	// Should show save confirmation
	view := m.View()
	if !strings.Contains(view, "Answers saved to") {
		t.Errorf("View() should contain save confirmation, got: %s", view)
	}
	// Should still be on step 4
	if m.step != 4 {
		t.Errorf("step = %d, want 4", m.step)
	}
}

func TestCtrlSWorksInTextInputMode(t *testing.T) {
	// Ctrl+S should now work even in text input mode
	cfg := config.NewDefaultWizardConfig()
	cfg.ProjectName = "text-save-test"
	m := NewWizard(cfg)
	m = runInitStep(m, 2) // Project Name (text input)

	// Type some text first
	for _, ch := range "my-project" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}

	// Press Ctrl+S - should save answers even in text input mode
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlS})

	view := m.View()
	if !strings.Contains(view, "Answers saved to") {
		t.Errorf("Ctrl+S should trigger save in text input mode, got: %s", view)
	}
}

func TestCtrlSSavesCurrentTextInputValue(t *testing.T) {
	// Bug fix: Ctrl+S should commit the current text input value
	// to WizardConfig before saving, otherwise the typed value is lost.
	cfg := config.NewDefaultWizardConfig()
	cfg.ProjectName = "save-current-value"
	m := NewWizard(cfg)
	m = runInitStep(m, 2) // Project Name (text input)

	// Type a new project name
	for _, ch := range "new-name" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}

	// Press Ctrl+S - should commit text and save
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlS})

	// The config should have the typed value
	if m.config.ProjectName != "new-name" {
		t.Errorf("ProjectName = %q, want %q (Ctrl+S should commit current text)", m.config.ProjectName, "new-name")
	}
}

func TestCtrlSSavesCheckboxTextFieldValue(t *testing.T) {
	// Ctrl+S in checkbox mode with active text field should commit
	// the text field value before saving.
	cfg := config.NewDefaultWizardConfig()
	cfg.ProjectName = "checkbox-save"
	cfg.HasGitHubAppIntegration = true // Enable conditional text fields
	m := NewWizard(cfg)
	m = runInitStep(m, 8) // GitHub Integration (checkbox)

	// Move to text field (past checkboxes)
	m.currentField = 1
	if len(m.textInputs) > 0 {
		m.textInput = m.textInputs[0]
		m.textInput.Focus()
	}

	// Type a value in the text field
	for _, ch := range "my-app-id" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}

	// Press Ctrl+S
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlS})

	// The config should have the typed value committed
	if m.config.GithubAppID != "my-app-id" {
		t.Errorf("GithubAppID = %q, want %q (Ctrl+S should commit checkbox text field)", m.config.GithubAppID, "my-app-id")
	}
}

func TestSaveMsgClearedOnStepAdvance(t *testing.T) {
	// After saving with Ctrl+S and advancing, the save message should be cleared
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 1) // Welcome step

	// Save
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlS})
	if !strings.Contains(m.View(), "Answers saved to") {
		t.Fatal("Expected save confirmation after Ctrl+S")
	}

	// Advance to next step
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	// Save message should be cleared
	if m.saveMsg != "" {
		t.Errorf("saveMsg should be cleared after advancing, got: %q", m.saveMsg)
	}
}

func TestNavigationHintsIncludeCtrlS(t *testing.T) {
	// Navigation hints should include ctrl+s=save for non-done, non-text-input steps
	m := NewWizard(nil)
	m = runInitStep(m, 1) // Welcome step

	view := m.View()
	if !strings.Contains(view, "ctrl+s=save") {
		t.Errorf("View() should contain 'ctrl+s=save' hint, got: %s", view)
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

func TestTextInputSelectAllOnFocus(t *testing.T) {
	// Use default config which has ProjectName = "hd-config"
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 2) // Project Name step - has default "hd-config"

	// The text input should have the default value
	if m.textInput.Value() != "hd-config" {
		t.Fatalf("textInput.Value() = %q, want %q", m.textInput.Value(), "hd-config")
	}

	// pendingDefault should be true
	if !m.pendingDefault {
		t.Fatal("pendingDefault should be true when text input has a default value")
	}

	// Simulate typing a character - should clear the default first
	for _, ch := range "my" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}

	// After typing, the value should be "my" (default was cleared and replaced)
	if m.textInput.Value() != "my" {
		t.Errorf("textInput.Value() after typing = %q, want %q", m.textInput.Value(), "my")
	}

	// pendingDefault should now be false
	if m.pendingDefault {
		t.Error("pendingDefault should be false after typing")
	}
}

func TestTextInputBackspaceClearsDefault(t *testing.T) {
	// Use default config which has ProjectName = "hd-config"
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 2)

	// pendingDefault should be true
	if !m.pendingDefault {
		t.Fatal("pendingDefault should be true")
	}

	// Simulate Backspace - should clear the entire default
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyBackspace})

	// Value should be empty (default was cleared)
	if m.textInput.Value() != "" {
		t.Errorf("textInput.Value() after backspace = %q, want empty", m.textInput.Value())
	}

	// pendingDefault should be false
	if m.pendingDefault {
		t.Error("pendingDefault should be false after backspace clears default")
	}
}

// Ensure textinput.Model is used (compile-time check)
var _ textinput.Model

// Bug fix tests: ctrl+q and ctrl+c behavior

func TestQInTextInputModeIsRegularChar(t *testing.T) {
	// q is now a regular character — no special quit handling.
	cfg := config.NewDefaultWizardConfig()
	cfg.ProjectName = "" // Clear default so we can type cleanly
	m := NewWizard(cfg)
	m = runInitStep(m, 2) // Project Name step (text input mode)

	// Press 'q' - should be added to text input as a regular character
	m, _ = updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	// The text input should contain 'q'
	if m.textInput.Value() != "q" {
		t.Errorf("textInput.Value() = %q, want %q", m.textInput.Value(), "q")
	}
	// Should still be on step 2
	if m.step != 2 {
		t.Errorf("step = %d, want 2 (should not advance)", m.step)
	}
	// err should be nil (no config generation)
	if m.err != nil {
		t.Errorf("err = %v, want nil", m.err)
	}
}

func TestCtrlCInTextInputModeQuits(t *testing.T) {
	// Ctrl+C should always quit, even in text input mode
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 2) // Project Name step (text input mode)

	// Press Ctrl+C - should quit
	_, cmd := updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyCtrlC})

	if cmd == nil {
		t.Error("ctrl+c in text input mode should produce a quit command")
	}
}

func TestCtrlQOnDoneScreenQuitsWithoutGenerating(t *testing.T) {
	// Pressing ctrl+q on the done screen should quit without generating config
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m.done = true

	// Press ctrl+q - should quit without generating
	_, cmd := updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyCtrlQ})

	if cmd == nil {
		t.Error("ctrl+q on done screen should produce a quit command")
	}
	// err should be nil (no config generation)
	if m.err != nil {
		t.Errorf("err = %v, want nil (no config generation on quit)", m.err)
	}
}

func TestCtrlQOnRadioStepQuits(t *testing.T) {
	// Pressing ctrl+q on a radio select step should quit immediately
	m := NewWizard(nil)
	m = runInitStep(m, 4) // Deployment Mode (radio select)

	_, cmd := updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyCtrlQ})

	if cmd == nil {
		t.Error("ctrl+q on radio step should produce a quit command")
	}
}

func TestCtrlQOnCheckboxStepQuits(t *testing.T) {
	// Pressing ctrl+q on a checkbox step should quit immediately
	m := NewWizard(nil)
	m = runInitStep(m, 8) // GitHub Integration (checkbox)

	_, cmd := updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyCtrlQ})

	if cmd == nil {
		t.Error("ctrl+q on checkbox step should produce a quit command")
	}
}

func TestCtrlQOnNavigateStepQuits(t *testing.T) {
	// Pressing ctrl+q on a navigate step should quit immediately
	m := NewWizard(nil)
	m = runInitStep(m, 1) // Welcome (navigate mode)

	_, cmd := updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyCtrlQ})

	if cmd == nil {
		t.Error("ctrl+q on navigate step should produce a quit command")
	}
}

func TestCtrlCAlwaysQuitsFromAnyMode(t *testing.T) {
	// Ctrl+C should quit from every mode
	modes := []struct {
		name string
		step int
	}{
		{"navigate", 1},
		{"text input", 2},
		{"radio", 4},
		{"checkbox", 8},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			m := NewWizard(nil)
			m = runInitStep(m, mode.step)

			_, cmd := updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyCtrlC})

			if cmd == nil {
				t.Errorf("ctrl+c in %s mode should produce a quit command", mode.name)
			}
		})
	}
}

func TestQInCheckboxTextFieldIsRegularChar(t *testing.T) {
	// q is a regular character everywhere, including checkbox text fields
	cfg := config.NewDefaultWizardConfig()
	cfg.HasGitHubAppIntegration = true // Enable to get text fields
	m := NewWizard(cfg)
	m = runInitStep(m, 8) // GitHub Integration (checkbox)

	// Move to text field (past checkboxes)
	m.currentField = 1
	if len(m.textInputs) > 0 {
		m.textInput = m.textInputs[0]
		m.textInput.Focus()
	}

	// Press 'q' - should be treated as text input
	m, _ = updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	if m.textInput.Value() != "q" {
		t.Errorf("textInput.Value() = %q, want %q", m.textInput.Value(), "q")
	}
	// Should still be on step 8
	if m.step != 8 {
		t.Errorf("step = %d, want 8 (should not advance)", m.step)
	}
}

// TestQuitSetsQuitFlag verifies that pressing q sets the quit flag
// so RunWizard() can distinguish quit from completion.
func TestQuitSetsQuitFlag(t *testing.T) {
	// Press ctrl+q on a navigate step (welcome)
	m := NewWizard(nil)
	m = runInitStep(m, 1)

	m, cmd := updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyCtrlQ})

	if !m.quit {
		t.Error("quit flag should be true after pressing ctrl+q")
	}
	if cmd == nil {
		t.Error("cmd should be tea.Quit (non-nil)")
	}
}

func TestQuitSetsQuitFlagOnRadioStep(t *testing.T) {
	m := NewWizard(nil)
	m = runInitStep(m, 4) // radio select

	m, _ = updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyCtrlQ})

	if !m.quit {
		t.Error("quit flag should be true after pressing ctrl+q on radio step")
	}
}

func TestQuitSetsQuitFlagOnCheckboxStep(t *testing.T) {
	m := NewWizard(nil)
	m = runInitStep(m, 8) // checkbox

	m, _ = updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyCtrlQ})

	if !m.quit {
		t.Error("quit flag should be true after pressing ctrl+q on checkbox step")
	}
}

func TestCtrlCSetsQuitFlag(t *testing.T) {
	m := NewWizard(nil)
	m = runInitStep(m, 1)

	m, _ = updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyCtrlC})

	if !m.quit {
		t.Error("quit flag should be true after pressing ctrl+c")
	}
}

func TestCompleteDoesNotSetQuitFlag(t *testing.T) {
	// When user completes the wizard (presses Enter on step 15),
	// quit should remain false.
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 15)

	m, _ = updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.quit {
		t.Error("quit flag should be false after completing the wizard")
	}
}

func TestQuitFlagInitiallyFalse(t *testing.T) {
	m := NewWizard(nil)
	if m.quit {
		t.Error("quit flag should be false initially")
	}
}

// ===== Dev mode secret field tests =====

func TestGhSecretLabels_VaultMode(t *testing.T) {
	cfg := &config.WizardConfig{SecretsBackend: "vault"}

	label, placeholder, help := ghSecretLabels(cfg, "key")
	if label != "Private key secret path" {
		t.Errorf("vault key label = %q, want %q", label, "Private key secret path")
	}
	if placeholder != "Path to secret containing the private key" {
		t.Errorf("vault key placeholder = %q, want %q", placeholder, "Path to secret containing the private key")
	}
	if help != "Path to the secret containing the GitHub App private key (only for github_app integration)" {
		t.Errorf("vault key help = %q, want path help", help)
	}

	label, placeholder, _ = ghSecretLabels(cfg, "token")
	if label != "Token secret path" {
		t.Errorf("vault token label = %q, want %q", label, "Token secret path")
	}
	if placeholder != "Path to secret containing the PAT" {
		t.Errorf("vault token placeholder = %q, want %q", placeholder, "Path to secret containing the PAT")
	}

	label, placeholder, _ = ghSecretLabels(cfg, "webhook")
	if label != "Webhook secret path" {
		t.Errorf("vault webhook label = %q, want %q", label, "Webhook secret path")
	}
	if placeholder != "Path to webhook signing secret" {
		t.Errorf("vault webhook placeholder = %q, want %q", placeholder, "Path to webhook signing secret")
	}
}

func TestGhSecretLabels_DevMode(t *testing.T) {
	cfg := &config.WizardConfig{SecretsBackend: "dev"}

	label, placeholder, help := ghSecretLabels(cfg, "key")
	if label != "Private key value" {
		t.Errorf("dev key label = %q, want %q", label, "Private key value")
	}
	if placeholder != "Plain text or $ENV_VAR" {
		t.Errorf("dev key placeholder = %q, want %q", placeholder, "Plain text or $ENV_VAR")
	}
	if help != "GitHub App private key value or env var name (only for github_app integration)" {
		t.Errorf("dev key help = %q, want value help", help)
	}

	label, placeholder, _ = ghSecretLabels(cfg, "token")
	if label != "Token value" {
		t.Errorf("dev token label = %q, want %q", label, "Token value")
	}
	if placeholder != "Plain text or $ENV_VAR" {
		t.Errorf("dev token placeholder = %q, want %q", placeholder, "Plain text or $ENV_VAR")
	}

	label, placeholder, _ = ghSecretLabels(cfg, "webhook")
	if label != "Webhook secret value" {
		t.Errorf("dev webhook label = %q, want %q", label, "Webhook secret value")
	}
	if placeholder != "Plain text or $ENV_VAR" {
		t.Errorf("dev webhook placeholder = %q, want %q", placeholder, "Plain text or $ENV_VAR")
	}
}

func TestSlackSecretLabels_VaultMode(t *testing.T) {
	cfg := &config.WizardConfig{SecretsBackend: "vault"}

	label, placeholder, _ := slackSecretLabels(cfg, "bot_token")
	if label != "Bot token secret path" {
		t.Errorf("vault bot_token label = %q, want %q", label, "Bot token secret path")
	}
	if placeholder != "Path to Slack bot token" {
		t.Errorf("vault bot_token placeholder = %q, want %q", placeholder, "Path to Slack bot token")
	}

	label, _, _ = slackSecretLabels(cfg, "signing")
	if label != "Signing secret path" {
		t.Errorf("vault signing label = %q, want %q", label, "Signing secret path")
	}

	label, _, _ = slackSecretLabels(cfg, "interaction")
	if label != "Interaction token" {
		t.Errorf("vault interaction label = %q, want %q", label, "Interaction token")
	}

	label, _, _ = slackSecretLabels(cfg, "slash")
	if label != "Slash command token" {
		t.Errorf("vault slash label = %q, want %q", label, "Slash command token")
	}
}

func TestSlackSecretLabels_DevMode(t *testing.T) {
	cfg := &config.WizardConfig{SecretsBackend: "dev"}

	label, placeholder, _ := slackSecretLabels(cfg, "bot_token")
	if label != "Bot token value" {
		t.Errorf("dev bot_token label = %q, want %q", label, "Bot token value")
	}
	if placeholder != "Plain text or $ENV_VAR" {
		t.Errorf("dev bot_token placeholder = %q, want %q", placeholder, "Plain text or $ENV_VAR")
	}

	label, _, _ = slackSecretLabels(cfg, "signing")
	if label != "Signing secret value" {
		t.Errorf("dev signing label = %q, want %q", label, "Signing secret value")
	}

	label, placeholder, _ = slackSecretLabels(cfg, "interaction")
	if label != "Interaction token value" {
		t.Errorf("dev interaction label = %q, want %q", label, "Interaction token value")
	}
	if placeholder != "Plain text or $ENV_VAR (optional)" {
		t.Errorf("dev interaction placeholder = %q, want %q", placeholder, "Plain text or $ENV_VAR (optional)")
	}

	label, placeholder, _ = slackSecretLabels(cfg, "slash")
	if label != "Slash command token value" {
		t.Errorf("dev slash label = %q, want %q", label, "Slash command token value")
	}
	if placeholder != "Plain text or $ENV_VAR (optional)" {
		t.Errorf("dev slash placeholder = %q, want %q", placeholder, "Plain text or $ENV_VAR (optional)")
	}
}

func TestGhSecretFieldKey(t *testing.T) {
	tests := []struct {
		label    string
		expected string
	}{
		{"Private key secret path", "key"},
		{"Private key value", "key"},
		{"Token secret path", "token"},
		{"Token value", "token"},
		{"Webhook secret path", "webhook"},
		{"Webhook secret value", "webhook"},
		{"App ID", ""},
		{"Installation ID", ""},
		{"Something else", ""},
	}

	for _, tt := range tests {
		got := ghSecretFieldKey(tt.label)
		if got != tt.expected {
			t.Errorf("ghSecretFieldKey(%q) = %q, want %q", tt.label, got, tt.expected)
		}
	}
}

func TestSlackSecretFieldKey(t *testing.T) {
	tests := []struct {
		label    string
		expected string
	}{
		{"Bot token secret path", "bot_token"},
		{"Bot token value", "bot_token"},
		{"Signing secret path", "signing"},
		{"Signing secret value", "signing"},
		{"Interaction token", "interaction"},
		{"Interaction token value", "interaction"},
		{"Slash command token", "slash"},
		{"Slash command token value", "slash"},
		{"Something else", ""},
	}

	for _, tt := range tests {
		got := slackSecretFieldKey(tt.label)
		if got != tt.expected {
			t.Errorf("slackSecretFieldKey(%q) = %q, want %q", tt.label, got, tt.expected)
		}
	}
}

func TestIsDevMode(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.WizardConfig
		expected bool
	}{
		{"nil config", nil, false},
		{"vault backend", &config.WizardConfig{SecretsBackend: "vault"}, false},
		{"dev backend", &config.WizardConfig{SecretsBackend: "dev"}, true},
		{"empty backend", &config.WizardConfig{SecretsBackend: ""}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDevMode(tt.cfg)
			if got != tt.expected {
				t.Errorf("isDevMode() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStep8DevModeLabels(t *testing.T) {
	// When secrets backend is dev, the GitHub integration step should show
	// dev-mode labels for secret fields
	cfg := config.NewDefaultWizardConfig()
	cfg.SecretsBackend = "dev"
	cfg.HasGitHubAppIntegration = true
	cfg.HasGithubPATIntegration = true
	m := NewWizard(cfg)
	m = runInitStep(m, 8)

	view := m.View()

	// In dev mode, should show "Private key value" instead of "Private key secret path"
	if !strings.Contains(view, "Private key value") {
		t.Errorf("Step 8 dev mode view should contain 'Private key value', got: %s", view)
	}
	if !strings.Contains(view, "Token value") {
		t.Errorf("Step 8 dev mode view should contain 'Token value', got: %s", view)
	}
	if !strings.Contains(view, "Webhook secret value") {
		t.Errorf("Step 8 dev mode view should contain 'Webhook secret value', got: %s", view)
	}
	// Should NOT show vault-mode labels
	if strings.Contains(view, "Private key secret path") {
		t.Errorf("Step 8 dev mode view should NOT contain 'Private key secret path'")
	}
	if strings.Contains(view, "Token secret path") {
		t.Errorf("Step 8 dev mode view should NOT contain 'Token secret path'")
	}
}

func TestStep8VaultModeLabels(t *testing.T) {
	// When secrets backend is vault, the GitHub integration step should show
	// vault-mode labels for secret fields (backward compatible)
	cfg := config.NewDefaultWizardConfig()
	cfg.SecretsBackend = "vault"
	cfg.HasGitHubAppIntegration = true
	cfg.HasGithubPATIntegration = true
	m := NewWizard(cfg)
	m = runInitStep(m, 8)

	view := m.View()

	// In vault mode, should show vault labels
	if !strings.Contains(view, "Private key secret path") {
		t.Errorf("Step 8 vault mode view should contain 'Private key secret path', got: %s", view)
	}
	if !strings.Contains(view, "Token secret path") {
		t.Errorf("Step 8 vault mode view should contain 'Token secret path', got: %s", view)
	}
	if !strings.Contains(view, "Webhook secret path") {
		t.Errorf("Step 8 vault mode view should contain 'Webhook secret path', got: %s", view)
	}
}

func TestStep9DevModeLabels(t *testing.T) {
	// When secrets backend is dev, the Slack integration step should show
	// dev-mode labels
	cfg := config.NewDefaultWizardConfig()
	cfg.SecretsBackend = "dev"
	m := NewWizard(cfg)
	m = runInitStep(m, 9)

	view := m.View()

	if !strings.Contains(view, "Bot token value") {
		t.Errorf("Step 9 dev mode view should contain 'Bot token value', got: %s", view)
	}
	if !strings.Contains(view, "Signing secret value") {
		t.Errorf("Step 9 dev mode view should contain 'Signing secret value', got: %s", view)
	}
	if !strings.Contains(view, "Interaction token value") {
		t.Errorf("Step 9 dev mode view should contain 'Interaction token value', got: %s", view)
	}
	if !strings.Contains(view, "Slash command token value") {
		t.Errorf("Step 9 dev mode view should contain 'Slash command token value', got: %s", view)
	}
	// Should NOT show vault-mode labels
	if strings.Contains(view, "Bot token secret path") {
		t.Errorf("Step 9 dev mode view should NOT contain 'Bot token secret path'")
	}
	if strings.Contains(view, "Signing secret path") {
		t.Errorf("Step 9 dev mode view should NOT contain 'Signing secret path'")
	}
}

func TestStep9VaultModeLabels(t *testing.T) {
	// When secrets backend is vault, the Slack integration step should show
	// vault-mode labels (backward compatible)
	cfg := config.NewDefaultWizardConfig()
	cfg.SecretsBackend = "vault"
	m := NewWizard(cfg)
	m = runInitStep(m, 9)

	view := m.View()

	if !strings.Contains(view, "Bot token secret path") {
		t.Errorf("Step 9 vault mode view should contain 'Bot token secret path', got: %s", view)
	}
	if !strings.Contains(view, "Signing secret path") {
		t.Errorf("Step 9 vault mode view should contain 'Signing secret path', got: %s", view)
	}
}

func TestSummaryDevModeSecretFormat(t *testing.T) {
	// Summary screen should show dev-mode secret format hint
	cfg := config.NewDefaultWizardConfig()
	cfg.SecretsBackend = "dev"
	m := NewWizard(cfg)
	m = runInitStep(m, 15)

	view := m.View()

	if !strings.Contains(view, "Secret format: plain values / $ENV_VAR refs") {
		t.Errorf("Summary dev mode should show dev secret format, got: %s", view)
	}
	if strings.Contains(view, "LOOKUP[vault,...]") {
		t.Errorf("Summary dev mode should NOT show vault secret format")
	}
}

func TestSummaryVaultModeSecretFormat(t *testing.T) {
	// Summary screen should show vault-mode secret format hint (backward compatible)
	cfg := config.NewDefaultWizardConfig()
	cfg.SecretsBackend = "vault"
	m := NewWizard(cfg)
	m = runInitStep(m, 15)

	view := m.View()

	if !strings.Contains(view, "Secret format: LOOKUP[vault,...] paths") {
		t.Errorf("Summary vault mode should show vault secret format, got: %s", view)
	}
	if strings.Contains(view, "$ENV_VAR refs") {
		t.Errorf("Summary vault mode should NOT show dev secret format")
	}
}
