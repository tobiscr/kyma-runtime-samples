// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and kyma-runtime-extension-samples
// SPDX-License-Identifier: Apache-2.0

// Package cmd implements the CLI commands for the auditlog tool.
package cmd

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

// config holds the global CLI configuration shared across all commands.
var config = struct {
	ServiceBindingFile string
	serviceBinding     serviceBinding
}{}

// newRootCmd creates and returns the root cobra command.
// It registers the --bindingFile flag and loads the service binding before any subcommand runs.
func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "auditlog",
		Short: "Access audit logs from SAP Audit Log Service v2",
		Long: `This CLI tool helps you to access
	audit logs from SAP Audit Log Service v2.
	You can use it to query and retrieve audit logs based on various criteria.`,
	}
	rootCmd.PersistentFlags().StringVarP(&config.ServiceBindingFile, "bindingFile", "b", "./servicebinding.json", "Path to the service binding file")

	// Load the service binding file before executing any subcommand.
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := readServiceBinding(); err != nil {
			return err
		}
		return nil
	}

	rootCmd.AddCommand(newGetCmd())
	return rootCmd
}

// Execute is the entry point called by main. It builds the command tree and runs the CLI.
func Execute() {
	err := newRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

// readServiceBinding reads and parses the JSON service binding file into the global config.
func readServiceBinding() error {
	data, err := os.ReadFile(config.ServiceBindingFile)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &config.serviceBinding); err != nil {
		return err
	}
	return err
}
