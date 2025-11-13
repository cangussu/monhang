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

	"github.com/cangussu/monhang/internal/monhang"
	"github.com/op/go-logging"
)

var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

func setupLog() {
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	backendLeveled := logging.AddModuleLevel(backend)
	backendLeveled.SetLevel(logging.DEBUG, "")
	logging.SetBackend(backendFormatter)
}

func version() {
	fmt.Println("monhang v0.0.1")
}

func usageExit() {
	version()
	fmt.Println(`Usage:

	monhang command [arguments]

The commands are:

	boot        bootstraps a workspace
	version     print monhang version

Use "monhang help [command]" for more information about a command.`)
	os.Exit(0)
}

var cmdHelp = &monhang.Command{
	Name: "help",
	Run: func(_ *monhang.Command, _ []string) {
		// TODO(cangussu): print the help for the command given in args
		usageExit()
	},
}

var commands = []*monhang.Command{
	monhang.CmdBoot,
	cmdHelp,
}

func init() {
	setupLog()
}

func main() {
	flag.Usage = usageExit
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("You must tell monhang what to do!")
		usageExit()
	}

	for _, cmd := range commands {
		if cmd.Name == args[0] {
			if err := cmd.Flag.Parse(args[1:]); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
				os.Exit(1)
			}
			cmd.Run(cmd, cmd.Flag.Args())
		}
	}
}
