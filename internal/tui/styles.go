// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If you have a separate written commercial agreement, you may use this file under those terms instead.

package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Color palette
const (
	primaryColor   = "#6366f1" // indigo
	secondaryColor = "#8b5cf6" // violet
	accentColor    = "#06b6d4" // cyan
	successColor   = "#22c55e" // green
	warningColor   = "#f59e0b" // amber
	errorColor     = "#ef4444" // red
	mutedColor     = "#6b7280" // gray
	textColor      = "#f3f4f6" // light
)

var (
	// TitleStyle is used for the main wizard title
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(primaryColor)).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(primaryColor)).
			Padding(1, 2).
			MarginBottom(1)

	// StepTitleStyle is used for individual step titles
	StepTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(secondaryColor)).
			MarginBottom(1)

	// DescriptionStyle is used for descriptive text
	DescriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(mutedColor)).
				MarginBottom(1)

	// LabelStyle is used for input labels
	LabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(textColor)).
			MarginBottom(0)

	// InputStyle is used for text inputs
	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(primaryColor)).
			Padding(0, 1).
			MarginBottom(1)

	// FocusedInputStyle is used for focused text inputs
	FocusedInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(accentColor)).
				Padding(0, 1).
				MarginBottom(1)

	// SelectedItemStyle is used for radio/option selections
	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(accentColor)).
				Bold(true)

	// UnselectedItemStyle is used for unselected options
	UnselectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(mutedColor))

	// HelpStyle is used for help text and hints
	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(mutedColor)).
			Italic(true).
			MarginTop(1)

	// ErrorStyle is used for error messages
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(errorColor)).
			Bold(true)

	// SuccessStyle is used for success messages
	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(successColor)).
			Bold(true)

	// WarningStyle is used for warning messages
	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(warningColor))

	// SummaryBoxStyle is used for the summary/review screen
	SummaryBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(primaryColor)).
				Padding(1, 2).
				MarginTop(1).
				MarginBottom(1)

	// ProgressBarStyle is used for the step progress indicator
	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(primaryColor))

	// ProgressBarEmptyStyle is used for unfilled progress
	ProgressBarEmptyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(mutedColor))

	// WelcomeStyle is used for the welcome screen
	WelcomeStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color(secondaryColor)).
			Padding(2, 4).
			MarginBottom(2)

	// ConfirmStyle is used for confirmation prompts
	ConfirmStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(warningColor)).
			MarginTop(1).
			MarginBottom(1)
)

// ProgressBar generates a visual progress bar showing current step
func ProgressBar(current, total int) string {
	if total == 0 {
		return ""
	}
	filled := (current * 20) / total
	bar := ""
	for i := 0; i < 20; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return ProgressBarStyle.Render(bar) + ProgressBarEmptyStyle.Render("") +
		lipgloss.NewStyle().Foreground(lipgloss.Color(mutedColor)).
			Render(fmt.Sprintf(" Step %d/%d", current, total))
}
