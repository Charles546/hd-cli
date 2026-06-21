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

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy [docker|source|kubernetes]",
	Short: "Deploy and manage Honeydipper instances",
	Long:  `Deploy Honeydipper using Docker, from source, or on Kubernetes.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deploymentType := args[0]
		fmt.Printf("hd deploy %s: not yet implemented\n", deploymentType)
		fmt.Printf("This command will deploy Honeydipper using %s.\n", deploymentType)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"docker", "source", "kubernetes"}, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
