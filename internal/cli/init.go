// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If you have a separate written commercial agreement, you may use this file under those terms instead.

package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Honeydipper project",
	Long:  `Initialize a new Honeydipper project with interactive bootstrap wizard.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hd init: not yet implemented")
		fmt.Println("This command will guide you through setting up a new Honeydipper project.")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
