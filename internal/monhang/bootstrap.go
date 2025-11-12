// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

package monhang

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

func runBoot(cmd *Command, args []string) {
	// Parse the toplevel project file
	proj, err := ParseProjectFile(getFilename())
	if err != nil {
		Check(err)
	}

	// Fetch toplevel component
	// proj.Fetch()
	proj.ProcessDeps()
	proj.Sort()
}

func init() {
	CmdBoot.Run = runBoot // break init loop
}
