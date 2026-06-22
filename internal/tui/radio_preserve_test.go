// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If a separate written commercial agreement exists, you may use this file under those terms instead.

package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Charles546/hd-cli/internal/config"
)

// TestRadioSelectionPreservedOnReenter verifies that pressing Enter to advance
// to the next step and then Esc to return preserves the radio selection.
func TestRadioSelectionPreservedOnReenter(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	if m.radioIndex != 0 {
		t.Fatalf("initial radioIndex = %d, want 0 (docker)", m.radioIndex)
	}

	// Advance to step 5
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != 5 {
		t.Fatalf("after Enter: step = %d, want 5", m.step)
	}

	// Go back to step 4
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.step != 4 {
		t.Fatalf("after Esc: step = %d, want 4", m.step)
	}

	if m.radioIndex != 0 {
		t.Errorf("radioIndex = %d, want 0 (docker)", m.radioIndex)
	}
	if m.config.DeploymentMode != "docker" {
		t.Errorf("DeploymentMode = %q, want %q", m.config.DeploymentMode, "docker")
	}
}

// TestRadioSelectionChangePreservedOnReenter verifies that changing the radio
// selection and then navigating away and back preserves the changed selection.
func TestRadioSelectionChangePreservedOnReenter(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// Change to "source" (index 1)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.radioIndex != 1 {
		t.Fatalf("after Down: radioIndex = %d, want 1 (source)", m.radioIndex)
	}

	// Advance to step 5
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != 5 {
		t.Fatalf("after Enter: step = %d, want 5", m.step)
	}

	// Go back to step 4
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.step != 4 {
		t.Fatalf("after Esc: step = %d, want 4", m.step)
	}

	if m.radioIndex != 1 {
		t.Errorf("radioIndex = %d, want 1 (source)", m.radioIndex)
	}
	if m.config.DeploymentMode != "source" {
		t.Errorf("DeploymentMode = %q, want %q", m.config.DeploymentMode, "source")
	}
}

// TestRadioSelectionRevertedOnEsc verifies that pressing Esc without committing
// (Enter) reverts the radio selection to the original value.
func TestRadioSelectionRevertedOnEsc(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	// Default is "vault" (index 0)
	if m.radioIndex != 0 {
		t.Fatalf("initial radioIndex = %d, want 0 (vault)", m.radioIndex)
	}

	// Move to "dev" (index 1) — this tentatively applies to config
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.radioIndex != 1 {
		t.Fatalf("after Down: radioIndex = %d, want 1 (dev)", m.radioIndex)
	}
	if m.config.SecretsBackend != "dev" {
		t.Fatalf("after Down: SecretsBackend = %q, want %q", m.config.SecretsBackend, "dev")
	}

	// Press Esc without committing — should revert to "vault"
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.step != 5 {
		t.Fatalf("after Esc: step = %d, want 5", m.step)
	}

	// Config should be reverted
	if m.config.SecretsBackend != "vault" {
		t.Errorf("after Esc: SecretsBackend = %q, want %q (reverted)", m.config.SecretsBackend, "vault")
	}

	// Navigate forward through step 5 (multi-field: 3 visible fields) to reach step 6
	// Step 5 has fields: EssentialsRepoURL, Branch, EssentialsCloneType (3 visible with default "none")
	for m.step == 5 {
		m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})
	}
	if m.step != 6 {
		t.Fatalf("after navigating through step 5: step = %d, want 6", m.step)
	}

	// Selection should be "vault" (the original value)
	if m.radioIndex != 0 {
		t.Errorf("after re-enter: radioIndex = %d, want 0 (vault)", m.radioIndex)
	}
	if m.config.SecretsBackend != "vault" {
		t.Errorf("after re-enter: SecretsBackend = %q, want %q", m.config.SecretsBackend, "vault")
	}
}

// TestRadioSelectionCommittedOnEnter verifies that pressing Enter commits
// the radio selection, and navigating back shows the committed value.
func TestRadioSelectionCommittedOnEnter(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	// Default is "vault" (index 0)
	if m.radioIndex != 0 {
		t.Fatalf("initial radioIndex = %d, want 0 (vault)", m.radioIndex)
	}

	// Move to "dev" (index 1)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.radioIndex != 1 {
		t.Fatalf("after Down: radioIndex = %d, want 1 (dev)", m.radioIndex)
	}

	// Press Enter to commit — "dev" has no conditional fields, advances to step 7
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != 7 {
		t.Fatalf("after Enter: step = %d, want 7", m.step)
	}
	if m.config.SecretsBackend != "dev" {
		t.Fatalf("after Enter: SecretsBackend = %q, want %q", m.config.SecretsBackend, "dev")
	}

	// Go back to step 6
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.step != 6 {
		t.Fatalf("after Esc: step = %d, want 6", m.step)
	}

	// Selection should be "dev" (committed)
	if m.radioIndex != 1 {
		t.Errorf("after re-enter: radioIndex = %d, want 1 (dev)", m.radioIndex)
	}
	if m.config.SecretsBackend != "dev" {
		t.Errorf("after re-enter: SecretsBackend = %q, want %q", m.config.SecretsBackend, "dev")
	}
}

// TestStep6RadioDevAdvancesThenEscPreserves verifies the specific scenario:
// on step 6, select "dev", press Enter to advance, press Esc to return.
func TestStep6RadioDevAdvancesThenEscPreserves(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.step != 7 {
		t.Fatalf("expected step 7, got %d", m.step)
	}

	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.step != 6 {
		t.Fatalf("expected step 6, got %d", m.step)
	}

	if m.radioIndex != 1 {
		t.Errorf("radioIndex = %d, want 1 (dev)", m.radioIndex)
	}
	if m.config.SecretsBackend != "dev" {
		t.Errorf("SecretsBackend = %q, want %q", m.config.SecretsBackend, "dev")
	}
}

// TestStep10RadioSelectionPreservedOnReenter verifies step 10 (AI Agent) radio
// selection is preserved when navigating away and back.
func TestStep10RadioSelectionPreservedOnReenter(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 10)

	if m.radioIndex != 1 {
		t.Fatalf("initial radioIndex = %d, want 1 (no)", m.radioIndex)
	}

	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})

	if m.step != 10 {
		t.Fatalf("expected step 10, got %d", m.step)
	}
	if m.radioIndex != 1 {
		t.Errorf("radioIndex = %d, want 1 (no)", m.radioIndex)
	}
}

// TestStep14RadioSelectionPreservedOnReenter verifies step 14 (GitHub Repo
// Creation) radio selection is preserved when navigating away and back.
func TestStep14RadioSelectionPreservedOnReenter(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 14)

	if m.radioIndex != 1 {
		t.Fatalf("initial radioIndex = %d, want 1 (no)", m.radioIndex)
	}

	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})

	if m.step != 14 {
		t.Fatalf("expected step 14, got %d", m.step)
	}
	if m.radioIndex != 1 {
		t.Errorf("radioIndex = %d, want 1 (no)", m.radioIndex)
	}
}

// TestRadioSelectionChangeThenAdvanceMultipleStepsAndReturn verifies that radio
// selections are preserved even after navigating through multiple steps.
func TestRadioSelectionChangeThenAdvanceMultipleStepsAndReturn(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.radioIndex != 2 {
		t.Fatalf("radioIndex = %d, want 2 (kubernetes)", m.radioIndex)
	}

	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.step > 4 {
		for m.step > 4 {
			m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})
		}
	}

	if m.radioIndex != 2 {
		t.Errorf("radioIndex = %d, want 2 (kubernetes)", m.radioIndex)
	}
	if m.config.DeploymentMode != "kubernetes" {
		t.Errorf("DeploymentMode = %q, want %q", m.config.DeploymentMode, "kubernetes")
	}
}

// TestStep6VaultFieldsFilledThenNavigateBackAndForth verifies that both the
// radio selection and conditional field values are preserved.
func TestStep6VaultFieldsFilledThenNavigateBackAndForth(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	// Select vault and switch to text input
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.mode != modeTextInput {
		t.Fatalf("expected modeTextInput, got %d", m.mode)
	}

	// Fill in fields
	m.textInput.SetValue("http://vault:8200")
	m.textInputs[0] = m.textInput
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	m.textInput.SetValue("token")
	m.textInputs[1] = m.textInput

	// Advance
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != 7 {
		t.Fatalf("expected step 7, got %d", m.step)
	}

	// Go back
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.step != 6 {
		t.Fatalf("expected step 6, got %d", m.step)
	}

	if m.radioIndex != 0 {
		t.Errorf("radioIndex = %d, want 0 (vault)", m.radioIndex)
	}
	if m.config.VaultAddress != "http://vault:8200" {
		t.Errorf("VaultAddress = %q, want %q", m.config.VaultAddress, "http://vault:8200")
	}
	if m.config.VaultAuthMethod != "token" {
		t.Errorf("VaultAuthMethod = %q, want %q", m.config.VaultAuthMethod, "token")
	}
}

// TestStep6RadioViewShowsSelectionOnReenter verifies the view renders the
// correct radio selection marker when re-entering a step.
func TestStep6RadioViewShowsSelectionOnReenter(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})

	view := m.View()
	if !strings.Contains(view, "● dev") {
		t.Errorf("View should contain '● dev' (selected), got: %s", view)
	}
}

// TestStep10RadioChangeToYesThenBackPreserves verifies that selecting "yes"
// on step 10, filling in AI fields, advancing, and coming back preserves
// both the radio selection and the field values.
func TestStep10RadioChangeToYesThenBackPreserves(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 10)

	// Change from "no" to "yes"
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyUp})
	if m.radioIndex != 0 {
		t.Fatalf("radioIndex = %d, want 0 (yes)", m.radioIndex)
	}

	// Switch to text input mode
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.mode != modeTextInput {
		t.Fatalf("expected modeTextInput, got %d", m.mode)
	}

	// Fill in AI fields
	m.textInput.SetValue("/path/to/key")
	m.textInputs[0] = m.textInput
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	m.textInput.SetValue("https://api.openai.com/v1")
	m.textInputs[1] = m.textInput
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	m.textInput.SetValue("gpt-4o")
	m.textInputs[2] = m.textInput
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyTab})

	m.textInput.SetValue("default")
	m.textInputs[3] = m.textInput

	// Advance
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != 11 {
		t.Fatalf("expected step 11, got %d", m.step)
	}

	// Go back
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.step != 10 {
		t.Fatalf("expected step 10, got %d", m.step)
	}

	if m.radioIndex != 0 {
		t.Errorf("radioIndex = %d, want 0 (yes)", m.radioIndex)
	}
	if !m.config.AIEnabled {
		t.Errorf("AIEnabled = false, want true")
	}
}

// TestRadioSelectionChangeOnReenter verifies that the user CAN change the
// radio selection when re-entering a step (arrow keys work after re-entry).
func TestRadioSelectionChangeOnReenter(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// Advance and come back
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})

	// Default is "docker" (index 0)
	if m.radioIndex != 0 {
		t.Fatalf("radioIndex = %d, want 0", m.radioIndex)
	}

	// Change the selection on re-entry
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.radioIndex != 1 {
		t.Errorf("after Down: radioIndex = %d, want 1 (source)", m.radioIndex)
	}
	if m.config.DeploymentMode != "source" {
		t.Errorf("after Down: DeploymentMode = %q, want %q", m.config.DeploymentMode, "source")
	}
}

// TestStep7RadioSelectionPreservedOnReenter verifies step 7 (Redis) radio
// selection is preserved when navigating away and back.
func TestStep7RadioSelectionPreservedOnReenter(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 7)

	// Default is "local" (index 0)
	if m.radioIndex != 0 {
		t.Fatalf("initial radioIndex = %d, want 0 (local)", m.radioIndex)
	}

	// Advance
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})

	// Go back
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})

	if m.step != 7 {
		t.Fatalf("expected step 7, got %d", m.step)
	}
	if m.radioIndex != 0 {
		t.Errorf("radioIndex = %d, want 0 (local)", m.radioIndex)
	}
}

// TestStep7RadioChangeToExternalThenBack verifies that selecting "external" on
// step 7, filling in the connection string, advancing, and coming back
// preserves the selection and field value.
func TestStep7RadioChangeToExternalThenBack(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 7)

	// Change to "external"
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.radioIndex != 1 {
		t.Fatalf("radioIndex = %d, want 1 (external)", m.radioIndex)
	}

	// Switch to text input mode (external has conditional field)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.mode != modeTextInput {
		t.Fatalf("expected modeTextInput, got %d", m.mode)
	}

	// Fill in connection string
	m.textInput.SetValue("redis://external:6379")
	m.textInputs[0] = m.textInput

	// Advance
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != 8 {
		t.Fatalf("expected step 8, got %d", m.step)
	}

	// Go back
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.step != 7 {
		t.Fatalf("expected step 7, got %d", m.step)
	}

	if m.radioIndex != 1 {
		t.Errorf("radioIndex = %d, want 1 (external)", m.radioIndex)
	}
	if m.config.RedisMode != "external" {
		t.Errorf("RedisMode = %q, want %q", m.config.RedisMode, "external")
	}
	if m.config.RedisConnString != "redis://external:6379" {
		t.Errorf("RedisConnString = %q, want %q", m.config.RedisConnString, "redis://external:6379")
	}
}

// TestRadioEscRevertThenReenterOriginal verifies the complete flow: change
// radio selection, press Esc to revert, navigate forward again, and confirm
// the original selection is shown.
func TestRadioEscRevertThenReenterOriginal(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 6)

	// Verify initial: "vault" (index 0)
	if m.radioIndex != 0 {
		t.Fatalf("initial radioIndex = %d, want 0 (vault)", m.radioIndex)
	}

	// Move to "dev" (tentative)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.radioIndex != 1 {
		t.Fatalf("after Down: radioIndex = %d, want 1 (dev)", m.radioIndex)
	}

	// Press Esc to go back (reverts config)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.step != 5 {
		t.Fatalf("after Esc: step = %d, want 5", m.step)
	}

	// Config should be reverted to "vault"
	if m.config.SecretsBackend != "vault" {
		t.Errorf("config.SecretsBackend = %q, want %q (reverted)", m.config.SecretsBackend, "vault")
	}

	// Navigate forward through step 5 (multi-field) to reach step 6
	for m.step == 5 {
		m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})
	}
	if m.step != 6 {
		t.Fatalf("after navigating through step 5: step = %d, want 6", m.step)
	}

	// Selection should be "vault" (the original value)
	if m.radioIndex != 0 {
		t.Errorf("radioIndex = %d, want 0 (vault)", m.radioIndex)
	}

	// Verify the view shows vault selected
	view := m.View()
	if !strings.Contains(view, "● vault") {
		t.Errorf("View should contain '● vault' (selected), got: %s", view)
	}
}

// TestRadioMultipleArrowPressesThenEscRevert verifies that pressing arrow
// keys multiple times and then Esc reverts to the original selection.
func TestRadioMultipleArrowPressesThenEscRevert(t *testing.T) {
	cfg := config.NewDefaultWizardConfig()
	m := NewWizard(cfg)
	m = runInitStep(m, 4)

	// Default is "docker" (index 0)
	if m.radioIndex != 0 {
		t.Fatalf("initial radioIndex = %d, want 0 (docker)", m.radioIndex)
	}

	// Press Down twice to reach "kubernetes" (index 2)
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.radioIndex != 2 {
		t.Fatalf("radioIndex = %d, want 2 (kubernetes)", m.radioIndex)
	}
	if m.config.DeploymentMode != "kubernetes" {
		t.Fatalf("config.DeploymentMode = %q, want %q", m.config.DeploymentMode, "kubernetes")
	}

	// Press Esc to go back (reverts to "docker")
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.config.DeploymentMode != "docker" {
		t.Errorf("after Esc: DeploymentMode = %q, want %q (reverted)", m.config.DeploymentMode, "docker")
	}

	// Navigate forward again
	m, _ = updateWizard(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != 4 {
		t.Fatalf("after Enter: step = %d, want 4", m.step)
	}

	// Selection should be "docker" (the original value)
	if m.radioIndex != 0 {
		t.Errorf("radioIndex = %d, want 0 (docker)", m.radioIndex)
	}
	if m.config.DeploymentMode != "docker" {
		t.Errorf("DeploymentMode = %q, want %q", m.config.DeploymentMode, "docker")
	}
}
