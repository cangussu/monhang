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

// Package main is the entry point for the monhang command-line tool.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/cangussu/monhang/internal/commands"
	"github.com/cangussu/monhang/internal/logging"
)

var (
	debugFlag = flag.Bool("debug", false, "Enable debug logging")
)

func setupLog() {
	// Check for debug mode from environment variable or CLI flag
	debug := *debugFlag || strings.ToLower(os.Getenv("MONHANG_DEBUG")) == "true"

	// Initialize zerolog
	logging.Initialize(debug, os.Stderr)
}

func version() {
	fmt.Println("monhang v0.0.1")
}

func usageExit() {
	version()
	fmt.Println(`Usage:

	monhang command [arguments]

The commands are:

	exec        run arbitrary commands inside each repo
	git         run git operations across all repos
	workspace   manage workspace components (alias: ws)
	version     print monhang version

Use "monhang help [command]" for more information about a command.`)
	os.Exit(0)
}

var cmdHelp = &commands.Command{
	Name: "help",
	Run: func(_ *commands.Command, _ []string) {
		// TODO(cangussu): print the help for the command given in args
		usageExit()
	},
}

var cmds = []*commands.Command{
	commands.CmdExec,
	commands.CmdGit,
	commands.CmdWorkspace,
	commands.CmdWs,
	cmdHelp,
}

func main() {
	flag.Usage = usageExit
	flag.Parse()

	// Setup logging after parsing flags
	setupLog()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("You must tell monhang what to do!")
		usageExit()
	}

	for _, cmd := range cmds {
		if cmd.Name == args[0] {
			if err := cmd.Flag.Parse(args[1:]); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
				os.Exit(1)
			}
			cmd.Run(cmd, cmd.Flag.Args())
		}
	}
}
