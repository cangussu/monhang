package components

import (
	"testing"
)

//nolint:gocognit,gocyclo // Test validation function with comprehensive checks
func validateComponentConfig(t *testing.T, comp *Component) {
	t.Helper()

	if comp.Name != "top-app" {
		t.Errorf("Expected name 'top-app', got '%s'", comp.Name)
	}

	if comp.Version != "1.0.3" {
		t.Errorf("Expected version '1.0.3', got '%s'", comp.Version)
	}

	if comp.Source != "git://github.com/monhang/monhang.git?version=v1.0.3" {
		t.Errorf("Expected source 'git://github.com/monhang/monhang.git?version=v1.0.3', got '%s'", comp.Source)
	}

	if comp.Description != "Top-level application" {
		t.Errorf("Expected description 'Top-level application', got '%s'", comp.Description)
	}

	// Validate child components
	if len(comp.Components) != 2 {
		t.Errorf("Expected 2 child components, got %d", len(comp.Components))
	}

	if len(comp.Components) > 0 {
		// First component
		child := comp.Components[0]
		if child.Name != "core" {
			t.Errorf("Expected first component name 'core', got '%s'", child.Name)
		}
		if child.Description != "Core library component" {
			t.Errorf("Expected first component description 'Core library component', got '%s'", child.Description)
		}
		if child.Source != "git://github.com/monhang/core.git?version=v1.0.0&type=git" {
			t.Errorf("Expected first component source 'git://github.com/monhang/core.git?version=v1.0.0&type=git', got '%s'", child.Source)
		}

		// First component's child
		if len(child.Components) != 1 {
			t.Errorf("Expected first component to have 1 child, got %d", len(child.Components))
		}
		if len(child.Components) > 0 {
			grandchild := child.Components[0]
			if grandchild.Name != "utils" {
				t.Errorf("Expected child component name 'utils', got '%s'", grandchild.Name)
			}
			if grandchild.Description != "Utility functions" {
				t.Errorf("Expected child component description 'Utility functions', got '%s'", grandchild.Description)
			}
			if grandchild.Source != "git://github.com/monhang/utils.git?version=v2.1.0&type=git" {
				t.Errorf("Expected child component source 'git://github.com/monhang/utils.git?version=v2.1.0&type=git', got '%s'", grandchild.Source)
			}
		}
	}

	if len(comp.Components) > 1 {
		// Second component
		child := comp.Components[1]
		if child.Name != "plugin" {
			t.Errorf("Expected second component name 'plugin', got '%s'", child.Name)
		}
		if child.Description != "Plugin system" {
			t.Errorf("Expected second component description 'Plugin system', got '%s'", child.Description)
		}
		if child.Source != "git://github.com/monhang/plugin.git?version=v3.0.1&type=git" {
			t.Errorf("Expected second component source 'git://github.com/monhang/plugin.git?version=v3.0.1&type=git', got '%s'", child.Source)
		}
		if len(child.Components) != 0 {
			t.Errorf("Expected second component to have 0 children, got %d", len(child.Components))
		}
	}
}

func TestParseJSONConfig(t *testing.T) {
	comp, err := ParseComponentFile("../testdata/monhang.json")
	if err != nil {
		t.Fatalf("Failed to parse JSON config: %v", err)
	}

	validateComponentConfig(t, comp)
}

func TestParseTOMLConfig(t *testing.T) {
	comp, err := ParseComponentFile("../testdata/monhang.toml")
	if err != nil {
		t.Fatalf("Failed to parse TOML config: %v", err)
	}

	validateComponentConfig(t, comp)
}

//nolint:govet // Component struct field alignment is acceptable for test readability
func TestResolveRepo(t *testing.T) {
	tests := []struct {
		name     string
		comp     Component
		expected string
	}{
		{
			name: "git:// URL should convert to https://",
			comp: Component{
				Source: "git://github.com/org/repo.git?version=v1.0.0",
			},
			expected: "https://github.com/org/repo.git",
		},
		{
			name: "https:// URL should be preserved",
			comp: Component{
				Source: "https://github.com/org/repo.git?version=v1.0.0",
			},
			expected: "https://github.com/org/repo.git",
		},
		{
			name: "file:// URL should be preserved",
			comp: Component{
				Source: "file:///home/user/repos/myrepo.git?version=v1.0.0",
			},
			expected: "file:///home/user/repos/myrepo.git",
		},
		{
			name: "URL without version query param",
			comp: Component{
				Source: "https://github.com/org/repo.git",
			},
			expected: "https://github.com/org/repo.git",
		},
		{
			name: "empty source returns empty string",
			comp: Component{
				Source: "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.comp.ResolveRepo()
			if result != tt.expected {
				t.Errorf("ResolveRepo() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

//nolint:govet // Component struct field alignment is acceptable for test readability
func TestGetVersion(t *testing.T) {
	tests := []struct {
		name     string
		comp     Component
		expected string
	}{
		{
			name: "version from source URL",
			comp: Component{
				Source:  "git://github.com/org/repo.git?version=v1.0.0",
				Version: "v2.0.0",
			},
			expected: "v1.0.0",
		},
		{
			name: "version from Version field when source has no version",
			comp: Component{
				Source:  "git://github.com/org/repo.git",
				Version: "v2.0.0",
			},
			expected: "v2.0.0",
		},
		{
			name: "version from Version field when no source",
			comp: Component{
				Version: "v3.0.0",
			},
			expected: "v3.0.0",
		},
		{
			name: "empty version",
			comp: Component{
				Source: "git://github.com/org/repo.git",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.comp.GetVersion()
			if result != tt.expected {
				t.Errorf("GetVersion() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

//nolint:govet // Component struct field alignment is acceptable for test readability
func TestGetType(t *testing.T) {
	tests := []struct {
		name     string
		comp     Component
		expected string
	}{
		{
			name: "type from query parameter",
			comp: Component{
				Source: "git://github.com/org/repo.git?type=svn",
			},
			expected: "svn",
		},
		{
			name: "type from git:// scheme",
			comp: Component{
				Source: "git://github.com/org/repo.git",
			},
			expected: "git",
		},
		{
			name: "type from https:// scheme",
			comp: Component{
				Source: "https://github.com/org/repo.git",
			},
			expected: "git",
		},
		{
			name: "type from file:// scheme",
			comp: Component{
				Source: "file:///path/to/repo.git",
			},
			expected: "git",
		},
		{
			name:     "default type when no source",
			comp:     Component{},
			expected: "git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.comp.GetType()
			if result != tt.expected {
				t.Errorf("GetType() = %q, expected %q", result, tt.expected)
			}
		})
	}
}
