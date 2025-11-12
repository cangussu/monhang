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

func TestParseJSONConfig(t *testing.T) {
	proj, err := parseProjectFile("test/monhang.json")
	if err != nil {
		t.Fatalf("Failed to parse JSON config: %v", err)
	}

	if proj.Name != "top-app" {
		t.Errorf("Expected name 'top-app', got '%s'", proj.Name)
	}

	if proj.Version != "1.0.3" {
		t.Errorf("Expected version '1.0.3', got '%s'", proj.Version)
	}

	if proj.Repo != "monhang.git" {
		t.Errorf("Expected repo 'monhang.git', got '%s'", proj.Repo)
	}

	if proj.Repoconfig == nil {
		t.Fatal("Expected repoconfig to be set")
	}

	if proj.Repoconfig.Type != "git" {
		t.Errorf("Expected repoconfig type 'git', got '%s'", proj.Repoconfig.Type)
	}

	if proj.Repoconfig.Base != "git@github.com:monhang/" {
		t.Errorf("Expected repoconfig base 'git@github.com:monhang/', got '%s'", proj.Repoconfig.Base)
	}

	if len(proj.Deps.Build) != 2 {
		t.Errorf("Expected 2 build dependencies, got %d", len(proj.Deps.Build))
	}

	if len(proj.Deps.Build) > 0 {
		if proj.Deps.Build[0].Name != "lib1" {
			t.Errorf("Expected first dep name 'lib1', got '%s'", proj.Deps.Build[0].Name)
		}
		if proj.Deps.Build[0].Version != "v1.0.0" {
			t.Errorf("Expected first dep version 'v1.0.0', got '%s'", proj.Deps.Build[0].Version)
		}
	}
}

func TestParseTOMLConfig(t *testing.T) {
	proj, err := parseProjectFile("test/monhang.toml")
	if err != nil {
		t.Fatalf("Failed to parse TOML config: %v", err)
	}

	if proj.Name != "top-app" {
		t.Errorf("Expected name 'top-app', got '%s'", proj.Name)
	}

	if proj.Version != "1.0.3" {
		t.Errorf("Expected version '1.0.3', got '%s'", proj.Version)
	}

	if proj.Repo != "monhang.git" {
		t.Errorf("Expected repo 'monhang.git', got '%s'", proj.Repo)
	}

	if proj.Repoconfig == nil {
		t.Fatal("Expected repoconfig to be set")
	}

	if proj.Repoconfig.Type != "git" {
		t.Errorf("Expected repoconfig type 'git', got '%s'", proj.Repoconfig.Type)
	}

	if proj.Repoconfig.Base != "git@github.com:monhang/" {
		t.Errorf("Expected repoconfig base 'git@github.com:monhang/', got '%s'", proj.Repoconfig.Base)
	}

	if len(proj.Deps.Build) != 2 {
		t.Errorf("Expected 2 build dependencies, got %d", len(proj.Deps.Build))
	}

	if len(proj.Deps.Build) > 0 {
		if proj.Deps.Build[0].Name != "lib1" {
			t.Errorf("Expected first dep name 'lib1', got '%s'", proj.Deps.Build[0].Name)
		}
		if proj.Deps.Build[0].Version != "v1.0.0" {
			t.Errorf("Expected first dep version 'v1.0.0', got '%s'", proj.Deps.Build[0].Version)
		}
	}
}
