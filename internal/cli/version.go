// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If you have a separate written commercial agreement, you may use this file under those terms instead.

package cli

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// Version information - set via Go linker flags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print the CLI version, build commit, build date, and attempt to detect daemon version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("CLI Version:    %s\n", Version)
		fmt.Printf("Build Commit:   %s\n", Commit)
		fmt.Printf("Build Date:     %s\n", BuildDate)

		// Try to detect daemon version
		daemonVersion := detectDaemonVersion()
		if daemonVersion != "" {
			fmt.Printf("Daemon Version: %s\n", daemonVersion)
		} else {
			fmt.Println("Daemon Version: not running or not detected")
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

// detectDaemonVersion attempts to detect the running daemon version
func detectDaemonVersion() string {
	// Try to get daemon version using the CLI itself
	// This is a placeholder - in a real implementation, you might:
	// - Check a PID file
	// - Connect to a daemon API endpoint
	// - Run a subprocess command
	
	// For now, try to run 'hd version' if daemon is available
	cmd := exec.Command("hd", "version")
	output, err := cmd.Output()
	if err == nil {
		// Parse output to extract daemon version
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Daemon Version:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					return strings.TrimSpace(parts[1])
				}
			}
		}
	}
	
	return ""
}
