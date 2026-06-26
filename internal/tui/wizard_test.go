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
			step: 3, key: "deployment_mode", value: "kubernetes",
			check: func(c *config.WizardConfig) error {
				if c.DeploymentMode != "kubernetes" {
					return fmt.Errorf("DeploymentMode = %q", c.DeploymentMode)
				}
				return nil
			},
		},
		{
			step: 7, key: "secrets_backend", value: "dev",
			check: func(c *config.WizardConfig) error {
				if c.SecretsBackend != "dev" {
					return fmt.Errorf("SecretsBackend = %q", c.SecretsBackend)
				}
				return nil
			},
		},
		{
			step: 11, key: "ai_enabled", value: "true",
			check: func(c *config.WizardConfig) error {
				if !c.AIEnabled {
					return fmt.Errorf("AIEnabled = false")
				}
				return nil
			},
		},
		{
			step: 4, key: "use_local_copy", value: "true",
			check: func(c *config.WizardConfig) error {
				if !c.UseLocalCopy {
					return fmt.Errorf("UseLocalCopy = false")
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
	if labels[6] != "Clone Auth" {
		t.Errorf("labels[6] = %q, want Clone Auth", labels[6])
	}
	if labels[7] != "Secrets Backend" {
		t.Errorf("labels[7] = %q, want Secrets Backend", labels[7])
	}
	if labels[8] != "Redis" {
		t.Errorf("labels[8] = %q, want Redis", labels[8])
	}
	if labels[9] != "GitHub Integration" {
		t.Errorf("labels[9] = %q, want GitHub Integration", labels[9])
	}
	if labels[10] != "Slack Integration" {
		t.Errorf("labels[10] = %q, want Slack Integration", labels[10])
	}
	if labels[11] != "AI Agent" {
		t.Errorf("labels[11] = %q, want AI Agent", labels[11])
	}
}

func TestIsStepRequired(t *testing.T) {
	requiredSteps := []int{2, 3, 4, 5, 6, 7, 8, 11, 12, 13, 14}
	for _, step := range requiredSteps {
		if !IsStepRequired(step) {
			t.Errorf("step %d should be required", step)
		}
	}
	optionalSteps := []int{1, 9, 10, 15}
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
			name:    "step 3 with valid mode",
			step:    3,
			setup:   func(c *config.WizardConfig) { c.DeploymentMode = "docker" },
			wantErr: false,
		},
		{
			name:    "step 3 with invalid mode",
			step:    3,
			setup:   func(c *config.WizardConfig) { c.DeploymentMode = "invalid" },
			wantErr: true,
		},
		{
			name:    "step 7 with secrets backend selected",
			step:    7,
			setup:   func(c *config.WizardConfig) { c.SecretsBackend = "vault" },
			wantErr: false,
		},
		{
			name:    "step 7 without secrets backend",
			step:    7,
			setup:   func(c *config.WizardConfig) { c.SecretsBackend = "" },
			wantErr: true,
		},
		{
			name:    "step 4 with yes and git remote URL",
			step:    4,
			setup:   func(c *config.WizardConfig) { c.GithubCreateRepo = true; c.GitRemoteURL = "git@github.com:user/repo.git" },
			wantErr: false,
		},
		{
			name:    "step 4 with yes but no git remote URL",
			step:    4,
			setup:   func(c *config.WizardConfig) { c.GithubCreateRepo = true; c.GitRemoteURL = "" },
			wantErr: true,
		},
		{
			name:    "step 4 with yes and local copy (git remote URL still required)",
			step:    4,
			setup:   func(c *config.WizardConfig) { c.GithubCreateRepo = true; c.UseLocalCopy = true; c.GitRemoteURL = "" },
			wantErr: true,
		},
		{
			name:    "step 4 with no",
			step:    4,
			setup:   func(c *config.WizardConfig) { c.GithubCreateRepo = false },
			wantErr: false,
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

	// Step 2 should be multi-field text input (project name + config dir)
	m = runInitStep(m, 2)
	if m.mode != modeTextInput {
		t.Errorf("Step 2 mode = %d, want modeTextInput(%d)", m.mode, modeTextInput)
	}
	if len(m.textInputs) != 2 {
		t.Errorf("Step 2 textInputs count = %d, want 2", len(m.textInputs))
	}
	if !m.textInput.Focused() {
		t.Error("Step 2 text input should be focused")
	}

	// Step 4 should be radio select
	m = runInitStep(m, 3)
	if m.mode != modeRadioSelect {
		t.Errorf("Step 4 mode = %d, want modeRadioSelect(%d)", m.mode, modeRadioSelect)
	}

	// Step 5 should be radio select (secure-exec)
	m = runInitStep(m, 5)
	if m.mode != modeRadioSelect {
		t.Errorf("Step 5 mode = %d, want modeRadioSelect(%d)", m.mode, modeRadioSelect)
	}

	// Step 7 should be radio select (Secrets Backend)
	m = runInitStep(m, 7)
	if m.mode != modeRadioSelect {
		t.Errorf("Step 7 mode = %d, want modeRadioSelect(%d)", m.mode, modeRadioSelect)
	}

	// Step 8 should be checkbox select (GitHub Integration)
	m = runInitStep(m, 9)
	if m.mode != modeCheckboxSelect {
		t.Errorf("Step 8 mode = %d, want modeCheckboxSelect(%d)", m.mode, modeCheckboxSelect)
	}

	// Step 9 should be multi-field text input (Slack Integration)
	m = runInitStep(m, 10)
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

	// Press Enter to move to next field (config dir)
	m, _ = updateWizard(m, tea.KeyMsg{Type:tea.KeyEnter})

	// Config should be updated
	if m.config.ProjectName != "test-proj" {
		t.Errorf("ProjectName = %q, want %q", m.config.ProjectName, "test-proj")
	}
	// Should still be on step 2, field 1 (config dir)
	if m.step != 2 {
		t.Errorf("step = %d, want 2", m.step)
	}
	if m.currentField != 1 {
		t.Errorf("currentField = %d, want 1", m.currentField)
	}

	// Press Enter again to advance to step 3
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != 3 {
		t.Errorf("step = %d, want 3", m.step)
	}
}

func TestMultiFieldTabNavigation(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 2) // Project Settings - 2 fields (project name + config dir)

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
	m = runInitStep(m, 2) // Project Settings - 2 fields

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
	m = runInitStep(m, 2)

	// Press Esc on first field - should go to previous step (welcome)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})

	if m.step != 1 {
		t.Errorf("step = %d, want 1 (previous step)", m.step)
	}
}

func TestRadioSelectionDownArrow(t *testing.T) {
	m := NewWizard(nil)
	m = runInitStep(m, 3) // Deployment Mode

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
	m = runInitStep(m, 3)

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
	m = runInitStep(m, 3)

	// Move to "source" (index 1)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})

	// Press Enter to commit
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	// Config should be updated
	if m.config.DeploymentMode != "source" {
		t.Errorf("DeploymentMode = %q, want %q", m.config.DeploymentMode, "source")
	}
	// Should advance to next step
	if m.step != 4 {
		t.Errorf("step = %d, want 4", m.step)
	}
}

func TestRadioSelectionEscGoesBack(t *testing.T) {
	m := NewWizard(nil)
	m = runInitStep(m, 3)

	// Press Esc to go back
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})

	if m.step != 2 {
		t.Errorf("step = %d, want 2", m.step)
	}
}

func TestRadioSelectionBackspaceGoesBack(t *testing.T) {
	m := NewWizard(nil)
	m = runInitStep(m, 3)

	// Press Backspace to go back
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyBackspace})

	if m.step != 2 {
		t.Errorf("step = %d, want 2", m.step)
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
	m = runInitStep(m, 2) // Project Settings (project name + config dir)

	if m.textInputs[0].Value() != "hd-config" {
		t.Errorf("field 0 = %q, want %q", m.textInputs[0].Value(), "hd-config")
	}
	if m.textInputs[1].Value() != "./hd-config" {
		t.Errorf("field 1 = %q, want %q", m.textInputs[1].Value(), "./hd-config")
	}
}

func TestRadioDefaultSelection(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 3) // Deployment Mode

	// Default is "docker" which is index 0
	if m.radioIndex != 0 {
		t.Errorf("radioIndex = %d, want 0 (docker)", m.radioIndex)
	}
}

func TestViewContainsStepTitle(t *testing.T) {
	m := NewWizard(nil)
	m = runInitStep(m, 2)

	view := m.View()
	if !strings.Contains(view, "Project name") {
		t.Errorf("View() should contain step title 'Project name', got: %s", view)
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
	m = runInitStep(m, 3)

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
	// Ctrl+Q on step 11 should quit without generating config
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 15)

	// Press ctrl+q - should quit without generating
	_, cmd := updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyCtrlQ})

	if cmd == nil {
		t.Error("ctrl+q on step 11 should produce a quit command")
	}
	// err should be nil (no config generation)
	if m.err != nil {
		t.Errorf("err = %v, want nil (no config generation on quit)", m.err)
	}
}

func TestViewSummaryScreen(t *testing.T) {
	// The summary is shown on step 11 (navigate mode), not on a done screen.
	m := NewWizard(nil)
	m = runInitStep(m, 15)

	view := m.View()
	// Step 16 should show the summary
	if !strings.Contains(view, "Configuration Summary") {
		t.Errorf("Step 16 View() should contain 'Configuration Summary', got: %s", view)
	}
	// Step 16 should show the generate prompt
	if !strings.Contains(view, "Press Enter to generate configs") {
		t.Errorf("Step 16 View() should contain 'Press Enter to generate configs', got: %s", view)
	}
}

func TestStep15ShowsSummary(t *testing.T) {
	// Step 16 should show the summary directly
	m := NewWizard(nil)
	m = runInitStep(m, 15)

	view := m.View()
	if !strings.Contains(view, "Configuration Summary") {
		t.Errorf("Step 16 View() should contain 'Configuration Summary', got: %s", view)
	}
	if !strings.Contains(view, "Project:") {
		t.Errorf("Step 16 View() should contain 'Project:' in summary, got: %s", view)
	}
	// Verify new order in summary
	if !strings.Contains(view, "Secrets backend:") {
		t.Errorf("Step 16 View() should contain 'Secrets backend:' in summary, got: %s", view)
	}
	if !strings.Contains(view, "Redis:") {
		t.Errorf("Step 16 View() should contain 'Redis:' in summary, got: %s", view)
	}
	// Step 16 should also show the generate prompt
	if !strings.Contains(view, "Press Enter to generate configs") {
		t.Errorf("Step 16 View() should contain 'Press Enter to generate configs', got: %s", view)
	}
	// Step 16 should show the save hint (from renderNavigation)
	if !strings.Contains(view, "ctrl+s=save") {
		t.Errorf("Step 16 View() should contain 'ctrl+s=save' hint, got: %s", view)
	}
}

func TestStep16EnterGeneratesDirectly(t *testing.T) {
	// Fix 1: Pressing Enter on step 11 should generate config directly
	// without going through the done/confirmation screen.
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 15) // Summary step (navigate mode)

	// Press Enter - should trigger generateConfig and return quit
	m, cmd := updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Error("expected quit command after enter on step 11")
	}
	// Should NOT be in done state (direct generation, no intermediate done)
	if m.done {
		t.Error("model should NOT be in done state")
	}
}

func TestStep16SaveAndGenerateDirectly(t *testing.T) {
	// Fix 1: Pressing 's' on step 11 should save answers and generate directly.
	cfg := config.NewDefaultWizardConfig()
	cfg.ProjectName = "save-gen-test"
	m := NewWizard(cfg)
	m = runInitStep(m, 15)

	// Press 's' - should save and generate
	m, cmd := updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})

	if cmd == nil {
		t.Error("expected quit command after 's' on step 11")
	}
	if m.done {
		t.Error("model should NOT be in done state")
	}
}

func TestStep15EscGoesBack(t *testing.T) {
	// Pressing Esc on step 11 should go back to previous step
	m := NewWizard(nil)
	m = runInitStep(m, 6)

	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})

	if m.step == 15 {
		t.Errorf("step = %d, should have gone back from step 11", m.step)
	}
}

func TestDoneScreenEnterConfirms(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 15)

	// Press Enter - should trigger generateConfig and return a quit command
	m, cmd := updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Error("expected quit command after enter on step 11")
	}
	// Should NOT be in done state (direct generation, no intermediate done)
	if m.done {
		t.Error("model should NOT be in done state")
	}
}

func TestDoneScreenEscGoesBack(t *testing.T) {
	m := NewWizard(nil)
	m = runInitStep(m, 15)

	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})

	// Esc on step 15 (summary) should go back to step 12 (docker mode)
	if m.step != 12 {
		t.Errorf("should be on step 12 after esc on step 15, got %d", m.step)
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
		{2, "Project Settings"},
		{3, "Deployment Mode"},
		{4, "GitHub Repo Creation"},
		{5, "Secure Execution"},
		{6, "Clone Auth"},
		{7, "Secrets Backend"},
		{8, "Redis"},
		{9, "GitHub Integration"},
		{10, "Slack Integration"},
		{11, "AI Agent"},
		{12, "Docker Configuration"},
		{13, "Kubernetes Configuration"},
		{14, "Source Configuration"},
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
	m = runInitStep(m, 3) // Deployment Mode (radio)

	// Press Ctrl+S
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlS})

	// Should show save confirmation
	view := m.View()
	if !strings.Contains(view, "Answers saved to") {
		t.Errorf("View() should contain save confirmation, got: %s", view)
	}
	// Should still be on step 3 (Ctrl+S saves but does not advance)
	if m.step != 3 {
		t.Errorf("step = %d, want 3", m.step)
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
	m = runInitStep(m, 9) // GitHub Integration (checkbox)

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
	// Pressing ctrl+q on step 11 should quit without generating config
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 15)

	// Press ctrl+q - should quit without generating
	_, cmd := updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyCtrlQ})

	if cmd == nil {
		t.Error("ctrl+q on step 11 should produce a quit command")
	}
	// err should be nil (no config generation)
	if m.err != nil {
		t.Errorf("err = %v, want nil (no config generation on quit)", m.err)
	}
}

func TestCtrlQOnRadioStepQuits(t *testing.T) {
	// Pressing ctrl+q on a radio select step should quit immediately
	m := NewWizard(nil)
	m = runInitStep(m, 3) // Deployment Mode (radio select)

	_, cmd := updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyCtrlQ})

	if cmd == nil {
		t.Error("ctrl+q on radio step should produce a quit command")
	}
}

func TestCtrlQOnCheckboxStepQuits(t *testing.T) {
	// Pressing ctrl+q on a checkbox step should quit immediately
	m := NewWizard(nil)
	m = runInitStep(m, 9) // GitHub Integration (checkbox)

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
	m = runInitStep(m, 9) // GitHub Integration (checkbox)

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
	// Should still be on step 9 (q is a regular character in text input)
	if m.step != 9 {
		t.Errorf("step = %d, want 9 (should not advance)", m.step)
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
	m = runInitStep(m, 3) // radio select

	m, _ = updateWizardCmd(m, tea.KeyMsg{Type: tea.KeyCtrlQ})

	if !m.quit {
		t.Error("quit flag should be true after pressing ctrl+q on radio step")
	}
}

func TestQuitSetsQuitFlagOnCheckboxStep(t *testing.T) {
	m := NewWizard(nil)
	m = runInitStep(m, 9) // checkbox

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
	// When user completes the wizard (presses Enter on step 11),
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
	if placeholder != "Plain text or $HD_ENV_VAR" {
		t.Errorf("dev key placeholder = %q, want %q", placeholder, "Plain text or $HD_ENV_VAR")
	}
	if help != "GitHub App private key value or env var reference (only for github_app integration)" {
		t.Errorf("dev key help = %q, want value help", help)
	}

	label, placeholder, _ = ghSecretLabels(cfg, "token")
	if label != "Token value" {
		t.Errorf("dev token label = %q, want %q", label, "Token value")
	}
	if placeholder != "Plain text or $HD_ENV_VAR" {
		t.Errorf("dev token placeholder = %q, want %q", placeholder, "Plain text or $HD_ENV_VAR")
	}

	label, placeholder, _ = ghSecretLabels(cfg, "webhook")
	if label != "Webhook secret value" {
		t.Errorf("dev webhook label = %q, want %q", label, "Webhook secret value")
	}
	if placeholder != "Plain text or $HD_ENV_VAR" {
		t.Errorf("dev webhook placeholder = %q, want %q", placeholder, "Plain text or $HD_ENV_VAR")
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
	if placeholder != "Plain text or $HD_ENV_VAR" {
		t.Errorf("dev bot_token placeholder = %q, want %q", placeholder, "Plain text or $HD_ENV_VAR")
	}

	label, _, _ = slackSecretLabels(cfg, "signing")
	if label != "Signing secret value" {
		t.Errorf("dev signing label = %q, want %q", label, "Signing secret value")
	}

	label, placeholder, _ = slackSecretLabels(cfg, "interaction")
	if label != "Interaction token value" {
		t.Errorf("dev interaction label = %q, want %q", label, "Interaction token value")
	}
	if placeholder != "Plain text or $HD_ENV_VAR (optional)" {
		t.Errorf("dev interaction placeholder = %q, want %q", placeholder, "Plain text or $HD_ENV_VAR (optional)")
	}

	label, placeholder, _ = slackSecretLabels(cfg, "slash")
	if label != "Slash command token value" {
		t.Errorf("dev slash label = %q, want %q", label, "Slash command token value")
	}
	if placeholder != "Plain text or $HD_ENV_VAR (optional)" {
		t.Errorf("dev slash placeholder = %q, want %q", placeholder, "Plain text or $HD_ENV_VAR (optional)")
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
	m = runInitStep(m, 9)

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
	m = runInitStep(m, 9)

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
	m = runInitStep(m, 10)

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
	m = runInitStep(m, 10)

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

	if !strings.Contains(view, "Secret format: plain values / $HD_* refs") {
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
	if strings.Contains(view, "$HD_* refs") {
		t.Errorf("Summary vault mode should NOT show dev secret format")
	}
}

// ===== Dev mode $HD_* env var validation tests =====

func TestValidateDevSecretValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		wantErr  string // empty string means no error
	}{
		{"empty value", "", ""},
		{"plain value", "my-secret-value", ""},
		{"valid HD_ ref", "$HD_GITHUB_TOKEN", ""},
		{"valid HD_ ref with suffix", "$HD_SLACK_BOT_TOKEN", ""},
		{"invalid non-HD ref", "$GITHUB_TOKEN", "must use $HD_ prefix"},
		{"invalid non-HD ref SLACK", "$SLACK_TOKEN", "must use $HD_ prefix"},
		{"just dollar sign", "$", "must use $HD_ prefix"},
		{"dollar with lowercase", "$hd_something", "must use $HD_ prefix"},
		{"HD_ without dollar prefix", "HD_GITHUB_TOKEN", ""}, // plain value, no $ prefix
		{"plain with spaces", "  my value  ", ""},            // plain value with spaces
		{"HD_ ref with spaces", "  $HD_GITHUB_TOKEN  ", ""},  // trimmed and valid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateDevSecretValue(tt.value)
			if tt.wantErr == "" {
				if got != "" {
					t.Errorf("validateDevSecretValue(%q) = %q, want no error", tt.value, got)
				}
			} else {
				if !strings.Contains(got, tt.wantErr) {
					t.Errorf("validateDevSecretValue(%q) = %q, want error containing %q", tt.value, got, tt.wantErr)
				}
			}
		})
	}
}

func TestStripEnvVarPrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"$HD_GITHUB_TOKEN", "HD_GITHUB_TOKEN"},
		{"$HD_SLACK_BOT_TOKEN", "HD_SLACK_BOT_TOKEN"},
		{"HD_GITHUB_TOKEN", "HD_GITHUB_TOKEN"}, // no $ prefix, unchanged
		{"plain-value", "plain-value"},          // no $ prefix, unchanged
		{"$GITHUB_TOKEN", "GITHUB_TOKEN"},      // strips $ even if not HD_
		{"", ""},                                // empty string
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := stripEnvVarPrefix(tt.input)
			if got != tt.expected {
				t.Errorf("stripEnvVarPrefix(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsDevSecretField(t *testing.T) {
	tests := []struct {
		step     int
		label    string
		expected bool
	}{
		{9, "Private key secret path", true},
		{9, "Private key value", true},
		{9, "Token secret path", true},
		{9, "Token value", true},
		{9, "Webhook secret path", true},
		{9, "Webhook secret value", true},
		{9, "App ID", false},
		{9, "Installation ID", false},
		{10, "Bot token secret path", true},
		{10, "Bot token value", true},
		{10, "Signing secret path", true},
		{10, "Signing secret value", true},
		{10, "Interaction token", true},
		{10, "Interaction token value", true},
		{10, "Slash command token", true},
		{10, "Slash command token value", true},
		{2, "Project name", false},
		{11, "API key secret path", false}, // AI step, not a dev secret field
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("step%d_%s", tt.step, tt.label), func(t *testing.T) {
			got := isDevSecretField(tt.step, tt.label)
			if got != tt.expected {
				t.Errorf("isDevSecretField(%d, %q) = %v, want %v", tt.step, tt.label, got, tt.expected)
			}
		})
	}
}

func TestDevModeSecretValidationRejectsNonHD(t *testing.T) {
	// When in dev mode, entering a non-HD_ env var ref should be rejected
	cfg := config.NewDefaultWizardConfig()
	cfg.ProjectName = "test-validation"
	cfg.SecretsBackend = "dev"
	cfg.HasGitHubAppIntegration = true
	m := NewWizard(cfg)
	m = runInitStep(m, 9)

	// Navigate from checkboxes to text fields:
	// Tab from checkbox 0 to checkbox 1, then Tab to first text field
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	// Now in first text field (App ID) — type a value
	for _, ch := range "12345" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	// Now in second text field (Installation ID) — type a value
	for _, ch := range "67890" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	// Now in third text field (Private key) — type a non-HD_ env var ref
	for _, ch := range "$GITHUB_TOKEN" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	// Now in fourth text field (Token) — type a value (required for PAT)
	for _, ch := range "$HD_TOKEN" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	// Now in fifth text field (Webhook) — type a value
	for _, ch := range "$HD_WEBHOOK" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}

	// Press Enter on last field — should fail validation for non-HD_ value
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	// Should show validation error
	if m.validationErr == "" {
		t.Error("expected validation error for non-HD_ env var ref, got none")
	}
	if !strings.Contains(m.validationErr, "HD_") {
		t.Errorf("validation error should mention HD_ prefix, got: %q", m.validationErr)
	}
	// Should still be on step 9
	if m.step != 9 {
		t.Errorf("step = %d, want 9 (should not advance on validation error)", m.step)
	}
}
func TestDevModeSecretValidationAcceptsHD(t *testing.T) {
	// When in dev mode, entering a valid HD_ env var ref should be accepted
	cfg := config.NewDefaultWizardConfig()
	cfg.ProjectName = "test-validation-hd"
	cfg.SecretsBackend = "dev"
	cfg.HasGitHubAppIntegration = true
	m := NewWizard(cfg)
	m = runInitStep(m, 9)

	// Navigate from checkboxes to text fields
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	// Now in first text field (App ID) — type a value
	for _, ch := range "12345" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	// Now in second text field (Installation ID) — type a value
	for _, ch := range "67890" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	// Now in third text field (Private key) — type a valid HD_ env var ref
	for _, ch := range "$HD_GITHUB_TOKEN" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}

	// Press Enter to advance — should succeed (no validation error)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	// Should NOT have a validation error about HD_ prefix
	if m.validationErr != "" && strings.Contains(m.validationErr, "HD_") {
		t.Errorf("should not get HD_ validation error for valid ref, got: %q", m.validationErr)
	}
	// The stored value should include the $ prefix (stripped only during template rendering)
	if m.config.GithubKeyPath != "$HD_GITHUB_TOKEN" {
		t.Errorf("GithubKeyPath = %q, want %q", m.config.GithubKeyPath, "$HD_GITHUB_TOKEN")
	}
}

func TestDevModeSecretAcceptsPlainValue(t *testing.T) {
	// When in dev mode, entering a plain value (no $) should be accepted
	cfg := config.NewDefaultWizardConfig()
	cfg.ProjectName = "test-plain"
	cfg.SecretsBackend = "dev"
	cfg.HasGitHubAppIntegration = true
	m := NewWizard(cfg)
	m = runInitStep(m, 9)

	// Navigate from checkboxes to text fields
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	// Now in first text field (App ID) — type a value
	for _, ch := range "12345" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	// Now in second text field (Installation ID) — type a value
	for _, ch := range "67890" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	// Now in third text field (Private key) — type a plain value
	for _, ch := range "my-private-key-content" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	// Now in fourth text field (Token) — type a value (required for PAT)
	for _, ch := range "$HD_TOKEN" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	// Now in fifth text field (Webhook) — type a value
	for _, ch := range "$HD_WEBHOOK" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}

	// Press Enter on last field — should succeed
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	// Should NOT have a validation error
	if m.validationErr != "" {
		t.Errorf("should not get validation error for plain value, got: %q", m.validationErr)
	}
	// The stored value should be the plain value as-is
	if m.config.GithubKeyPath != "my-private-key-content" {
		t.Errorf("GithubKeyPath = %q, want %q", m.config.GithubKeyPath, "my-private-key-content")
	}
}

func TestVaultModeSecretNoValidation(t *testing.T) {
	// In vault mode, no HD_ validation should occur
	cfg := config.NewDefaultWizardConfig()
	cfg.ProjectName = "test-vault-no-val"
	cfg.SecretsBackend = "vault"
	cfg.HasGitHubAppIntegration = true
	m := NewWizard(cfg)
	m = runInitStep(m, 9)

	// Navigate from checkboxes to text fields
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	// Now in first text field (App ID) — type a value
	for _, ch := range "12345" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	// Now in second text field (Installation ID) — type a value
	for _, ch := range "67890" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	// Now in third text field (Private key) — type a vault path
	for _, ch := range "secrets/github/key" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	// Now in fourth text field (Token) — type a value (required for PAT)
	for _, ch := range "secrets/token" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	// Now in fifth text field (Webhook) — type a value
	for _, ch := range "secrets/webhook" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}

	// Press Enter on last field — should succeed
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	// Should NOT have a validation error
	if m.validationErr != "" {
		t.Errorf("should not get validation error in vault mode, got: %q", m.validationErr)
	}
	if m.config.GithubKeyPath != "secrets/github/key" {
		t.Errorf("GithubKeyPath = %q, want %q", m.config.GithubKeyPath, "secrets/github/key")
	}
}


func TestStep14CheckboxNotCheckedByDefault(t *testing.T) {
	// Step 14 starts with both checkboxes unchecked
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	if m.config.GithubCreateRepo {
		t.Error("GithubCreateRepo should be false by default")
	}
	if m.config.UseLocalCopy {
		t.Error("UseLocalCopy should be false by default")
	}
	// No conditional text fields when no checkboxes are checked
	if len(m.textInputs) != 0 {
		t.Errorf("Step 14 default textInputs count = %d, want 0", len(m.textInputs))
	}
}

func TestStep14SpaceToggleFirstCheckbox(t *testing.T) {
	// Pressing Space on the first checkbox should toggle GithubCreateRepo
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// First checkbox ("Create GitHub repo") is at checkboxIndex 0
	if m.checkboxIndex != 0 {
		t.Fatalf("checkboxIndex = %d, want 0", m.checkboxIndex)
	}
	if m.currentField != 0 {
		t.Fatalf("currentField = %d, want 0", m.currentField)
	}

	// Press Space to toggle "Create GitHub repo" on
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeySpace})

	if !m.config.GithubCreateRepo {
		t.Error("GithubCreateRepo should be true after pressing space on first checkbox")
	}

	// Now the "Use local copy" checkbox should appear (condition: GithubCreateRepo == true)
	view := m.View()
	if !strings.Contains(view, "Use local copy instead of clone") {
		t.Errorf("step 4 view should contain 'Use local copy' checkbox, got:\n%s", view)
	}

	// Text inputs should now include git remote URL
	if len(m.textInputs) != 1 {
		t.Errorf("textInputs count = %d, want 1 (git remote URL)", len(m.textInputs))
	}
}

func TestStep14SecondCheckboxVisibleWhenFirstChecked(t *testing.T) {
	// After checking "Create GitHub repo", the second checkbox should be visible
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// Toggle first checkbox on
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeySpace})
	if !m.config.GithubCreateRepo {
		t.Fatal("GithubCreateRepo should be true")
	}

	// Navigate to second checkbox
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.checkboxIndex != 1 {
		t.Errorf("checkboxIndex = %d, want 1", m.checkboxIndex)
	}

	// Toggle "Use local copy" on
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeySpace})
	if !m.config.UseLocalCopy {
		t.Error("UseLocalCopy should be true after toggling second checkbox")
	}

	// Git remote URL should still be visible, but clone auth fields should be hidden
	if len(m.textInputs) != 1 {
		t.Errorf("textInputs count = %d, want 1 (git remote URL visible, secure-exec hidden)", len(m.textInputs))
	}
}

func TestStep14SecondCheckboxHiddenWhenFirstUnchecked(t *testing.T) {
	// When "Create GitHub repo" is unchecked, "Use local copy" should be hidden
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// Toggle first checkbox on, then off
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeySpace}) // on
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeySpace}) // off

	if m.config.GithubCreateRepo {
		t.Error("GithubCreateRepo should be false after toggling off")
	}

	// View should NOT contain "Use local copy"
	view := m.View()
	if strings.Contains(view, "Use local copy instead of clone") {
		t.Errorf("step 4 view should NOT contain 'Use local copy' when GithubCreateRepo is false, got:\n%s", view)
	}

	// No text inputs
	if len(m.textInputs) != 0 {
		t.Errorf("textInputs count = %d, want 0", len(m.textInputs))
	}
}

func TestStep14EnterTogglesAndAdvances(t *testing.T) {
	// Pressing Enter on a checkbox toggles it and advances to the next one
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// Press Enter on first checkbox: toggles GithubCreateRepo on, moves to checkbox 1
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	if !m.config.GithubCreateRepo {
		t.Error("GithubCreateRepo should be true after Enter on first checkbox")
	}
	if m.checkboxIndex != 1 {
		t.Errorf("checkboxIndex = %d, want 1 (Enter should advance to next checkbox)", m.checkboxIndex)
	}
	if m.currentField != 0 {
		t.Errorf("currentField = %d, want 0", m.currentField)
	}
}

func TestStep14EnterOnLastCheckboxWithTextInput(t *testing.T) {
	// Pressing Enter on the last checkbox when text inputs exist switches to text input mode
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// Toggle first checkbox on (Space), move to second checkbox
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeySpace}) // toggle GithubCreateRepo on
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})   // move to checkbox 1

	// textInputs should have git remote URL + secure-exec driver
	if len(m.textInputs) != 1 {
		t.Fatalf("textInputs count = %d, want 1", len(m.textInputs))
	}

	// Press Enter on last checkbox: toggles UseLocalCopy on, textInputs still exist
	// (git remote URL is always visible when GithubCreateRepo is true,
	// and secure-exec driver is visible when !UseLocalCopy)
	// When UseLocalCopy is toggled on, secure-exec driver becomes hidden,
	// leaving only 1 text input (git remote URL)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	// UseLocalCopy should be toggled on
	if !m.config.UseLocalCopy {
		t.Error("UseLocalCopy should be true after Enter on second checkbox")
	}

	// Text inputs still exist (git remote URL always visible), so should switch to text input mode
	if m.step != 4 {
		t.Errorf("step = %d, want 4 (should stay on step 4 with text inputs)", m.step)
	}
	if m.currentField != 1 {
		t.Errorf("currentField = %d, want 1 (should focus text input)", m.currentField)
	}
}

func TestStep14EnterOnLastCheckboxAdvancesWhenNoTextInputs(t *testing.T) {
	// When text inputs exist, pressing Enter on last checkbox switches to text input mode
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// Press Enter twice: toggle GithubCreateRepo (move to cb1), toggle UseLocalCopy
	// After both toggles, git remote URL text input still exists (always visible when
	// GithubCreateRepo is true), so Enter switches to text input mode
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter}) // toggle GithubCreateRepo, move to cb1
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter}) // toggle UseLocalCopy, switch to text input

	if m.step != 4 {
		t.Errorf("step = %d, want 4 (should stay on step 4 with text inputs)", m.step)
	}
	if m.currentField != 1 {
		t.Errorf("currentField = %d, want 1 (should focus text input)", m.currentField)
	}
}

func TestStep14EnterOnLastCheckboxNoAdvanceWhenTextInputRequired(t *testing.T) {
	// When git remote URL is required but empty, pressing Enter on last checkbox
	// switches to text input mode (not advancing), so validation happens when user
	// tries to advance from text input mode
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// Toggle GithubCreateRepo on
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeySpace})

	// Now there are text inputs (git remote URL + secure-exec driver)
	if len(m.textInputs) != 1 {
		t.Fatalf("textInputs count = %d, want 1", len(m.textInputs))
	}

	// Move to last checkbox (index 1)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})

	// Press Enter on last checkbox: toggles UseLocalCopy, text inputs still exist
	// (git remote URL always visible), so switches to text input mode
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.step != 4 {
		t.Errorf("step = %d, want 4 (should stay on step 4 with text inputs)", m.step)
	}
	if m.currentField != 1 {
		t.Errorf("currentField = %d, want 1 (should focus text input)", m.currentField)
	}
}

func TestStep14GitRemoteURLFieldVisibleAfterCheckingCreateRepo(t *testing.T) {
	// After checking "Create GitHub repo", git remote URL and clone auth fields should appear
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// Toggle "Create GitHub repo" on
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeySpace})

	// Text inputs should have 2 fields (git remote URL + secure-exec driver)
	if len(m.textInputs) != 1 {
		t.Errorf("textInputs count = %d, want 1", len(m.textInputs))
	}

	view := m.View()
	if !strings.Contains(view, "Git remote URL") {
		t.Errorf("view should contain 'Git remote URL', got:\n%s", view)
	}
}

func TestStep14CheckboxVisibleInTextInputMode(t *testing.T) {
	// When in text input mode, checkboxes should still be visible above the text fields
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// Toggle "Create GitHub repo" on, move to second checkbox, press Enter to switch to text input
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeySpace}) // toggle GithubCreateRepo on
	// Now text inputs exist (git remote URL + secure-exec driver)
	if len(m.textInputs) != 1 {
		t.Fatalf("textInputs count = %d, want 1", len(m.textInputs))
	}

	// Press Down to move to checkbox 1, then Down again to move to text inputs
	// Down from last checkbox goes to textInputs[0] at currentField=1
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown}) // to checkbox 1
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown}) // to text inputs (currentField=1)

	if m.currentField != 1 {
		t.Fatalf("currentField = %d, want 1", m.currentField)
	}

	// View should still contain checkboxes AND the git remote URL field
	view := m.View()
	if !strings.Contains(view, "Create GitHub repo") {
		t.Errorf("view should contain 'Create GitHub repo' checkbox in text input mode, got:\n%s", view)
	}
	if !strings.Contains(view, "Git remote URL") {
		t.Errorf("view should contain 'Git remote URL' in text input mode, got:\n%s", view)
	}
}

func TestStep14TextInputModeTabNavigation(t *testing.T) {
	// Tab should navigate through text fields in checkbox step text input mode
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// Toggle "Create GitHub repo" on
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeySpace})

	// There should be 2 text inputs (git remote URL + secure-exec driver)
	if len(m.textInputs) != 1 {
		t.Fatalf("textInputs count = %d, want 1", len(m.textInputs))
	}

	// Move to text input: Down past last checkbox goes to textInputs[0] (Git remote URL)
	m, _ = updateWizard(m, tea.KeyMsg{Type:tea.KeyDown}) // to checkbox 1
	m, _ = updateWizard(m, tea.KeyMsg{Type:tea.KeyDown}) // to text input (Git remote URL)

	// Type in the git remote URL
	for _, ch := range "git@github.com:user/repo.git" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}

	// Press Enter on the text input - should advance to step 5
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != 5 {
		t.Errorf("step = %d, want 5", m.step)
	}
	if m.validationErr != "" {
		t.Errorf("expected no validation error, got: %s", m.validationErr)
	}
}
func TestStep14TextInputEscBackToCheckboxes(t *testing.T) {
	// Pressing Esc on first text field should go back to checkboxes
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// Toggle "Create GitHub repo" on
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeySpace})

	// Move to text input
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown}) // to checkbox 1
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown}) // to text input

	if m.currentField != 1 {
		t.Fatalf("currentField = %d, want 1", m.currentField)
	}

	// Esc should go back to checkboxes
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})

	if m.currentField != 0 {
		t.Errorf("currentField = %d, want 0 (should go back to checkboxes)", m.currentField)
	}
}

func TestStep14NoCheckboxesNoAdvance(t *testing.T) {
	// With no checkboxes checked, pressing Enter on the last checkbox should toggle it and advance
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// Initially only 1 visible checkbox: "Create GitHub repo" (checkbox 0)
	// "Use local copy" is hidden because GithubCreateRepo is false
	if m.visibleCheckboxCount(getStepInfo(4)) != 1 {
		t.Fatalf("visibleCheckboxCount = %d, want 1", m.visibleCheckboxCount(getStepInfo(4)))
	}

	// Press Enter: toggles GithubCreateRepo on, moves to next checkbox
	// But now GithubCreateRepo is true, so "Use local copy" becomes visible
	// visibleCheckboxCount becomes 2, and checkboxIndex stays at 0
	// Actually: Enter toggles first checkbox, then tries to advance.
	// visibleCount was 1 before toggle, so checkboxIndex < visibleCount-1 is false
	// Then it checks textInputs — GithubCreateRepo is true, so 1 text input exists
	// So it should switch to text input mode

	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	// After toggle, GithubCreateRepo is true, and the step should have switched to text input mode
	// because there are text inputs
	if !m.config.GithubCreateRepo {
		t.Error("GithubCreateRepo should be true")
	}
	// Wait, Enter in checkbox mode doesn't change mode to modeTextInput.
	// It stays in modeCheckboxSelect. The mode only changes for radio steps.
	// Let me re-read the code...

	// Actually, in handleCheckboxSelect, Enter on checkbox:
	// 1. Toggles the checkbox
	// 2. visibleCount is recalculated AFTER the toggle
	// 3. if checkboxIndex < visibleCount-1 -> checkboxIndex++
	//    else if len(m.textInputs) > 0 -> currentField = 1, focus text
	//    else -> advance step

	// After toggle: GithubCreateRepo=true, UseLocalCopy=false
	// visibleCount = 2 (both checkboxes visible)
	// checkboxIndex was 0, visibleCount-1 = 1, so 0 < 1 -> checkboxIndex becomes 1
	// Mode stays modeCheckboxSelect

	if m.checkboxIndex != 1 {
		t.Errorf("checkboxIndex = %d, want 1 (Enter toggles and advances to next checkbox)", m.checkboxIndex)
	}
	if m.mode != modeCheckboxSelect {
		t.Errorf("mode = %d, want modeCheckboxSelect(%d)", m.mode, modeCheckboxSelect)
	}
}

func TestStep14EnterOnNoCheckboxesAdvances(t *testing.T) {
	// When no checkboxes are checked, pressing Enter toggles through both checkboxes
	// and switches to text input mode (git remote URL always visible when GithubCreateRepo)
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// Default: GithubCreateRepo=false, only 1 visible checkbox
	// Press Enter on checkbox 0: toggles GithubCreateRepo on, moves to cb1
	// Press Enter on checkbox 1: toggles UseLocalCopy on, textInputs still exist
	// (git remote URL always visible), switches to text input mode

	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter}) // toggle GithubCreateRepo, move to cb1
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter}) // toggle UseLocalCopy, switch to text input

	// Should stay on step 4 with text input focused
	if m.step != 4 {
		t.Errorf("step = %d, want 4 (should stay on step 4 with text inputs)", m.step)
	}
	if m.currentField != 1 {
		t.Errorf("currentField = %d, want 1 (should focus text input)", m.currentField)
	}
}

func TestStep14EscOnFirstCheckboxGoesBack(t *testing.T) {
	// Pressing Esc with no checkboxes checked should go back to previous step
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// Press Esc
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})

	// Should go back to previous required step (step 3)
	if m.step != 3 {
		t.Errorf("step = %d, want 3", m.step)
	}
}

func TestStep14GitRemoteURLRequiredWhenCreateRepoAndNoLocalCopy(t *testing.T) {
	// When GithubCreateRepo is true (regardless of UseLocalCopy), git remote URL is required
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// Toggle "Create GitHub repo" on
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeySpace})

	// Move to text input field (Git remote URL is the first text input)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown}) // to checkbox 1
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown}) // to text input (Git remote URL)

	// Press Enter on the text input - should trigger validation error
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	// Should NOT advance - should show validation error for empty git remote URL
	if m.step != 4 {
		t.Errorf("step = %d, want 4 (should not advance with empty git remote URL)", m.step)
	}
	if m.validationErr == "" {
		t.Error("expected validation error for empty git remote URL")
	}
	if !strings.Contains(m.validationErr, "git remote URL is required") {
		t.Errorf("validation error should mention git remote URL, got: %s", m.validationErr)
	}
}
func TestStep14GitRemoteURLValidationPasses(t *testing.T) {
	// When git remote URL is filled, should advance
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// Toggle "Create GitHub repo" on
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeySpace})

	// Move to text input
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown}) // to checkbox 1
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown}) // to text input (Git remote URL)

	// Type in the git remote URL
	for _, ch := range "git@github.com:user/repo.git" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}

	// Press Enter — should advance to step 5
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != 5 {
		t.Errorf("step = %d, want 5", m.step)
	}
	if m.validationErr != "" {
		t.Errorf("expected no validation error, got: %s", m.validationErr)
	}
}
func TestStep14UseLocalCopySpaceToggles(t *testing.T) {
	// Pressing space on the "Use local copy" checkbox should toggle UseLocalCopy
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// Toggle "Create GitHub repo" on
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeySpace})

	// Move to second checkbox
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})

	if m.config.UseLocalCopy {
		t.Fatal("UseLocalCopy should be false initially")
	}

	// Press space to toggle UseLocalCopy
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeySpace})

	if !m.config.UseLocalCopy {
		t.Error("UseLocalCopy should be true after pressing space")
	}

	// Press space again to toggle back
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeySpace})

	if m.config.UseLocalCopy {
		t.Error("UseLocalCopy should be false after pressing space again")
	}
}

func TestStep14UseLocalCopyHidesGitRemoteURL(t *testing.T) {
	// When UseLocalCopy is checked, the git remote URL field should still be visible
	// (UseLocalCopy only affects the REPO env var in docker-compose, not the git remote URL field)
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// Toggle "Create GitHub repo" on
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeySpace})

	// Before toggling local copy: textInputs should have 2 fields (git remote URL + secure-exec driver)
	if len(m.textInputs) != 1 {
		t.Fatalf("expected 2 text inputs before local copy (git remote URL + secure-exec driver), got %d", len(m.textInputs))
	}

	// Move to second checkbox and toggle UseLocalCopy on
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})
	m, _ = updateWizard(m, tea.KeyMsg{Type:tea.KeySpace})

	// After checking local copy: secure-exec driver is hidden (!UseLocalCopy condition fails)
	// Git remote URL is still visible (no condition on UseLocalCopy)
	if len(m.textInputs) != 1 {
		t.Errorf("expected 1 text input after local copy (secure-exec hidden), got %d", len(m.textInputs))
	}
}

func TestStep14UseLocalCopyNoValidationRequired(t *testing.T) {
	// When UseLocalCopy is pre-configured as true, git remote URL is still required
	// (UseLocalCopy only affects the REPO env var in docker-compose)
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = true
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// textInputs should have 1 field (git remote URL still visible)
	if len(m.textInputs) != 1 {
		t.Fatalf("expected 1 text input when UseLocalCopy is true, got %d", len(m.textInputs))
	}

	// ValidateStepComplete should fail without git remote URL
	err := m.validateCurrentStep()
	if err == nil {
		t.Error("validateCurrentStep should fail when git remote URL is empty, even with UseLocalCopy=true")
	}
}


func TestStep14CheckboxRebuildsOnToggle(t *testing.T) {
	// Toggling checkboxes should rebuild text inputs correctly
	// When GithubCreateRepo=true && !UseLocalCopy: textInputs=2 (git remote URL + secure-exec driver)
	// When GithubCreateRepo=true && UseLocalCopy: textInputs=1 (git remote URL only, secure-exec hidden)
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// Initially: GithubCreateRepo=false, textInputs=0
	if len(m.textInputs) != 0 {
		t.Fatalf("default textInputs count = %d, want 0", len(m.textInputs))
	}

	// Toggle GithubCreateRepo on
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeySpace})
	// Now has: git remote URL + secure-exec driver = 2
	if len(m.textInputs) != 1 {
		t.Errorf("after GithubCreateRepo=true, textInputs count = %d, want 1", len(m.textInputs))
	}

	// Move to second checkbox and toggle UseLocalCopy on
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeySpace})
	// UseLocalCopy=true hides secure-exec driver, leaving only git remote URL = 1
	if len(m.textInputs) != 1 {
		t.Errorf("after UseLocalCopy=true, textInputs count = %d, want 1 (secure-exec hidden)", len(m.textInputs))
	}

	// Toggle UseLocalCopy off
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeySpace})
	// UseLocalCopy=false shows secure-exec driver again = 2
	if len(m.textInputs) != 1 {
		t.Errorf("after UseLocalCopy=false, textInputs count = %d, want 1", len(m.textInputs))
	}
}

func TestStep14LeftArrowTogglesCheckbox(t *testing.T) {
	// Left arrow should also toggle the current checkbox (like Space)
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	if m.config.GithubCreateRepo {
		t.Fatal("GithubCreateRepo should be false initially")
	}

	// Press left arrow to toggle
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyLeft})

	// Left arrow is not handled in checkbox mode - it falls through to text input
	// delegate, which does nothing when currentField==0. So left arrow should NOT toggle.
	// The test expects this to not toggle.
	if m.config.GithubCreateRepo {
		t.Error("GithubCreateRepo should still be false after left arrow (left arrow not supported in checkbox mode)")
	}
}

func TestStep14TabNavigatesBetweenCheckboxes(t *testing.T) {
	// Tab should move between checkboxes without toggling
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// Initially at checkboxIndex 0
	if m.checkboxIndex != 0 {
		t.Fatalf("checkboxIndex = %d, want 0", m.checkboxIndex)
	}

	// Tab should move to next checkbox (but there's only 1 visible initially)
	// So with only 1 checkbox visible, Tab on last checkbox with no text inputs
	// should advance to next step
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	// Since only 1 checkbox is visible and no text inputs, Tab should advance
	if m.step != 5 {
		t.Errorf("step = %d, want 5 (Tab on single checkbox with no text inputs should advance)", m.step)
	}
}

func TestStep14ViewContainsCheckboxLabel(t *testing.T) {
	// The view should always contain the checkbox label "Create GitHub repo:"
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	view := m.View()
	if !strings.Contains(view, "Create GitHub repo:") {
		t.Errorf("view should contain 'Create GitHub repo:', got:\n%s", view)
	}
	if !strings.Contains(view, "☐ Create GitHub repo") {
		t.Errorf("view should contain unchecked checkbox, got:\n%s", view)
	}
}

func TestStep6ConditionalFieldsHiddenWhenDev(t *testing.T) {
	// When "dev" is selected, Vault address and auth method should be hidden
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 7)

	// Default is "vault" (index 0), so conditional fields should be visible
	if len(m.textInputs) != 2 {
		t.Errorf("Step 7 default (vault) textInputs count = %d, want 2", len(m.textInputs))
	}

	// Move down to "dev" (index 1)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})

	// Now SecretsBackend should be "dev"
	if m.config.SecretsBackend != "dev" {
		t.Errorf("SecretsBackend = %q, want %q", m.config.SecretsBackend, "dev")
	}

	// No text inputs should be visible for "dev" mode
	if len(m.textInputs) != 0 {
		t.Errorf("Step 6 'dev' textInputs count = %d, want 0", len(m.textInputs))
	}

	// View should NOT contain Vault fields
	view := m.View()
	if strings.Contains(view, "Vault address") {
		t.Errorf("Step 6 view should NOT contain 'Vault address' when 'dev' is selected")
	}
	if strings.Contains(view, "Auth method") {
		t.Errorf("Step 6 view should NOT contain 'Auth method' when 'dev' is selected")
	}
}

func TestStep6ConditionalFieldsVisibleWhenVault(t *testing.T) {
	// When "vault" is selected, Vault address and auth method should be visible
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 7)

	// Default is "vault" (index 0)
	if m.config.SecretsBackend != "vault" {
		t.Fatalf("Default SecretsBackend = %q, want %q", m.config.SecretsBackend, "vault")
	}

	// Text inputs should be built for the 2 conditional fields
	if len(m.textInputs) != 2 {
		t.Errorf("Step 7 'vault' textInputs count = %d, want 2", len(m.textInputs))
	}

	// View should contain the conditional field labels
	view := m.View()
	if !strings.Contains(view, "Vault address") {
		t.Errorf("Step 6 view should contain 'Vault address' when 'vault' is selected, got: %s", view)
	}
	if !strings.Contains(view, "Auth method") {
		t.Errorf("Step 6 view should contain 'Auth method' when 'vault' is selected, got: %s", view)
	}
}

func TestStep6EnterOnVaultSwitchesToTextInput(t *testing.T) {
	// Pressing Enter on "vault" should switch to text input mode
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 7)

	// Default is "vault" (index 0), press Enter
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	// Should be in text input mode
	if m.mode != modeTextInput {
		t.Errorf("mode = %d, want modeTextInput(%d)", m.mode, modeTextInput)
	}
	if m.step != 7 {
		t.Errorf("step = %d, want 7 (should not advance)", m.step)
	}
	if !m.textInput.Focused() {
		t.Error("text input should be focused after switching to text input mode")
	}
}

func TestStep6EnterOnDevAdvances(t *testing.T) {
	// Pressing Enter on "dev" should advance to next step
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 7)

	// Move down to "dev" (index 1)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})

	// Press Enter
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	// Should have advanced to step 8
	if m.step != 8 {
		t.Errorf("step = %d, want 8", m.step)
	}
}

func TestStep6RadioRebuildsOnArrowKey(t *testing.T) {
	// Moving the radio selection should rebuild text inputs immediately
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 7)

	// Default "vault" (index 0) — 2 text inputs
	if len(m.textInputs) != 2 {
		t.Errorf("default textInputs count = %d, want 2", len(m.textInputs))
	}

	// Move down to "dev" (index 1)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})
	if len(m.textInputs) != 0 {
		t.Errorf("after 'dev' textInputs count = %d, want 0", len(m.textInputs))
	}

	// Move back up to "vault" (index 0)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyUp})
	if len(m.textInputs) != 2 {
		t.Errorf("after back to vault textInputs count = %d, want 2", len(m.textInputs))
	}
}

// ===== Step 7 conditional field tests (Redis) =====

func TestStep7ConditionalFieldHiddenWhenLocal(t *testing.T) {
	// When "local" is selected, Redis connection string should be hidden
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 8)

	// Default is "local" (index 0)
	if m.config.RedisMode != "local" {
		t.Fatalf("Default RedisMode = %q, want %q", m.config.RedisMode, "local")
	}

	// No text inputs should be visible for "local" mode
	if len(m.textInputs) != 0 {
		t.Errorf("Step 7 'local' textInputs count = %d, want 0", len(m.textInputs))
	}

	// View should NOT contain connection string field
	view := m.View()
	if strings.Contains(view, "Connection string") {
		t.Errorf("Step 7 view should NOT contain 'Connection string' when 'local' is selected")
	}
}

func TestStep7ConditionalFieldVisibleWhenExternal(t *testing.T) {
	// When "external" is selected, Redis connection string should be visible
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 8)

	// Move down to "external" (index 1)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})

	// Now RedisMode should be "external"
	if m.config.RedisMode != "external" {
		t.Errorf("RedisMode = %q, want %q", m.config.RedisMode, "external")
	}

	// Text input should be built for the connection string field
	if len(m.textInputs) != 1 {
		t.Errorf("Step 7 'external' textInputs count = %d, want 1", len(m.textInputs))
	}

	// View should contain the connection string field
	view := m.View()
	if !strings.Contains(view, "Connection string") {
		t.Errorf("Step 7 view should contain 'Connection string' when 'external' is selected, got: %s", view)
	}
}

func TestStep7EnterOnLocalAdvances(t *testing.T) {
	// Pressing Enter on "local" should advance to next step
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 8)

	// Default is "local" (index 0), press Enter
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	// Should have advanced to next required step (step 9 for default config)
	if m.step != 9 {
		t.Errorf("step = %d, want 9", m.step)
	}
}

func TestStep7EnterOnExternalSwitchesToTextInput(t *testing.T) {
	// Pressing Enter on "external" should switch to text input mode
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 8)

	// Move down to "external" (index 1)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})

	// Press Enter
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	// Should be in text input mode
	if m.mode != modeTextInput {
		t.Errorf("mode = %d, want modeTextInput(%d)", m.mode, modeTextInput)
	}
	if m.step != 8 {
		t.Errorf("step = %d, want 8 (should not advance)", m.step)
	}
}

// ===== Step 10 conditional field tests (AI Agent) =====

func TestStep10ConditionalFieldsHiddenWhenNo(t *testing.T) {
	// When "no" is selected, AI agent fields should be hidden
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 11)

	// Default is "no" (AIEnabled = false, radio index 1)
	if m.config.AIEnabled {
		t.Fatal("Default AIEnabled should be false")
	}

	// No text inputs should be visible for "no" selection
	if len(m.textInputs) != 0 {
		t.Errorf("Step 10 'no' textInputs count = %d, want 0", len(m.textInputs))
	}

	// View should NOT contain AI agent fields
	view := m.View()
	if strings.Contains(view, "API key secret path") {
		t.Errorf("Step 10 view should NOT contain 'API key secret path' when 'no' is selected")
	}
	if strings.Contains(view, "Base URL") {
		t.Errorf("Step 10 view should NOT contain 'Base URL' when 'no' is selected")
	}
	if strings.Contains(view, "Model") {
		t.Errorf("Step 10 view should NOT contain 'Model' when 'no' is selected")
	}
	if strings.Contains(view, "Engine name") {
		t.Errorf("Step 10 view should NOT contain 'Engine name' when 'no' is selected")
	}
}

func TestStep10ConditionalFieldsVisibleWhenYes(t *testing.T) {
	// When "yes" is selected, AI agent fields should be visible
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 11)

	// Move up to "yes" (index 0)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyUp})

	// Now AIEnabled should be true
	if !m.config.AIEnabled {
		t.Error("AIEnabled should be true after moving to 'yes'")
	}

	// Text inputs should be built for the 4 conditional fields
	if len(m.textInputs) != 4 {
		t.Errorf("Step 10 'yes' textInputs count = %d, want 4", len(m.textInputs))
	}

	// View should contain the conditional field labels
	view := m.View()
	if !strings.Contains(view, "API key secret path") {
		t.Errorf("Step 10 view should contain 'API key secret path' when 'yes' is selected, got: %s", view)
	}
	if !strings.Contains(view, "Base URL") {
		t.Errorf("Step 10 view should contain 'Base URL' when 'yes' is selected, got: %s", view)
	}
	if !strings.Contains(view, "Model") {
		t.Errorf("Step 10 view should contain 'Model' when 'yes' is selected, got: %s", view)
	}
	if !strings.Contains(view, "Engine name") {
		t.Errorf("Step 10 view should contain 'Engine name' when 'yes' is selected, got: %s", view)
	}
}

func TestStep10EnterOnYesSwitchesToTextInput(t *testing.T) {
	// Pressing Enter on "yes" should switch to text input mode
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 11)

	// Move up to "yes" (index 0)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyUp})

	// Press Enter
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	// Should be in text input mode
	if m.mode != modeTextInput {
		t.Errorf("mode = %d, want modeTextInput(%d)", m.mode, modeTextInput)
	}
	if m.step != 11 {
		t.Errorf("step = %d, want 11 (should not advance)", m.step)
	}
	if !m.textInput.Focused() {
		t.Error("text input should be focused after switching to text input mode")
	}
}

func TestStep10EnterOnNoAdvances(t *testing.T) {
	// Pressing Enter on "no" should advance to next step
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 11)

	// Default is "no" (index 1), press Enter
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	// Should have advanced to next required step (step 12 for docker mode)
	if m.step != 12 {
		t.Errorf("step = %d, want 12", m.step)
	}
}

func TestStep10RadioRebuildsOnArrowKey(t *testing.T) {
	// Moving the radio selection should rebuild text inputs immediately
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 11)

	// Default "no" (index 1) — no text inputs
	if len(m.textInputs) != 0 {
		t.Errorf("default textInputs count = %d, want 0", len(m.textInputs))
	}

	// Move up to "yes" (index 0)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyUp})
	if len(m.textInputs) != 4 {
		t.Errorf("after 'yes' textInputs count = %d, want 4", len(m.textInputs))
	}

	// Move back down to "no" (index 1)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})
	if len(m.textInputs) != 0 {
		t.Errorf("after back to 'no' textInputs count = %d, want 0", len(m.textInputs))
	}
}


func TestStep10ConditionalFieldTabNavigation(t *testing.T) {
	// After selecting "yes" and pressing Enter, tab should navigate through all 4 conditional fields
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 11)

	// Move to "yes" and press Enter to switch to text input mode
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyUp})
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.mode != modeTextInput {
		t.Fatalf("expected modeTextInput, got %d", m.mode)
	}

	// The AI fields have default values (AIModel: "gpt-4o", AIBaseURL: "https://api.openai.com/v1", AIEngineName: "default")
	// We need to clear them before typing new values. Use Ctrl+A then type, or just clear.
	// For the first field (API key secret path), no default value so just type
	for _, ch := range "secret/path" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}

	// Tab to second field (Base URL) — has default "https://api.openai.com/v1"
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})
	if m.currentField != 1 {
		t.Errorf("currentField = %d, want 1", m.currentField)
	}

	// Clear the default value and type new value
	m.textInput.SetValue("")
	for _, ch := range "https://api.example.com" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}

	// Tab to third field (Model) — has default "gpt-4o"
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})
	if m.currentField != 2 {
		t.Errorf("currentField = %d, want 2", m.currentField)
	}

	// Clear the default value and type new value
	m.textInput.SetValue("")
	for _, ch := range "gpt-4o" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}

	// Tab to fourth field (Engine name) — has default "default"
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})
	if m.currentField != 3 {
		t.Errorf("currentField = %d, want 3", m.currentField)
	}

	// Clear the default value and type new value
	m.textInput.SetValue("")
	for _, ch := range "ai-engine" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}

	// Press Enter on last field — should advance to next step
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != 12 {
		t.Errorf("step = %d, want 12", m.step)
	}

	// Verify all values were saved
	if m.config.AIAPIKeyPath != "secret/path" {
		t.Errorf("AIAPIKeyPath = %q, want %q", m.config.AIAPIKeyPath, "secret/path")
	}
	if m.config.AIBaseURL != "https://api.example.com" {
		t.Errorf("AIBaseURL = %q, want %q", m.config.AIBaseURL, "https://api.example.com")
	}
	if m.config.AIModel != "gpt-4o" {
		t.Errorf("AIModel = %q, want %q", m.config.AIModel, "gpt-4o")
	}
	if m.config.AIEngineName != "ai-engine" {
		t.Errorf("AIEngineName = %q, want %q", m.config.AIEngineName, "ai-engine")
	}
}


// ===== Clone auth tests =====

func TestStep15CloneAuthNone(t *testing.T) {
	// When clone auth is "none", no additional fields are needed
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.GitRemoteURL = "https://github.com/user/repo.git"
	cfg.ConfigRepoCloneAuth = "none"
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	err := m.validateCurrentStep()
	if err != nil {
		t.Errorf("validateCurrentStep should pass with none auth, got: %v", err)
	}
}

func TestStep15CloneAuthPAT(t *testing.T) {
	// When clone auth is "pat", PAT value is required
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.GitRemoteURL = "https://github.com/user/repo.git"
	cfg.ConfigRepoCloneAuth = "pat"
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	err := m.validateCurrentStep()
	if err == nil {
		t.Error("validateCurrentStep should fail without PAT value")
	}
	if err != nil && !strings.Contains(err.Error(), "PAT") {
		t.Errorf("unexpected error: %v", err)
	}

	// Now set the PAT value (env var reference)
	cfg.ConfigRepoPATValue = "$MY_PAT"
	m2 := NewWizard(cfg)
	m2 = runInitStep(m2, 6)

	err = m2.validateCurrentStep()
	if err != nil {
		t.Errorf("validateCurrentStep should pass with PAT value set, got: %v", err)
	}

	// Also test with raw PAT value
	cfg.ConfigRepoPATValue = "ghp_xxxxx"
	m3 := NewWizard(cfg)
	m3 = runInitStep(m3, 6)

	err = m3.validateCurrentStep()
	if err != nil {
		t.Errorf("validateCurrentStep should pass with raw PAT value, got: %v", err)
	}
}

func TestStep15CloneAuthGitHubApp(t *testing.T) {
	// When clone auth is "github_app", all 4 fields are required
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.GitRemoteURL = "https://github.com/user/repo.git"
	cfg.ConfigRepoCloneAuth = "github_app"
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	err := m.validateCurrentStep()
	if err == nil {
		t.Error("validateCurrentStep should fail without GitHub App ID")
	}

	// Set ID but missing others
	cfg.ConfigRepoGHAppID = "12345"
	m2 := NewWizard(cfg)
	m2 = runInitStep(m2, 6)
	err = m2.validateCurrentStep()
	if err == nil {
		t.Error("validateCurrentStep should fail without Installation ID")
	}

	// Set ID and Installation ID but missing key
	cfg.ConfigRepoGHInstallID = "67890"
	m3 := NewWizard(cfg)
	m3 = runInitStep(m3, 6)
	err = m3.validateCurrentStep()
	if err == nil {
		t.Error("validateCurrentStep should fail without Private Key")
	}

	// Set all fields
	cfg.ConfigRepoGHAppKey = "fake-key-content"
	m4 := NewWizard(cfg)
	m4 = runInitStep(m4, 6)
	err = m4.validateCurrentStep()
	if err != nil {
		t.Errorf("validateCurrentStep should pass with all github_app fields, got: %v", err)
	}
}

func TestStep15CloneAuthSSH(t *testing.T) {
	// When clone auth is "ssh", URL must start with git@ and key or file is required
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.GitRemoteURL = "https://github.com/user/repo.git"
	cfg.ConfigRepoCloneAuth = "ssh"
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	// URL doesn't start with git@
	err := m.validateCurrentStep()
	if err == nil {
		t.Error("validateCurrentStep should fail with https URL for ssh auth")
	}
	if err != nil && !strings.Contains(err.Error(), "git@") {
		t.Errorf("unexpected error: %v", err)
	}

	// Fix URL but no key or file
	cfg.GitRemoteURL = "git@github.com:user/repo.git"
	m2 := NewWizard(cfg)
	m2 = runInitStep(m2, 6)
	err = m2.validateCurrentStep()
	if err == nil {
		t.Error("validateCurrentStep should fail without SSH key or file")
	}

	// Set SSH key content
	cfg.ConfigRepoSSHKey = "fake-ssh-key"
	m3 := NewWizard(cfg)
	m3 = runInitStep(m3, 6)
	err = m3.validateCurrentStep()
	if err != nil {
		t.Errorf("validateCurrentStep should pass with SSH key, got: %v", err)
	}

	// Set SSH key file instead
	cfg.ConfigRepoSSHKey = ""
	cfg.ConfigRepoSSHFile = "/home/user/.ssh/id_rsa"
	m4 := NewWizard(cfg)
	m4 = runInitStep(m4, 6)
	err = m4.validateCurrentStep()
	if err != nil {
		t.Errorf("validateCurrentStep should pass with SSH key file, got: %v", err)
	}
}

func TestStep15CloneAuthInvalid(t *testing.T) {
	// Invalid auth method should fail
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.GitRemoteURL = "https://github.com/user/repo.git"
	cfg.GithubCreateRepo = true
	cfg.ConfigRepoCloneAuth = "invalid_auth"
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	err := m.validateCurrentStep()
	if err == nil {
		t.Error("validateCurrentStep should fail with invalid auth method")
	}
	if err != nil && !strings.Contains(err.Error(), "invalid clone auth method") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStep15CloneAuthSkippedForLocalCopy(t *testing.T) {
	// When UseLocalCopy is true, clone auth is not validated
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = true
	cfg.GitRemoteURL = "git@github.com:user/repo.git"
	cfg.ConfigRepoCloneAuth = "ssh" // would normally require key
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	err := m.validateCurrentStep()
	if err != nil {
		t.Errorf("validateCurrentStep should pass with UseLocalCopy=true (no clone auth validation), got: %v", err)
	}
}

func TestStep15CloneAuthNoneNoAdditionalFields(t *testing.T) {
	// With none auth and https URL, no additional fields needed
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.GitRemoteURL = "https://github.com/user/repo.git"
	cfg.ConfigRepoCloneAuth = "none"
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	err := m.validateCurrentStep()
	if err != nil {
		t.Errorf("validateCurrentStep should pass with none auth and https URL, got: %v", err)
	}
}

func TestStep15CloneAuthConditionalFieldsVisible(t *testing.T) {
	// When clone auth is "pat", the PAT field should be visible
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.GitRemoteURL = "https://github.com/user/repo.git"
	cfg.ConfigRepoCloneAuth = "pat"
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	// The clone auth field and PAT field should be visible
	stepInfo := getStepInfo(6)
	visibleFields := 0
	for _, field := range stepInfo.fields {
		if field.condition == nil || field.condition(m.config) {
			visibleFields++
		}
	}

	// Should have: Clone auth method, PAT = 2
	if visibleFields != 2 {
		t.Errorf("expected 2 visible fields for pat auth, got %d", visibleFields)
	}
}

func TestStep15CloneAuthSSHFieldsVisible(t *testing.T) {
	// When clone auth is "ssh", SSH key fields should be visible
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.GitRemoteURL = "git@github.com:user/repo.git"
	cfg.ConfigRepoCloneAuth = "ssh"
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	stepInfo := getStepInfo(6)
	visibleFields := 0
	for _, field := range stepInfo.fields {
		if field.condition == nil || field.condition(m.config) {
			visibleFields++
		}
	}

	// Should have: Clone auth method, SSH key content, SSH key file, SSH key passphrase = 4
	if visibleFields != 4 {
		t.Errorf("expected 4 visible fields for ssh auth, got %d", visibleFields)
	}
}

func TestStep15CloneAuthHiddenForLocalCopy(t *testing.T) {
	// When UseLocalCopy is true, clone auth fields should be hidden
	// (they are on step 11, which is conditionally shown based on !UseLocalCopy)
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = true
	cfg.ConfigRepoCloneAuth = "pat"
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	stepInfo := getStepInfo(4)
	visibleFields := 0
	for _, field := range stepInfo.fields {
		if field.condition == nil || field.condition(m.config) {
			visibleFields++
		}
	}

	// Should only have: Git remote URL = 1 (secure-exec driver hidden when UseLocalCopy=true)
	if visibleFields != 1 {
		t.Errorf("expected 1 visible field when UseLocalCopy=true, got %d", visibleFields)
	}
}

func TestStep15SummaryShowsCloneAuth(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.ConfigRepoCloneAuth = "github_app"
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	view := m.View()
	if !strings.Contains(view, "Clone auth method") {
		t.Errorf("summary should show clone auth method, got view:\n%s", view)
	}
}


func TestStep15PlaceholderDefaultNoSecureExec(t *testing.T) {
	// When no secure-exec driver is selected, default placeholders should be shown
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.SecureExecDriver = "none"
	cfg.ConfigRepoCloneAuth = "pat"

	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	// The first visible field is "Clone auth method", second is "PAT"
	// With buildMultiFieldInputs, the textInputs should have the default placeholder
	if len(m.textInputs) < 2 {
		t.Fatal("expected at least 2 text inputs for step 11 with pat auth")
	}
	// PAT field is the 2nd visible field (clone auth method, PAT)
	patInput := m.textInputs[1]
	if patInput.Placeholder != "ghp_xxxxx or $MY_PAT" {
		t.Errorf("PAT placeholder = %q, want %q", patInput.Placeholder, "ghp_xxxxx or $MY_PAT")
	}
}

func TestStep6PlaceholderWithSecureExecVault(t *testing.T) {
	// When hd-driver-vault is selected, placeholders should show hd-lookup paths
	// Private key value field is hidden, but GH App key secret path is shown
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.SecureExecDriver = "hd-driver-vault"
	cfg.ConfigRepoCloneAuth = "github_app"

	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	if len(m.textInputs) != 4 {
		t.Fatalf("expected 4 text inputs for step 11 with github_app auth + secure-exec, got %d", len(m.textInputs))
	}
	// Fields: Clone auth method, GitHub App ID, Installation ID, GH App key secret path
	// Index 1 = GitHub App ID
	if m.textInputs[1].Placeholder != "hd-lookup:/secrets/data/project/gh_app_id" {
		t.Errorf("GitHub App ID placeholder = %q, want %q", m.textInputs[1].Placeholder, "hd-lookup:/secrets/data/project/gh_app_id")
	}
	// Index 2 = Installation ID
	if m.textInputs[2].Placeholder != "hd-lookup:/secrets/data/project/gh_install_id" {
		t.Errorf("Installation ID placeholder = %q, want %q", m.textInputs[2].Placeholder, "hd-lookup:/secrets/data/project/gh_install_id")
	}
	// Index 3 = GH App key secret path
	if m.textInputs[3].Placeholder != "/secrets/data/project/gh_app_key" {
		t.Errorf("GH App key path placeholder = %q, want %q", m.textInputs[3].Placeholder, "hd-lookup:/secrets/data/project/gh_app_key")
	}
}

func TestStep6PlaceholderWithSecureExecGcloud(t *testing.T) {
	// When gcloud-secret is selected, placeholders should show hd-lookup paths
	// SSH key content is hidden, but SSH key secret path is shown
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.SecureExecDriver = "gcloud-secret"
	cfg.ConfigRepoCloneAuth = "ssh"

	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	if len(m.textInputs) != 4 {
		t.Fatalf("expected 4 text inputs for step 11 with ssh auth + secure-exec, got %d", len(m.textInputs))
	}
	// Fields: Clone auth method, SSH key file, SSH key passphrase, SSH key secret path
	// Index 1 = SSH key file path
	if m.textInputs[1].Placeholder != "/home/user/.ssh/id_rsa" {
		t.Errorf("SSH key file placeholder = %q, want %q", m.textInputs[1].Placeholder, "/home/user/.ssh/id_rsa")
	}
	// Index 3 = SSH key secret path
	if m.textInputs[3].Placeholder != "/secrets/data/project/ssh_key" {
		t.Errorf("SSH key path placeholder = %q, want %q", m.textInputs[3].Placeholder, "hd-lookup:/secrets/data/project/ssh_key")
	}
}

func TestStep15PlaceholderSSHKeyContentDefault(t *testing.T) {
	// Without secure-exec, SSH key content should show BEGIN OPENSSH placeholder
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.SecureExecDriver = ""
	cfg.ConfigRepoCloneAuth = "ssh"

	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	if len(m.textInputs) < 2 {
		t.Fatal("expected at least 2 text inputs")
	}
	// Index 1 = SSH key content
	if m.textInputs[1].Placeholder != "-----BEGIN OPENSSH PRIVATE KEY-----\n..." {
		t.Errorf("SSH key content placeholder = %q, want default", m.textInputs[1].Placeholder)
	}
}


func TestStep6SecureExecSecretPathFieldsVisible(t *testing.T) {
	// When secure-exec is enabled with PAT auth, the PAT secret path field should be visible
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.SecureExecDriver = "hd-driver-vault"
	cfg.ConfigRepoCloneAuth = "pat"

	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	// Fields: Clone auth method, PAT value, PAT secret path = 3
	if len(m.textInputs) != 3 {
		t.Fatalf("expected 3 text inputs for step 11 with pat + secure-exec, got %d", len(m.textInputs))
	}
	// Index 2 = PAT secret path (PAT value field is still shown at index 1)
	if m.textInputs[2].Placeholder != "/secrets/data/project/pat" {
		t.Errorf("PAT secret path placeholder = %q, want %q", m.textInputs[2].Placeholder, "hd-lookup:/secrets/data/project/pat")
	}
}

func TestStep5SecureExecPATValueAccepted(t *testing.T) {
	// When PAT secret path is set, validation should pass without PAT value
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.GitRemoteURL = "https://github.com/user/repo.git"
	cfg.ConfigRepoCloneAuth = "pat"
	cfg.SecureExecDriver = "hd-driver-vault"
	cfg.ConfigRepoPATPath = "hd-lookup:/secrets/data/project/pat"
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	err := m.validateCurrentStep()
	if err != nil {
		t.Errorf("validateCurrentStep should pass with PAT secret path set, got: %v", err)
	}
}

func TestStep6SecureExecSSHKeyPathAccepted(t *testing.T) {
	// When SSH key secret path is set, validation should pass without SSH key content
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.GitRemoteURL = "git@github.com:user/repo.git"
	cfg.ConfigRepoCloneAuth = "ssh"
	cfg.SecureExecDriver = "gcloud-secret"
	cfg.ConfigRepoSSHKeyPath = "hd-lookup:/secrets/data/project/ssh_key"
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	err := m.validateCurrentStep()
	if err != nil {
		t.Errorf("validateCurrentStep should pass with SSH key secret path set, got: %v", err)
	}
}

func TestStep6SecureExecGHAppKeyPathAccepted(t *testing.T) {
	// When GH App key secret path is set, validation should pass without private key value
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.GitRemoteURL = "https://github.com/user/repo.git"
	cfg.ConfigRepoCloneAuth = "github_app"
	cfg.SecureExecDriver = "hd-driver-vault"
	cfg.ConfigRepoGHAppID = "12345"
	cfg.ConfigRepoGHInstallID = "67890"
	cfg.ConfigRepoGHAppKeyPath = "hd-lookup:/secrets/data/project/gh_app_key"
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	err := m.validateCurrentStep()
	if err != nil {
		t.Errorf("validateCurrentStep should pass with GH App key secret path set, got: %v", err)
	}
}

func TestStep15SecureExecFieldsHiddenWithoutDriver(t *testing.T) {
	// When no secure-exec driver is set, secret path fields should be hidden
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.SecureExecDriver = "none"
	cfg.ConfigRepoCloneAuth = "pat"

	stepInfo := getStepInfo(6)
	visibleFields := 0
	for _, field := range stepInfo.fields {
		if field.condition == nil || field.condition(cfg) {
			visibleFields++
		}
	}
	// Should have: Clone auth method, PAT value = 2
	// (no PAT secret path since driver is "none")
	if visibleFields != 2 {
		t.Errorf("expected 2 visible fields with pat + no secure-exec, got %d", visibleFields)
	}
}

// ===== Space key in checkbox text fields =====

func TestSpaceKeyInCheckboxTextField(t *testing.T) {
	// When in text field mode within a multi-field step, Space should insert
	// a space character.
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	// Step 15 is a multi-field step (no checkboxes), currentField starts at 0
	if m.currentField != 0 {
		t.Fatalf("currentField = %d, want 0", m.currentField)
	}

	// Type "hello" then Space then "world"
	for _, ch := range "hello" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeySpace})
	for _, ch := range "world" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}

	// The text input should contain "hello world" with a space
	if m.textInput.Value() != "hello world" {
		t.Errorf("textInput.Value() = %q, want %q", m.textInput.Value(), "hello world")
	}
}

func TestSpaceKeyStillTogglesCheckbox(t *testing.T) {
	// When on checkbox row (currentField == 0), Space should still toggle checkboxes.
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	if m.currentField != 0 {
		t.Fatalf("currentField = %d, want 0", m.currentField)
	}
	if m.config.GithubCreateRepo {
		t.Fatal("GithubCreateRepo should be false initially")
	}

	// Press Space to toggle
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeySpace})

	if !m.config.GithubCreateRepo {
		t.Error("GithubCreateRepo should be true after pressing space on checkbox")
	}
}

// ===== Paste mode tests =====

func TestPasteModeToggleWithCtrlE(t *testing.T) {
	// Ctrl+E in text input mode should enter paste mode.
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 2) // Project Name step (text input)

	// Type something first
	for _, ch := range "test" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}

	// Press Ctrl+E to enter paste mode
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlE})

	if !m.pasteModeActive {
		t.Error("pasteModeActive should be true after Ctrl+E")
	}
}

func TestPasteModeShowsInView(t *testing.T) {
	// When paste mode is active, the view should show the paste overlay.
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 2)

	// Enter paste mode
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlE})

	view := m.View()
	if !strings.Contains(view, "Paste Mode") {
		t.Errorf("View() should contain 'Paste Mode' header, got: %s", view)
	}
	if !strings.Contains(view, "Ctrl+D to save") {
		t.Errorf("View() should contain 'Ctrl+D to save' hint, got: %s", view)
	}
}

func TestPasteModeSavesOnCtrlD(t *testing.T) {
	// Ctrl+D in paste mode should save the content and exit paste mode.
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 2)

	// Enter paste mode
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlE})

	// Type multi-line content into the textarea
	m.pasteModeTextArea.SetValue("line1\nline2\nline3")

	// Press Ctrl+D to save
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlD})

	if m.pasteModeActive {
		t.Error("pasteModeActive should be false after Ctrl+D")
	}
	// The config should have the multi-line value preserved
	if m.config.ProjectName != "line1\nline2\nline3" {
		t.Errorf("ProjectName = %q, want multi-line content", m.config.ProjectName)
	}
}

func TestPasteModeCancelsOnEsc(t *testing.T) {
	// Escape in paste mode should cancel and restore the original value.
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 2)

	// Type something first
	for _, ch := range "original" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		m, _ = updateWizard(m, msg)
	}

	// Enter paste mode
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlE})

	// Change the content in the textarea
	m.pasteModeTextArea.SetValue("changed content")

	// Press Esc to cancel
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})

	if m.pasteModeActive {
		t.Error("pasteModeActive should be false after Esc")
	}
	// The original value should be preserved
	if m.textInput.Value() != "original" {
		t.Errorf("textInput.Value() = %q, want %q (original value should be restored)", m.textInput.Value(), "original")
	}
}

func TestPasteModeInCheckboxTextField(t *testing.T) {
	// Ctrl+E should work in multi-field step text fields too.
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	// Step 15 is a multi-field step (no checkboxes), currentField starts at 0
	if m.currentField != 0 {
		t.Fatalf("currentField = %d, want 0", m.currentField)
	}

	// Press Ctrl+E to enter paste mode
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlE})

	if !m.pasteModeActive {
		t.Error("pasteModeActive should be true after Ctrl+E in checkbox text field")
	}

	// Type multi-line content
	m.pasteModeTextArea.SetValue("ssh-rsa AAAA...")

	// Save with Ctrl+D
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlD})

	if m.pasteModeActive {
		t.Error("pasteModeActive should be false after Ctrl+D")
	}
	if m.textInput.Value() != "ssh-rsa AAAA..." {
		t.Errorf("textInput.Value() = %q, want %q", m.textInput.Value(), "ssh-rsa AAAA...")
	}
}

func TestPasteModeMultiLineSSHKey(t *testing.T) {
	// Simulate pasting a multi-line SSH private key.
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.GitRemoteURL = "git@github.com:user/repo.git"
	cfg.ConfigRepoCloneAuth = "ssh"
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	// Navigate to the SSH key content field in step 11 (multi-field, no checkboxes)
	// Visible fields with ssh auth: Clone auth (0), SSH key content (1), SSH key file (2), SSH key passphrase (3), Enable secure exec (4)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab}) // to SSH key content

	if m.currentField != 1 {
		t.Fatalf("currentField = %d, want 1 (SSH key content)", m.currentField)
	}

	// Enter paste mode
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlE})

	// Paste multi-line SSH key
	sshKey := "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAA\n-----END OPENSSH PRIVATE KEY-----\n"
	m.pasteModeTextArea.SetValue(sshKey)

	// Save with Ctrl+D
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlD})

	if m.pasteModeActive {
		t.Error("pasteModeActive should be false after Ctrl+D")
	}

	// The config should have the multi-line SSH key preserved
	if m.config.ConfigRepoSSHKey != sshKey {
		t.Errorf("ConfigRepoSSHKey not preserved correctly, got %q", m.config.ConfigRepoSSHKey)
	}
}

func TestPasteModeMultiLineSSHKeyPreservedOnTab(t *testing.T) {
	// After pasting a multi-line SSH key and pressing Tab to leave the field,
	// the config should still have the raw multi-line value (not the collapsed display).
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.GitRemoteURL = "git@github.com:user/repo.git"
	cfg.ConfigRepoCloneAuth = "ssh"
	m := NewWizard(cfg)
	m = runInitStep(m, 6)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})   // to SSH key content

	if m.currentField != 1 {
		t.Fatalf("currentField = %d, want 1 (SSH key content)", m.currentField)
	}

	// Enter paste mode and paste multi-line SSH key
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlE})
	sshKey := "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAA\n-----END OPENSSH PRIVATE KEY-----\n"
	m.pasteModeTextArea.SetValue(sshKey)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlD})

	// Press Tab to move to next field
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	// The config should still have the multi-line SSH key
	if m.config.ConfigRepoSSHKey != sshKey {
		t.Errorf("ConfigRepoSSHKey was corrupted after Tab!\nGot:  %q\nWant: %q", m.config.ConfigRepoSSHKey, sshKey)
	}
	if !strings.Contains(m.config.ConfigRepoSSHKey, "\n") {
		t.Errorf("ConfigRepoSSHKey lost newlines after Tab: %q", m.config.ConfigRepoSSHKey)
	}
}

func TestPasteModeMultiLineSSHKeyPreservedOnEnter(t *testing.T) {
	// After pasting a multi-line SSH key and pressing Enter to leave the field,
	// the config should still have the raw multi-line value.
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.GitRemoteURL = "git@github.com:user/repo.git"
	cfg.ConfigRepoCloneAuth = "ssh"
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	// Navigate to the SSH key content field in step 11 (multi-field, no checkboxes)
	// Visible fields with ssh auth: Clone auth (0), SSH key content (1)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	if m.currentField != 1 {
		t.Fatalf("currentField = %d, want 1 (SSH key content)", m.currentField)
	}

	// Enter paste mode and paste multi-line SSH key
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlE})
	sshKey := "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAA\n-----END OPENSSH PRIVATE KEY-----\n"
	m.pasteModeTextArea.SetValue(sshKey)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlD})

	// Press Enter to move to next field
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	// The config should still have the multi-line SSH key
	if m.config.ConfigRepoSSHKey != sshKey {
		t.Errorf("ConfigRepoSSHKey was corrupted after Enter!\nGot:  %q\nWant: %q", m.config.ConfigRepoSSHKey, sshKey)
	}
	if !strings.Contains(m.config.ConfigRepoSSHKey, "\n") {
		t.Errorf("ConfigRepoSSHKey lost newlines after Enter: %q", m.config.ConfigRepoSSHKey)
	}
}

func TestPasteModeHintShownInNavigation(t *testing.T) {
	// Navigation hints should include ctrl+e=paste mode in text input mode.
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 2) // Project Name step (text input)

	view := m.View()
	if !strings.Contains(view, "ctrl+e=paste mode") {
		t.Errorf("View() should contain 'ctrl+e=paste mode' hint in text input mode, got: %s", view)
	}
}

func TestPasteModeHintShownInCheckboxTextField(t *testing.T) {
	// Navigation hints should include ctrl+e=paste mode in checkbox text fields.
	cfg := config.NewDefaultWizardConfig()
	cfg.GithubCreateRepo = true
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	// Move to text input
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown}) // to checkbox 1
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown}) // to text input

	view := m.View()
	if !strings.Contains(view, "ctrl+e=paste mode") {
		t.Errorf("View() should contain 'ctrl+e=paste mode' hint in checkbox text field mode, got: %s", view)
	}
}

func TestPasteModeCtrlCQuits(t *testing.T) {
	// Ctrl+C should quit even in paste mode.
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 2)

	// Enter paste mode
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlE})

	if !m.pasteModeActive {
		t.Fatal("should be in paste mode")
	}

	// Press Ctrl+C - should quit
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if cmd == nil {
		t.Error("ctrl+c in paste mode should produce a quit command")
	}
	if !m.quit {
		t.Error("quit flag should be true after ctrl+c in paste mode")
	}
}

func TestPasteModeCtrlSSaves(t *testing.T) {
	// Ctrl+S should save answers even in paste mode.
	cfg := config.NewDefaultWizardConfig()
	cfg.ProjectName = "paste-save-test"
	m := NewWizard(cfg)
	m = runInitStep(m, 2)

	// Enter paste mode
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlE})

	if !m.pasteModeActive {
		t.Fatal("should be in paste mode")
	}

	// Press Ctrl+S - should save (and commit current text input value)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlS})

	// Should show save confirmation
	view := m.View()
	if !strings.Contains(view, "Answers saved to") {
		t.Errorf("Ctrl+S in paste mode should trigger save, got: %s", view)
	}
}

// ===== Multi-line paste mode: raw value preservation tests =====

func TestPasteModeRawValuePreservedInConfig(t *testing.T) {
	// After paste mode saves multi-line content, the config should have
	// the raw multi-line value even though the textinput displays a
	// collapsed single-line version.
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 2) // Project Name step (text input)

	// Enter paste mode
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlE})

	// Paste multi-line content
	sshKey := "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAA\n-----END OPENSSH PRIVATE KEY-----\n"
	m.pasteModeTextArea.SetValue(sshKey)

	// Save with Ctrl+D
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlD})

	if m.pasteModeActive {
		t.Fatal("pasteModeActive should be false after Ctrl+D")
	}

	// The config should have the raw multi-line value
	if m.config.ProjectName != sshKey {
		t.Errorf("config.ProjectName should have raw multi-line value, got %q", m.config.ProjectName)
	}

	// The textinput should show a collapsed display value
	if !strings.Contains(m.textInput.Value(), "...") {
		t.Errorf("textinput should show collapsed value with '...', got %q", m.textInput.Value())
	}
	if strings.Contains(m.textInput.Value(), "\n") {
		t.Errorf("textinput should NOT contain newlines, got %q", m.textInput.Value())
	}
}

func TestPasteModeRawValueReopened(t *testing.T) {
	// When re-entering paste mode on a field with a raw multi-line value,
	// the textarea should show the raw value, not the collapsed display.
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 2)

	// Enter paste mode and save multi-line content
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlE})
	m.pasteModeTextArea.SetValue("line1\nline2\nline3")
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlD})

	// Re-enter paste mode
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlE})

	if !m.pasteModeActive {
		t.Fatal("should be in paste mode")
	}

	// The textarea should have the raw multi-line value
	if m.pasteModeTextArea.Value() != "line1\nline2\nline3" {
		t.Errorf("textarea should have raw multi-line value, got %q", m.pasteModeTextArea.Value())
	}
}

func TestPasteModeSingleLineNoRawValue(t *testing.T) {
	// When pasting single-line content, rawFieldValues should not be set.
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 2)

	// Enter paste mode and save single-line content
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlE})
	m.pasteModeTextArea.SetValue("single-line-value")
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlD})

	// rawFieldValues should not contain this field
	if m.rawFieldValues != nil {
		key := m.pasteModeFieldKey()
		if _, ok := m.rawFieldValues[key]; ok {
			t.Error("rawFieldValues should not contain single-line value")
		}
	}

	// The textinput should show the exact value
	if m.textInput.Value() != "single-line-value" {
		t.Errorf("textInput.Value() = %q, want %q", m.textInput.Value(), "single-line-value")
	}
}

func TestCollapseForDisplay(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"single line", "hello world", "hello world"},
		{"multi-line", "line1\nline2\nline3", "line1..."},
		{"empty", "", ""},
		{"single newline", "hello\n", "hello..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collapseForDisplay(tt.input)
			if got != tt.expected {
				t.Errorf("collapseForDisplay(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// ===== Multi-line paste: Tab navigation in checkbox step =====

func TestPasteModeCheckboxTabPreservesRawValueInConfig(t *testing.T) {
	// Reproduces the user-reported bug: after pasting a multi-line SSH key in a
	// checkbox-step text field and pressing Tab to leave the field, the config
	// must still contain the raw multi-line value (not collapsed to single line).
	cfg := config.NewDefaultWizardConfig()
	cfg.ProjectName = "test-project"
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.ConfigRepoCloneAuth = "ssh"
	cfg.GitRemoteURL = "git@github.com:test/repo.git"
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})   // to SSH key content

	if m.currentField != 1 {
		t.Fatalf("currentField = %d, want 1 (SSH key content)", m.currentField)
	}

	// Enter paste mode and paste multi-line SSH key
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlE})
	sshKey := "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAA\n-----END OPENSSH PRIVATE KEY-----\n"
	m.pasteModeTextArea.SetValue(sshKey)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlD})

	// Press Tab to move to next field — this is the exact user action that triggers the bug
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	// Config must still have the raw multi-line value
	if m.config.ConfigRepoSSHKey != sshKey {
		t.Errorf("config.ConfigRepoSSHKey corrupted after Tab!\nGot:  %q\nWant: %q", m.config.ConfigRepoSSHKey, sshKey)
	}
	if !strings.Contains(m.config.ConfigRepoSSHKey, "\n") {
		t.Errorf("config.ConfigRepoSSHKey lost newlines after Tab: %q", m.config.ConfigRepoSSHKey)
	}
}

func TestPasteModeCheckboxEnterPreservesRawValueInConfig(t *testing.T) {
	// Same scenario but using Enter instead of Tab to leave the field
	cfg := config.NewDefaultWizardConfig()
	cfg.ProjectName = "test-project"
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.ConfigRepoCloneAuth = "ssh"
	cfg.GitRemoteURL = "git@github.com:test/repo.git"
	m := NewWizard(cfg)
	m = runInitStep(m, 6)
	// Navigate to SSH key content field
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})   // to SSH key content

	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlE})
	sshKey := "line1\nline2\nline3"
	m.pasteModeTextArea.SetValue(sshKey)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlD})

	// Press Enter to move to next field
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.config.ConfigRepoSSHKey != sshKey {
		t.Errorf("config.ConfigRepoSSHKey corrupted after Enter!\nGot:  %q\nWant: %q", m.config.ConfigRepoSSHKey, sshKey)
	}
	if !strings.Contains(m.config.ConfigRepoSSHKey, "\n") {
		t.Errorf("config.ConfigRepoSSHKey lost newlines after Enter: %q", m.config.ConfigRepoSSHKey)
	}
}

func TestPasteModeCheckboxStepNavigationPreservesRawValue(t *testing.T) {
	// Full round-trip: paste → Tab → advance to next step → Esc back → verify config
	cfg := config.NewDefaultWizardConfig()
	cfg.ProjectName = "test-project"
	cfg.GithubCreateRepo = true
	cfg.UseLocalCopy = false
	cfg.ConfigRepoCloneAuth = "ssh"
	cfg.GitRemoteURL = "git@github.com:test/repo.git"
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	// Navigate to SSH key content field
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})   // to SSH key content
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlE})
	sshKey := "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAA\n-----END OPENSSH PRIVATE KEY-----\n"
	m.pasteModeTextArea.SetValue(sshKey)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyCtrlD})

	// Tab to next field
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	// Navigate to last text field and advance to next step
	for m.currentField < len(m.textInputs) {
		m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})
	}
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	// Config must still have raw value after advancing
	if m.config.ConfigRepoSSHKey != sshKey {
		t.Errorf("config corrupted after advancing to next step: got %q", m.config.ConfigRepoSSHKey)
	}

	// Go back to step 4
	if m.step == 15 {
		m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})
		if m.step == 14 {
			// Navigate to SSH key field and verify
			m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})
			m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})
			m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})
			m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

			if m.config.ConfigRepoSSHKey != sshKey {
				t.Errorf("config corrupted after round-trip: got %q", m.config.ConfigRepoSSHKey)
			}
			if !strings.Contains(m.config.ConfigRepoSSHKey, "\n") {
				t.Errorf("config lost newlines after round-trip: got %q", m.config.ConfigRepoSSHKey)
			}
		}
	}
}
