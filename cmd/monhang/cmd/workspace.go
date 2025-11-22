// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"

	"github.com/cangussu/monhang/internal/commands"
	"github.com/cangussu/monhang/internal/components"
	"github.com/cangussu/monhang/internal/logging"
	"github.com/spf13/cobra"
)

var workspaceConfigFile string

// workspaceCmd represents the workspace command.
var workspaceCmd = &cobra.Command{
	Use:     "workspace",
	Aliases: []string{"ws"},
	Short:   "Manage workspace components",
	Long: `Workspace manages components defined in the configuration file.

Available subcommands:
  sync    synchronize workspace components from manifest

Examples:
  monhang workspace sync
  monhang ws sync
  monhang workspace sync -f custom-config.json`,
	Run: runWorkspace,
}

func init() {
	rootCmd.AddCommand(workspaceCmd)
	workspaceCmd.Flags().StringVarP(&workspaceConfigFile, "file", "f", "./monhang.json", "configuration file")
}

func runWorkspace(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("Error: No subcommand specified")
		fmt.Println("Usage: monhang workspace <subcommand> [args]")
		fmt.Println("\nSubcommands: sync")
		os.Exit(1)
	}

	subcommand := args[0]

	logging.GetLogger("workspace").Info().
		Str("subcommand", subcommand).
		Msg("Starting workspace command")

	// Execute based on subcommand
	switch subcommand {
	case "sync":
		handleWorkspaceSync(workspaceConfigFile)
	default:
		logging.GetLogger("workspace").Error().Str("subcommand", subcommand).Msg("Unknown workspace subcommand")
		fmt.Printf("Error: Unknown subcommand '%s'\n", subcommand)
		fmt.Println("Available subcommands: sync")
		os.Exit(1)
	}

	logging.GetLogger("workspace").Info().Str("subcommand", subcommand).Msg("Workspace command completed")
}

// handleWorkspaceSync processes the sync subcommand.
func handleWorkspaceSync(filename string) {
	logging.GetLogger("workspace").Info().Str("file", filename).Msg("Starting workspace sync")

	proj, err := components.ParseProjectFile(filename)
	Check(err)

	if len(proj.Components) == 0 {
		logging.GetLogger("workspace").Info().Msg("No components defined in manifest")
		fmt.Println("No components defined in manifest")
		return
	}

	// Collect results
	results := &commands.SyncResults{}

	// Collect all components (flatten tree)
	allComponents := flattenComponents(proj.Components)

	// Run interactive sync
	if err := commands.RunInteractiveSync(filename, allComponents, results); err != nil {
		fmt.Printf("Error running interactive sync: %v\n", err)
		os.Exit(1)
	}

	logging.GetLogger("workspace").Info().Msg("Workspace sync completed")
}

// flattenComponents flattens the component tree into a list.
func flattenComponents(comps []components.Component) []components.Component {
	// Pre-allocate with estimated capacity
	result := make([]components.Component, 0, len(comps)*2)
	for _, comp := range comps {
		result = append(result, comp)
		if len(comp.Children) > 0 {
			result = append(result, flattenComponents(comp.Children)...)
		}
	}
	return result
}
