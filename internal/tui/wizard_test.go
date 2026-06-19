// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If a separate written commercial agreement exists, you may use this file under those terms instead.

package tui

import (
	"fmt"
	"testing"

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
