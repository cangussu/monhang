package main

import (
	"testing"
)

func TestGitFetch(t *testing.T) {
	// Duck typing git:
	var givenArgs []string
	git = func(args []string) {
		givenArgs = args
	}

	ref := ComponentRef{
		Name:    "teste1",
		Repo:    "this.that",
		Version: "1.0.0",
	}

	ref.Fetch()
	if len(givenArgs) > 3 {
		mglog.Error("invalid number of arguments:", len(givenArgs), givenArgs)
	}

	if givenArgs[0] != "clone" {
		mglog.Error("invalid git command:", givenArgs[0])
	}

	if givenArgs[1] != ref.Repo {
		mglog.Error("invalid URL:", givenArgs[1])
	}

	if givenArgs[2] != ref.Name {
		mglog.Error("invalid repo name:", givenArgs[2])
	}
}
