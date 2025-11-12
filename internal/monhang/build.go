// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

package monhang

var CmdBuild = &Command{
	Name:  "build",
	Args:  "[components...]",
	Short: "builds given components",
	Long: `
Builds the given components. If none, builds all dependencies and toplevel
component.
`,
}

func runBuild(cmd *Command, args []string) {
	// TODO(tgomes): load the workspace configuration
	var config Project

	// Topologically sort the dependencies and build
	config.Sort()
}

func init() {
	CmdBuild.Run = runBuild // break init loop
}
