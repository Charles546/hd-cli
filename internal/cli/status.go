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

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check daemon status",
	Long:  `Check the status of the running Honeydipper daemon.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hd status: not yet implemented")
		fmt.Println("This command will check the daemon status.")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
