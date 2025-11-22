// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

package cmd

import (
	"github.com/cangussu/monhang/internal/components"
	"github.com/cangussu/monhang/internal/logging"
	"github.com/spf13/cobra"
)

var bootConfigFile string

// bootCmd represents the boot command.
var bootCmd = &cobra.Command{
	Use:   "boot [configfile]",
	Short: "Bootstrap a component and its dependencies",
	Long: `Boot fetches and sets up the workspace for the component described in the given configuration file.

Examples:
  monhang boot
  monhang boot -f custom-config.json
  monhang boot path/to/monhang.json`,
	Run: runBoot,
}

func init() {
	rootCmd.AddCommand(bootCmd)
	bootCmd.Flags().StringVarP(&bootConfigFile, "file", "f", "./monhang.json", "configuration file")
}

func runBoot(cmd *cobra.Command, args []string) {
	filename := bootConfigFile

	// If a positional argument is provided, use that instead
	if len(args) > 0 {
		filename = args[0]
	}

	logging.GetLogger("bootstrap").Info().Str("config", filename).Msg("Starting bootstrap process")

	// Parse the toplevel project file
	proj, err := components.ParseProjectFile(filename)
	if err != nil {
		logging.GetLogger("bootstrap").Error().Err(err).Str("filename", filename).Msg("Failed to parse project file")
		Check(err)
	}

	logging.GetLogger("bootstrap").Info().Str("project", proj.Name).Str("version", proj.Version).Msg("Project loaded")

	// Fetch the main component
	logging.GetLogger("bootstrap").Info().Msg("Fetching component...")
	proj.Fetch()
	logging.GetLogger("bootstrap").Info().Msg("Component fetched successfully")
}
