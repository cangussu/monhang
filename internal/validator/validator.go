// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

package validator

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed schema.json
var schemaData []byte

var compiledSchema *jsonschema.Schema

// CompileSchema loads and compiles the embedded JSON schema.
// The schema is compiled once and cached for subsequent validations.
func CompileSchema() (*jsonschema.Schema, error) {
	if compiledSchema != nil {
		return compiledSchema, nil
	}

	compiler := jsonschema.NewCompiler()
	// Set draft to 2019-09 which is compatible with Draft-07 schemas
	compiler.DefaultDraft(jsonschema.Draft2019)

	// Parse the schema JSON first
	var schemaDoc interface{}
	if err := json.Unmarshal(schemaData, &schemaDoc); err != nil {
		return nil, fmt.Errorf("failed to parse embedded schema: %w", err)
	}

	if err := compiler.AddResource("schema.json", schemaDoc); err != nil {
		return nil, fmt.Errorf("failed to add schema resource: %w", err)
	}

	schema, err := compiler.Compile("schema.json")
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}

	compiledSchema = schema
	return compiledSchema, nil
}

// ValidateJSON validates raw JSON bytes against the schema.
func ValidateJSON(data []byte) error {
	schema, err := CompileSchema()
	if err != nil {
		return fmt.Errorf("schema compilation error: %w", err)
	}

	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("JSON parse error: %w", err)
	}

	if err := schema.Validate(v); err != nil {
		return FormatValidationError(err)
	}

	return nil
}

// ValidateTOML converts TOML to JSON and validates against the schema.
func ValidateTOML(data []byte) error {
	// Parse TOML into interface{}
	var v interface{}
	if err := toml.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("TOML parse error: %w", err)
	}

	// Convert to JSON for schema validation
	jsonData, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("TOML to JSON conversion error: %w", err)
	}

	return ValidateJSON(jsonData)
}

// ValidateComponent validates a Component struct against the schema.
func ValidateComponent(comp interface{}) error {
	// Convert struct to JSON
	jsonData, err := json.Marshal(comp)
	if err != nil {
		return fmt.Errorf("failed to marshal component: %w", err)
	}

	return ValidateJSON(jsonData)
}

// FormatValidationError converts jsonschema validation errors into user-friendly messages.
func FormatValidationError(err error) error {
	if err == nil {
		return nil
	}

	validationErr, ok := err.(*jsonschema.ValidationError)
	if !ok {
		return err
	}

	// Get the error string and format it
	errorMsg := validationErr.Error()

	// Add helpful hints based on common error patterns
	var hints []string
	if strings.Contains(errorMsg, "missing properties") {
		hints = append(hints, "Hint: Every component must have a 'name' field")
	}
	if strings.Contains(errorMsg, "does not match pattern") {
		switch {
		case strings.Contains(errorMsg, "source"):
			hints = append(hints, "Hint: Source URLs must start with git://, https://, file://, ssh://, or use SSH format (git@host:path)")
		case strings.Contains(errorMsg, "name"):
			hints = append(hints, "Hint: Names can only contain letters, numbers, underscores, and hyphens")
		case strings.Contains(errorMsg, "version"):
			hints = append(hints, "Hint: Versions must follow semantic versioning (e.g., v1.0.0 or 1.0.0)")
		}
	}
	if strings.Contains(errorMsg, "additionalProperties") {
		hints = append(hints, "Hint: Check for typos in field names")
	}

	// Build formatted error message
	var result strings.Builder
	result.WriteString("configuration validation failed:\n")
	result.WriteString("  ")
	result.WriteString(errorMsg)

	if len(hints) > 0 {
		result.WriteString("\n\n")
		for _, hint := range hints {
			result.WriteString("  ")
			result.WriteString(hint)
			result.WriteString("\n")
		}
	}

	return fmt.Errorf("%s", result.String())
}
