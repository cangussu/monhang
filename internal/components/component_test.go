package components

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
		t.Errorf("invalid number of arguments: %d %v", len(givenArgs), givenArgs)
	}

	if givenArgs[0] != "clone" {
		t.Errorf("invalid git command: %s", givenArgs[0])
	}

	if givenArgs[1] != ref.Repo {
		t.Errorf("invalid URL: %s", givenArgs[1])
	}

	if givenArgs[2] != ref.Name {
		t.Errorf("invalid repo name: %s", givenArgs[2])
	}
}

//nolint:gocognit,gocyclo // Test validation function with comprehensive checks
func validateProjectConfig(t *testing.T, proj *Project) {
	t.Helper()

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

	// Validate components
	if len(proj.Components) != 2 {
		t.Errorf("Expected 2 components, got %d", len(proj.Components))
	}

	if len(proj.Components) > 0 {
		// First component
		comp := proj.Components[0]
		if comp.Name != "core" {
			t.Errorf("Expected first component name 'core', got '%s'", comp.Name)
		}
		if comp.Description != "Core library component" {
			t.Errorf("Expected first component description 'Core library component', got '%s'", comp.Description)
		}
		if comp.Source != "git://github.com/monhang/core.git?version=v1.0.0&type=git" {
			t.Errorf("Expected first component source 'git://github.com/monhang/core.git?version=v1.0.0&type=git', got '%s'", comp.Source)
		}

		// First component's child
		if len(comp.Children) != 1 {
			t.Errorf("Expected first component to have 1 child, got %d", len(comp.Children))
		}
		if len(comp.Children) > 0 {
			child := comp.Children[0]
			if child.Name != "utils" {
				t.Errorf("Expected child component name 'utils', got '%s'", child.Name)
			}
			if child.Description != "Utility functions" {
				t.Errorf("Expected child component description 'Utility functions', got '%s'", child.Description)
			}
			if child.Source != "git://github.com/monhang/utils.git?version=v2.1.0&type=git" {
				t.Errorf("Expected child component source 'git://github.com/monhang/utils.git?version=v2.1.0&type=git', got '%s'", child.Source)
			}
		}
	}

	if len(proj.Components) > 1 {
		// Second component
		comp := proj.Components[1]
		if comp.Name != "plugin" {
			t.Errorf("Expected second component name 'plugin', got '%s'", comp.Name)
		}
		if comp.Description != "Plugin system" {
			t.Errorf("Expected second component description 'Plugin system', got '%s'", comp.Description)
		}
		if comp.Source != "git://github.com/monhang/plugin.git?version=v3.0.1&type=git" {
			t.Errorf("Expected second component source 'git://github.com/monhang/plugin.git?version=v3.0.1&type=git', got '%s'", comp.Source)
		}
		if len(comp.Children) != 0 {
			t.Errorf("Expected second component to have 0 children, got %d", len(comp.Children))
		}
	}
}

func TestParseJSONConfig(t *testing.T) {
	proj, err := ParseProjectFile("../testdata/monhang.json")
	if err != nil {
		t.Fatalf("Failed to parse JSON config: %v", err)
	}

	validateProjectConfig(t, proj)
}

func TestParseTOMLConfig(t *testing.T) {
	proj, err := ParseProjectFile("../testdata/monhang.toml")
	if err != nil {
		t.Fatalf("Failed to parse TOML config: %v", err)
	}

	validateProjectConfig(t, proj)
}
