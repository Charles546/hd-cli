// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If you have a separate written commercial agreement, you may use this file under those terms instead.

package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	verbose   bool
	configDir string
	output    string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hd",
	Short: "Honeydipper CLI - Deploy, configure, and manage Honeydipper instances",
	Long: `hd is a command-line interface for Honeydipper, an event-driven,
rule-based orchestration platform. Use it to deploy, configure, and
manage Honeydipper instances.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVar(&configDir, "config-dir", "", "Configuration directory")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "text", "Output format (json or text)")
}
