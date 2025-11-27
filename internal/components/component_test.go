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
		{
			name: "SSH format URL should be preserved",
			comp: Component{
				Source: "git@github.com:monhang/monhang.git?version=v1.0.0",
			},
			expected: "git@github.com:monhang/monhang.git",
		},
		{
			name: "SSH format URL without query params",
			comp: Component{
				Source: "git@github.com:monhang/monhang.git",
			},
			expected: "git@github.com:monhang/monhang.git",
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
		{
			name: "version from SSH format source URL",
			comp: Component{
				Source:  "git@github.com:monhang/monhang.git?version=v1.2.3",
				Version: "v2.0.0",
			},
			expected: "v1.2.3",
		},
		{
			name: "SSH format with no version falls back to Version field",
			comp: Component{
				Source:  "git@github.com:monhang/monhang.git",
				Version: "v3.0.0",
			},
			expected: "v3.0.0",
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
		{
			name: "type from SSH format defaults to git",
			comp: Component{
				Source: "git@github.com:monhang/monhang.git",
			},
			expected: "git",
		},
		{
			name: "type from SSH format with explicit type parameter",
			comp: Component{
				Source: "git@github.com:monhang/monhang.git?type=svn",
			},
			expected: "svn",
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

//nolint:dupl // Table-driven tests have similar structure but test different scenarios
func TestDeriveNameFromURL_StandardURLs(t *testing.T) {
	tests := []struct {
		name     string
		repoURL  string
		expected string
	}{
		{
			name:     "HTTPS URL with .git suffix",
			repoURL:  "https://github.com/org/repo.git",
			expected: "repo",
		},
		{
			name:     "HTTPS URL without .git suffix",
			repoURL:  "https://github.com/org/repo",
			expected: "repo",
		},
		{
			name:     "URL with trailing slash and .git",
			repoURL:  "https://github.com/org/repo.git/",
			expected: "repo",
		},
		{
			name:     "URL with trailing slash without .git",
			repoURL:  "https://github.com/org/repo/",
			expected: "repo",
		},
		{
			name:     "file:// URL with .git",
			repoURL:  "file:///home/user/repos/myrepo.git",
			expected: "myrepo",
		},
		{
			name:     "file:// URL without .git",
			repoURL:  "file:///home/user/repos/myrepo",
			expected: "myrepo",
		},
		{
			name:     "git protocol URL",
			repoURL:  "git://github.com/org/repo.git",
			expected: "repo",
		},
		{
			name:     "SSH with port number",
			repoURL:  "ssh://git@github.com:22/org/repo.git",
			expected: "repo",
		},
		{
			name:     "complex path with multiple levels",
			repoURL:  "https://example.com/path/to/deep/repo.git",
			expected: "repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deriveNameFromURL(tt.repoURL)
			if result != tt.expected {
				t.Errorf("deriveNameFromURL(%q) = %q, expected %q", tt.repoURL, result, tt.expected)
			}
		})
	}
}

func TestDeriveNameFromURL_SSHFormat(t *testing.T) {
	tests := []struct {
		name     string
		repoURL  string
		expected string
	}{
		{
			name:     "SSH format with .git suffix",
			repoURL:  "git@host.xz:foo/.git",
			expected: "foo",
		},
		{
			name:     "SSH format without .git suffix",
			repoURL:  "git@host.xz:foo",
			expected: "foo",
		},
		{
			name:     "SSH format with nested path and .git",
			repoURL:  "git@github.com:org/repo.git",
			expected: "repo",
		},
		{
			name:     "SSH format with nested path without .git",
			repoURL:  "git@github.com:org/repo",
			expected: "repo",
		},
		{
			name:     "SSH format with just .git directory",
			repoURL:  "git@host.xz:.git",
			expected: "component",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deriveNameFromURL(tt.repoURL)
			if result != tt.expected {
				t.Errorf("deriveNameFromURL(%q) = %q, expected %q", tt.repoURL, result, tt.expected)
			}
		})
	}
}

func TestDeriveNameFromURL_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		repoURL  string
		expected string
	}{
		{
			name:     "repo name with hyphens",
			repoURL:  "https://github.com/org/my-repo-name.git",
			expected: "my-repo-name",
		},
		{
			name:     "repo name with underscores",
			repoURL:  "https://github.com/org/my_repo_name.git",
			expected: "my_repo_name",
		},
		{
			name:     "repo name with dots",
			repoURL:  "https://github.com/org/my.repo.name.git",
			expected: "my.repo.name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deriveNameFromURL(tt.repoURL)
			if result != tt.expected {
				t.Errorf("deriveNameFromURL(%q) = %q, expected %q", tt.repoURL, result, tt.expected)
			}
		})
	}
}

//nolint:dupl // Table-driven tests have similar structure but test different scenarios
func TestDeriveNameFromURL_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		repoURL  string
		expected string
	}{
		{
			name:     "empty string returns default",
			repoURL:  "",
			expected: "component",
		},
		{
			name:     "just dot returns default",
			repoURL:  ".",
			expected: "component",
		},
		{
			name:     "URL ending with only .git",
			repoURL:  "https://example.com/.git",
			expected: "component",
		},
		{
			name:     "file path with .git suffix",
			repoURL:  "/path/to/repo.git",
			expected: "repo",
		},
		{
			name:     "file path without .git suffix",
			repoURL:  "/path/to/repo",
			expected: "repo",
		},
		{
			name:     "relative path with .git",
			repoURL:  "../repo.git",
			expected: "repo",
		},
		{
			name:     "current directory relative path",
			repoURL:  "./repo.git",
			expected: "repo",
		},
		{
			name:     "Windows-style path",
			repoURL:  "C:\\Users\\user\\repos\\repo.git",
			expected: "repo",
		},
		{
			name:     "single word repo",
			repoURL:  "myrepo",
			expected: "myrepo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deriveNameFromURL(tt.repoURL)
			if result != tt.expected {
				t.Errorf("deriveNameFromURL(%q) = %q, expected %q", tt.repoURL, result, tt.expected)
			}
		})
	}
}

//nolint:govet // Component struct field alignment is acceptable for test readability
func TestGetName(t *testing.T) {
	tests := []struct {
		name     string
		comp     Component
		expected string
	}{
		{
			name: "explicit name is used",
			comp: Component{
				Name:   "explicit-name",
				Source: "https://github.com/org/repo.git",
			},
			expected: "explicit-name",
		},
		{
			name: "name derived from HTTPS source",
			comp: Component{
				Source: "https://github.com/org/derived-repo.git",
			},
			expected: "derived-repo",
		},
		{
			name: "name derived from SSH source",
			comp: Component{
				Source: "git@github.com:org/ssh-repo.git",
			},
			expected: "ssh-repo",
		},
		{
			name: "name derived from file source",
			comp: Component{
				Source: "file:///path/to/file-repo.git",
			},
			expected: "file-repo",
		},
		{
			name:     "default name when no source or name",
			comp:     Component{},
			expected: "component",
		},
		{
			name: "empty name uses derived name",
			comp: Component{
				Name:   "",
				Source: "https://github.com/org/derived.git",
			},
			expected: "derived",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.comp.GetName()
			if result != tt.expected {
				t.Errorf("GetName() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestParseComponentFile_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		expectError bool
	}{
		{
			name:        "missing name is now valid (derived from source)",
			filename:    "../testdata/invalid/missing-name.json",
			expectError: false,
		},
		{
			name:        "invalid source URL",
			filename:    "../testdata/invalid/invalid-source-url.json",
			expectError: true,
		},
		{
			name:        "invalid version format",
			filename:    "../testdata/invalid/invalid-version-format.json",
			expectError: true,
		},
		{
			name:        "extra fields",
			filename:    "../testdata/invalid/extra-fields.json",
			expectError: true,
		},
		{
			name:        "nested component without name is now valid (derived from source)",
			filename:    "../testdata/invalid/invalid-nested-component.json",
			expectError: false,
		},
		{
			name:        "valid JSON config",
			filename:    "../testdata/monhang.json",
			expectError: false,
		},
		{
			name:        "valid TOML config",
			filename:    "../testdata/monhang.toml",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseComponentFile(tt.filename)
			if tt.expectError && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}
