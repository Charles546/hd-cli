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
	initCmd.Flags().StringVar(&initConfigFile, "config", "", "Read wizard answers from a YAML file (non-interactive mode)")
	initCmd.Flags().BoolVar(&initNonInteractive, "non-interactive", false, "Run in non-interactive mode (requires --config)")
	initCmd.Flags().StringVar(&initOutputDir, "output", "", "Override the output directory (default: from config or ./<project-name>)")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Mode 1: Non-interactive with config file
	if initConfigFile != "" {
		return runNonInteractive()
	}

	// Mode 2: Non-interactive flag without config file — error
	if initNonInteractive {
		return fmt.Errorf("--non-interactive requires --config <answers-file>")
	}

	// Mode 3: Interactive TUI wizard
	return runInteractive()
}

// runInteractive starts the Bubble Tea TUI wizard.
func runInteractive() error {
	cfg, err := tui.RunWizard()
	if err != nil {
		return fmt.Errorf("wizard error: %w", err)
	}

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

// runNonInteractive reads answers from a YAML file and generates config.
func runNonInteractive() error {
	// Validate the answers file first
	if err := config.ValidateAnswersFile(initConfigFile); err != nil {
		return fmt.Errorf("invalid answers file: %w", err)
	}

	generator := config.NewGenerator()

	// Determine output dir
	outputDir := initOutputDir
	if outputDir == "" {
		outputDir = "./hd-config"
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
