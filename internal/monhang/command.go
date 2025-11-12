// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

package monhang

import (
	"flag"
)

// Command is an implementation of a godep command
// like godep save or godep go.
type Command struct {
	// Run runs the command.
	// The args are the arguments after the command name.
	Run func(cmd *Command, args []string)

	// Name of the command
	Name string

	// Args the command would expect
	Args string

	// Short is the short description shown in the 'godep help' output.
	Short string

	// Long is the long message shown in the
	// 'godep help <this-command>' output.
	Long string

	// Flag is a set of flags specific to this command.
	Flag flag.FlagSet
}

// Check panics if there's an error
func Check(e error) {
	if e != nil {
		panic(e)
	}
}
