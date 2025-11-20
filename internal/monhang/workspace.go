// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

package monhang

import (
	"fmt"
	"os"

	"github.com/cangussu/monhang/internal/logging"
)

// CmdWorkspace is the workspace command for managing workspace components.
var CmdWorkspace = &Command{
	Name:  "workspace",
	Args:  "<subcommand> [args...]",
	Short: "manage workspace components",
	Long: `
Workspace manages components defined in the configuration file.

Subcommands:
	sync    synchronize workspace components from manifest

Examples:
	monhang workspace sync
	monhang ws sync

Options:
	-f <file>    configuration file (default: ./monhang.json)
`,
}

// CmdWs is the alias for CmdWorkspace.
var CmdWs = &Command{
	Name:  "ws",
	Args:  "<subcommand> [args...]",
	Short: "alias for workspace",
	Long: `
Workspace (ws) manages components defined in the configuration file.

Subcommands:
	sync    synchronize workspace components from manifest

Examples:
	monhang ws sync
	monhang workspace sync

Options:
	-f <file>    configuration file (default: ./monhang.json)
`,
}

var (
	workspaceF = CmdWorkspace.Flag.String("f", "./monhang.json", "configuration file")
	wsF        = CmdWs.Flag.String("f", "./monhang.json", "configuration file")
)

// handleWorkspaceSync processes the sync subcommand.
func handleWorkspaceSync(filename string) {
	logging.GetLogger("workspace").Info().Str("file", filename).Msg("Starting workspace sync")

	proj, err := ParseProjectFile(filename)
	Check(err)

	if len(proj.Components) == 0 {
		logging.GetLogger("workspace").Info().Msg("No components defined in manifest")
		fmt.Println("No components defined in manifest")
		return
	}

	fmt.Printf("Syncing %d component(s) from %s\n", len(proj.Components), filename)

	// Process each component in the tree
	for _, comp := range proj.Components {
		syncComponent(comp, 0)
	}

	logging.GetLogger("workspace").Info().Msg("Workspace sync completed")
	fmt.Println("\nWorkspace sync completed!")
}

// syncComponent synchronizes a component and its children.
func syncComponent(comp Component, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	logging.GetLogger("workspace").Debug().
		Str("name", comp.Name).
		Str("source", comp.Source).
		Int("depth", depth).
		Msg("Processing component")

	fmt.Printf("%s- %s\n", indent, comp.Name)
	if comp.Description != "" {
		fmt.Printf("%s  Description: %s\n", indent, comp.Description)
	}
	fmt.Printf("%s  Source: %s\n", indent, comp.Source)

	// Process children recursively
	for _, child := range comp.Children {
		syncComponent(child, depth+1)
	}
}

func runWorkspace(_ *Command, args []string) {
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
		handleWorkspaceSync(*workspaceF)
	default:
		logging.GetLogger("workspace").Error().Str("subcommand", subcommand).Msg("Unknown workspace subcommand")
		fmt.Printf("Error: Unknown subcommand '%s'\n", subcommand)
		fmt.Println("Available subcommands: sync")
		os.Exit(1)
	}

	logging.GetLogger("workspace").Info().Str("subcommand", subcommand).Msg("Workspace command completed")
}

func runWs(_ *Command, args []string) {
	if len(args) == 0 {
		fmt.Println("Error: No subcommand specified")
		fmt.Println("Usage: monhang ws <subcommand> [args]")
		fmt.Println("\nSubcommands: sync")
		os.Exit(1)
	}

	subcommand := args[0]

	logging.GetLogger("workspace").Info().
		Str("subcommand", subcommand).
		Str("alias", "ws").
		Msg("Starting workspace command (via ws alias)")

	// Execute based on subcommand
	switch subcommand {
	case "sync":
		handleWorkspaceSync(*wsF)
	default:
		logging.GetLogger("workspace").Error().Str("subcommand", subcommand).Msg("Unknown workspace subcommand")
		fmt.Printf("Error: Unknown subcommand '%s'\n", subcommand)
		fmt.Println("Available subcommands: sync")
		os.Exit(1)
	}

	logging.GetLogger("workspace").Info().Str("subcommand", subcommand).Msg("Workspace command completed (via ws alias)")
}

func init() {
	CmdWorkspace.Run = runWorkspace
	CmdWs.Run = runWs
}
