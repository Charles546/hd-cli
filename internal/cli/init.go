// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If a separate written commercial agreement exists, you may use this file under those terms instead.

package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/Charles546/hd-cli/internal/config"
	"github.com/Charles546/hd-cli/internal/tui"
)

var (
	initDryRun         bool
	initConfigFile     string
	initNonInteractive bool
	initOutputDir      string
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Honeydipper project",
	Long:  `Initialize a new Honeydipper project with an interactive bootstrap wizard or a config file.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolVar(&initDryRun, "dry-run", false, "Print generated config files without writing them")
	initCmd.Flags().StringVar(&initConfigFile, "config", "", "Read wizard answers from a YAML file")
	initCmd.Flags().BoolVar(&initNonInteractive, "non-interactive", false, "Run in non-interactive mode (requires --config)")
	initCmd.Flags().StringVar(&initOutputDir, "output", "", "Override the output directory (default: from config or ./<project-name>)")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Mode 1: Config file + non-interactive → generate directly from file
	if initConfigFile != "" && initNonInteractive {
		return runNonInteractive()
	}

	// Mode 2: Config file without non-interactive → load answers, pre-populate, show TUI
	if initConfigFile != "" && !initNonInteractive {
		return runInteractiveWithConfig(initConfigFile)
	}

	// Mode 3: Non-interactive flag without config file — error
	if initNonInteractive {
		return fmt.Errorf("--non-interactive requires --config <answers-file>")
	}

	// Mode 4: Interactive TUI wizard with defaults
	return runInteractive()
}

// runInteractive starts the Bubble Tea TUI wizard with default config.
func runInteractive() error {
	cfg, err := tui.RunWizard()
	if err != nil {
		return fmt.Errorf("wizard error: %w", err)
	}

	return generateAndPrintSummary(cfg)
}

// runInteractiveWithConfig loads answers from a YAML file, pre-populates the
// wizard, and starts the TUI for review/modification.
func runInteractiveWithConfig(configPath string) error {
	// Load the answers file
	cfg, err := loadAnswersFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to load answers file: %w", err)
	}

	// Apply defaults for any missing values
	config.ApplyDefaults(cfg)

	// Start the TUI wizard with pre-populated config
	updatedCfg, err := tui.RunWizardWithConfig(cfg)
	if err != nil {
		return fmt.Errorf("wizard error: %w", err)
	}

	return generateAndPrintSummary(updatedCfg)
}

// generateAndPrintSummary generates config files and prints the post-generation summary.
func generateAndPrintSummary(cfg *config.WizardConfig) error {
	// Override output dir if flag is set
	outputDir := cfg.ConfigDir
	if initOutputDir != "" {
		outputDir = initOutputDir
	}

	// Confirm before writing
	fmt.Println()
	fmt.Printf("Config files will be written to: %s\n", outputDir)
	if initDryRun {
		fmt.Println("DRY RUN: No files will be written to disk.")
	}

	generator := config.NewGenerator()
	if err := generator.Generate(cfg, outputDir, initDryRun); err != nil {
		return fmt.Errorf("config generation failed: %w", err)
	}

	if !initDryRun {
		fmt.Printf("\nConfiguration files written to %s\n", outputDir)
		fmt.Println("\nNext steps:")
		fmt.Println("  1. Review the generated config files")
		fmt.Println("  2. Set up your secrets backend (Vault or environment variables)")
		fmt.Println("  3. Run 'hd deploy' to start Honeydipper")
	}

	return nil
}

// runNonInteractive reads answers from a YAML file and generates config without the TUI.
func runNonInteractive() error {
	// Validate the answers file first
	if err := config.ValidateAnswersFile(initConfigFile); err != nil {
		return fmt.Errorf("invalid answers file: %w", err)
	}

	// Load the answers to get ConfigDir (needed for default output directory)
	cfg, err := loadAnswersFile(initConfigFile)
	if err != nil {
		return fmt.Errorf("failed to load answers file: %w", err)
	}
	// Apply defaults in case ConfigDir is empty
	config.ApplyDefaults(cfg)

	generator := config.NewGenerator()

	// Determine output dir: use --output flag, or from answers file, or default
	outputDir := initOutputDir
	if outputDir == "" {
		outputDir = cfg.ConfigDir
	}

	if err := generator.GenerateFromAnswersFile(initConfigFile, outputDir, initDryRun); err != nil {
		return fmt.Errorf("config generation failed: %w", err)
	}

	if !initDryRun {
		fmt.Printf("Configuration files written to %s\n", outputDir)
	} else {
		fmt.Println("Dry run complete. No files were written.")
	}

	return nil
}

// loadAnswersFile reads a YAML file and unmarshals it into a WizardConfig.
func loadAnswersFile(path string) (*config.WizardConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read answers file: %w", err)
	}

	cfg := &config.WizardConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse answers file: %w", err)
	}

	return cfg, nil
}

// promptConfirmation asks the user to confirm an action.
// It is used for interactive confirmation prompts.
func promptConfirmation(msg string) bool {
	fmt.Printf("%s (yes/no): ", msg)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "yes" || response == "y"
}

// Ensure promptConfirmation is referenced to avoid unused error.
var _ = promptConfirmation
