// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

package commands

import (
	"github.com/cangussu/monhang/internal/components"
	"github.com/cangussu/monhang/internal/logging"
)

// CmdBoot is the boot command for bootstrapping a workspace.
var CmdBoot = &Command{
	Name:  "boot",
	Args:  "[configfile]",
	Short: "bootstrap a component and its dependencies",
	Long: `
Boot fetches and setups the workspace for the component described in the given configuration file.
`,
}

var bootF = CmdBoot.Flag.String("f", "<defaultconfig>", "configuration file")

func getFilename() string {
	if *bootF != "<defaultconfig>" {
		return *bootF
	}
	return "./monhang.json"
}

func runBoot(_ *Command, _ []string) {
	filename := getFilename()
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

func init() {
	CmdBoot.Run = runBoot // break init loop
}
