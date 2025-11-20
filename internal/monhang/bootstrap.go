// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

package monhang

import (
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
	proj, err := ParseProjectFile(filename)
	if err != nil {
		Check(err)
	}

	logging.GetLogger("bootstrap").Info().Str("project", proj.Name).Str("version", proj.Version).Msg("Project loaded")

	// Process and sort dependencies
	proj.ProcessDeps()
	proj.Sort()

	// Fetch all dependencies
	logging.GetLogger("bootstrap").Info().Msg("Fetching dependencies...")
	fetchedCount := 0
	for _, node := range proj.sorted {
		if node.Value == nil {
			continue
		}
		comp, ok := (*node.Value).(ComponentRef)
		if !ok {
			logging.GetLogger("bootstrap").Warn().Msg("Skipping non-ComponentRef node")
			continue
		}
		logging.GetLogger("bootstrap").Info().Str("component", comp.Name).Str("version", comp.Version).Msg("Fetching component")
		comp.Fetch()
		fetchedCount++
	}
	logging.GetLogger("bootstrap").Info().Int("count", fetchedCount).Msg("All dependencies fetched successfully")
}

func init() {
	CmdBoot.Run = runBoot // break init loop
}
