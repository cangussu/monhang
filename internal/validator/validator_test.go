// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

package validator

import (
	"strings"
	"testing"
)

func TestCompileSchema(t *testing.T) {
	schema, err := CompileSchema()
	if err != nil {
		t.Fatalf("Failed to compile schema: %v", err)
	}

	if schema == nil {
		t.Fatal("Schema is nil")
	}

	// Test that schema is cached
	schema2, err := CompileSchema()
	if err != nil {
		t.Fatalf("Failed to compile schema second time: %v", err)
	}

	if schema != schema2 {
		t.Error("Schema should be cached and return same instance")
	}
}

func TestValidateJSON_Valid(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{
			name: "minimal valid config",
			json: `{"name": "test-component"}`,
		},
		{
			name: "full valid config",
			json: `{
				"name": "test-component",
				"version": "1.0.0",
				"source": "git://github.com/org/repo.git",
				"description": "Test component"
			}`,
		},
		{
			name: "with nested components",
			json: `{
				"name": "parent",
				"components": [
					{
						"name": "child1",
						"source": "https://github.com/org/child.git"
					},
					{
						"name": "child2"
					}
				]
			}`,
		},
		{
			name: "version with v prefix",
			json: `{"name": "test", "version": "v1.0.0"}`,
		},
		{
			name: "version without v prefix",
			json: `{"name": "test", "version": "1.0.0"}`,
		},
		{
			name: "version with prerelease",
			json: `{"name": "test", "version": "1.0.0-beta.1"}`,
		},
		{
			name: "https source",
			json: `{"name": "test", "source": "https://github.com/org/repo.git"}`,
		},
		{
			name: "file source",
			json: `{"name": "test", "source": "file:///path/to/repo.git"}`,
		},
		{
			name: "ssh source",
			json: `{"name": "test", "source": "ssh://git@github.com/org/repo.git"}`,
		},
		{
			name: "SSH git format (git@host:path)",
			json: `{"name": "test", "source": "git@github.com:monhang/monhang.git"}`,
		},
		{
			name: "SSH git format with query params",
			json: `{"name": "test", "source": "git@github.com:monhang/monhang.git?version=v1.0.0"}`,
		},
		{
			name: "empty components array",
			json: `{"name": "test", "components": []}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJSON([]byte(tt.json))
			if err != nil {
				t.Errorf("Expected valid JSON, got error: %v", err)
			}
		})
	}
}

func TestValidateJSON_Invalid(t *testing.T) {
	tests := []struct {
		name          string
		json          string
		expectedError string
	}{
		{
			name:          "invalid source URL no scheme",
			json:          `{"name": "test", "source": "github.com/org/repo"}`,
			expectedError: "does not match pattern",
		},
		{
			name:          "invalid source URL wrong scheme",
			json:          `{"name": "test", "source": "ftp://github.com/org/repo"}`,
			expectedError: "does not match pattern",
		},
		{
			name:          "invalid version format",
			json:          `{"name": "test", "version": "invalid"}`,
			expectedError: "does not match pattern",
		},
		{
			name:          "invalid name with spaces",
			json:          `{"name": "test component"}`,
			expectedError: "does not match pattern",
		},
		{
			name:          "invalid name with special chars",
			json:          `{"name": "test@component"}`,
			expectedError: "does not match pattern",
		},
		{
			name:          "additional unexpected property",
			json:          `{"name": "test", "unexpected": "value"}`,
			expectedError: "additional properties",
		},
		{
			name:          "empty name",
			json:          `{"name": ""}`,
			expectedError: "does not match pattern",
		},
		{
			name:          "wrong type for components",
			json:          `{"name": "test", "components": "not-an-array"}`,
			expectedError: "want array",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJSON([]byte(tt.json))
			if err == nil {
				t.Error("Expected validation error, got nil")
				return
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing %q, got: %v", tt.expectedError, err)
			}
		})
	}
}

func TestValidateTOML_Valid(t *testing.T) {
	tests := []struct {
		name string
		toml string
	}{
		{
			name: "minimal valid config",
			toml: `name = "test-component"`,
		},
		{
			name: "full valid config",
			toml: `
name = "test-component"
version = "1.0.0"
source = "git://github.com/org/repo.git"
description = "Test component"
`,
		},
		{
			name: "with nested components",
			toml: `
name = "parent"

[[components]]
name = "child1"
source = "https://github.com/org/child.git"

[[components]]
name = "child2"
`,
		},
		{
			name: "SSH git format source",
			toml: `
name = "test-component"
source = "git@github.com:monhang/monhang.git"
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTOML([]byte(tt.toml))
			if err != nil {
				t.Errorf("Expected valid TOML, got error: %v", err)
			}
		})
	}
}

func TestValidateTOML_Invalid(t *testing.T) {
	tests := []struct {
		name          string
		toml          string
		expectedError string
	}{
		{
			name: "invalid source URL",
			toml: `name = "test"
source = "invalid-url"`,
			expectedError: "does not match pattern",
		},
		{
			name:          "invalid name",
			toml:          `name = "test component"`,
			expectedError: "does not match pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTOML([]byte(tt.toml))
			if err == nil {
				t.Error("Expected validation error, got nil")
				return
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing %q, got: %v", tt.expectedError, err)
			}
		})
	}
}

// ValidateComponent is tested indirectly through ParseComponentFile tests
// which validate both JSON and TOML formats after parsing into Component structs.

func TestFormatValidationError(t *testing.T) {
	// Test nil error
	err := FormatValidationError(nil)
	if err != nil {
		t.Error("FormatValidationError(nil) should return nil")
	}

	// Test with validation error containing helpful messages
	jsonData := []byte(`{"name": "test", "version": "invalid"}`) // invalid version format
	validationErr := ValidateJSON(jsonData)

	if validationErr == nil {
		t.Fatal("Expected validation error for invalid version format")
	}

	errorMsg := validationErr.Error()
	if !strings.Contains(errorMsg, "configuration validation failed") {
		t.Errorf("Error message should contain 'configuration validation failed', got: %s", errorMsg)
	}
}
