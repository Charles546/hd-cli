// Copyright 2026 Chun Huang (Charles).
//
// This Source Code Form is dual-licensed.
// By default, this file is licensed under the GNU Affero General Public License v3.0.
// If you have a separate written commercial agreement, you may use this file under those terms instead.

package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Run executes a command and returns combined stdout and stderr output
func Run(cmd string, args ...string) (string, error) {
	command := exec.Command(cmd, args...)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	output := stdout.String() + stderr.String()

	if err != nil {
		return output, fmt.Errorf("command failed: %s %s: %w", cmd, strings.Join(args, " "), err)
	}

	return output, nil
}

// RunStreaming executes a command with stdout and stderr streamed to terminal
func RunStreaming(cmd string, args ...string) error {
	command := exec.Command(cmd, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err := command.Run()
	if err != nil {
		return fmt.Errorf("command failed: %s %s: %w", cmd, strings.Join(args, " "), err)
	}

	return nil
}

// RunInDir executes a command in a specific directory
func RunInDir(dir, cmd string, args ...string) (string, error) {
	command := exec.Command(cmd, args...)
	command.Dir = dir

	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	output := stdout.String() + stderr.String()

	if err != nil {
		return output, fmt.Errorf("command failed in %s: %s %s: %w", dir, cmd, strings.Join(args, " "), err)
	}

	return output, nil
}

// RunInDirStreaming executes a command in a specific directory with streaming output
func RunInDirStreaming(dir, cmd string, args ...string) error {
	command := exec.Command(cmd, args...)
	command.Dir = dir
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err := command.Run()
	if err != nil {
		return fmt.Errorf("command failed in %s: %s %s: %w", dir, cmd, strings.Join(args, " "), err)
	}

	return nil
}

// IsCommandAvailable checks if a binary exists in $PATH
func IsCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// RunWithInput executes a command with stdin input
func RunWithInput(cmd string, args ...string) (string, string, error) {
	command := exec.Command(cmd, args...)
	
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	return stdout.String(), stderr.String(), err
}

// RunWithStdin executes a command with custom stdin
func RunWithStdin(stdin io.Reader, cmd string, args ...string) (string, string, error) {
	command := exec.Command(cmd, args...)
	command.Stdin = stdin

	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	return stdout.String(), stderr.String(), err
}
