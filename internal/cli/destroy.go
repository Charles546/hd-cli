// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If you have a separate written commercial agreement, you may use this file under those terms instead.

package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Tear down Honeydipper instance",
	Long:  `Tear down and remove a Honeydipper instance with confirmation.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("WARNING: This will tear down your Honeydipper instance.")
		fmt.Print("Are you sure you want to continue? (yes/no): ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response == "yes" || response == "y" {
			fmt.Println("hd destroy: not yet implemented")
			fmt.Println("This command will tear down the Honeydipper instance.")
		} else {
			fmt.Println("Destroy cancelled.")
		}
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
