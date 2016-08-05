// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

package main

var cmdBoot = &Command{
	Name:  "boot",
	Args:  "[configfile]",
	Short: "bootstrap a component and its dependencies",
	Long: `
Boot fetches and setups the workspace for the component described in the given configuration file.
`,
}

var bootF = cmdBoot.Flag.String("f", "<defaultconfig>", "configuration file")

func getFilename() string {
	if *bootF != "<defaultconfig>" {
		return *bootF
	}
	return "./monhang.json"
}

func runBoot(cmd *Command, args []string) {
	// Parse the toplevel project file
	proj, err := parseProjectFile(getFilename())
	if err != nil {
		check(err)
	}

	// Fetch toplevel component
	// proj.Fetch()
	proj.processDeps()
	proj.Sort()
}

func init() {
	cmdBoot.Run = runBoot // break init loop
}
