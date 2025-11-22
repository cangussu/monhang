/*
   Monhang - component management tool
   Copyright (C) 2016  Thiago Cangussu de Castro Gomes

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/cangussu/monhang/internal/logging"
	"github.com/spf13/cobra"
)

var (
	debugFlag bool
	// Version is the current version of monhang.
	Version = "0.0.1"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "monhang",
	Short: "Component management tool for monorepos",
	Long: `Monhang is a component management tool designed for managing multi-repository workspaces.

It helps you bootstrap, synchronize, and manage components across multiple repositories,
making it easier to work with complex monorepo structures.`,
	Version: Version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setupLogging()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "enable debug logging")
}

func setupLogging() {
	// Check for debug mode from environment variable or CLI flag
	debug := debugFlag || strings.ToLower(os.Getenv("MONHANG_DEBUG")) == "true"

	// Initialize zerolog
	logging.Initialize(debug, os.Stderr)
}

// Check panics if there's an error.
func Check(e error) {
	if e != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", e)
		os.Exit(1)
	}
}
