// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

// Package monhang provides component management functionality.
package components

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/cangussu/monhang/internal/logging"
	"github.com/cangussu/monhang/internal/validator"
)

const (
	// DefaultRepoType is the default repository type when not specified.
	DefaultRepoType = "git"
)

// Component represents a component in the workspace.
// This is a recursive structure where a component can have children components.
// The top-level manifest is itself a Component.
//
// The Source URL encodes both the repository location and metadata:
// - Schema determines the type: git://, https://, file:// all indicate git repositories
// - Query parameter ?type=<type> can override the schema-based type detection
// - Query parameter ?version=<version> specifies the version/tag/branch
// - Default type is "git" if not specified
//
// Examples:
//   - git://github.com/org/repo.git?version=v1.0.0
//   - https://github.com/org/repo.git?version=main
//   - file:///path/to/local/repo.git?version=v2.0.0
type Component struct {
	Source      string       `json:"source,omitempty" toml:"source,omitempty"`
	Name        string       `json:"name" toml:"name"`
	Description string       `json:"description,omitempty" toml:"description,omitempty"`
	Version     string       `json:"version,omitempty" toml:"version,omitempty"`
	Components  []*Component `json:"components,omitempty" toml:"components,omitempty"`
}

// HasRepo returns true if the component has a source configured.
func (comp *Component) HasRepo() bool {
	return comp.Source != ""
}

// parseSourceURL extracts the repository URL, version, and type from a component source.
// Source format examples:
//   - git://github.com/org/repo.git?version=v1.0.0&type=git
//   - https://github.com/org/repo.git?version=main
//   - file:///path/to/repo.git?version=v2.0.0
//
// Returns:
//   - repoURL: cleaned repository URL suitable for git clone
//   - version: version/tag/branch to checkout (from ?version param)
//   - repoType: repository type, determined by:
//     1. Query parameter ?type=<type> if present
//     2. URL scheme (git://, https://, file:// → "git")
//     3. Default: "git"
func parseSourceURL(source string) (repoURL, version, repoType string) {
	// Parse the URL
	u, err := url.Parse(source)
	if err != nil {
		// If parsing fails, return the source as-is with default type
		return source, "", DefaultRepoType
	}

	// Extract metadata from query parameters
	version = u.Query().Get("version")
	repoType = u.Query().Get("type")

	// If type not specified in query params, determine from scheme
	if repoType == "" {
		switch u.Scheme {
		case "git", "https", "http", "file", "ssh":
			repoType = DefaultRepoType
		default:
			repoType = DefaultRepoType // Default to git
		}
	}

	// Remove query string to get clean repo URL
	u.RawQuery = ""

	// Convert git:// to https:// for cloning
	repoURL = u.String()
	if strings.HasPrefix(repoURL, "git://") {
		repoURL = "https://" + strings.TrimPrefix(repoURL, "git://")
	}

	return repoURL, version, repoType
}

// ResolveRepo returns the full repository URL for this component.
// It parses the Source field and removes query parameters.
func (comp *Component) ResolveRepo() string {
	if comp.Source == "" {
		return ""
	}

	repoURL, _, _ := parseSourceURL(comp.Source)
	return repoURL
}

// GetVersion returns the version specified in the source URL.
// Returns empty string if no version is specified.
func (comp *Component) GetVersion() string {
	if comp.Source == "" {
		return comp.Version
	}

	_, version, _ := parseSourceURL(comp.Source)
	if version != "" {
		return version
	}
	return comp.Version
}

// GetType returns the repository type determined from the source URL.
// Returns "git" by default.
func (comp *Component) GetType() string {
	if comp.Source == "" {
		return DefaultRepoType
	}

	_, _, repoType := parseSourceURL(comp.Source)
	return repoType
}

// ParseComponentFile parses a component configuration file (JSON or TOML format).
// The file represents the top-level component manifest.
func ParseComponentFile(filename string) (*Component, error) {
	var data []byte
	// #nosec G304 -- filename is a config file path provided by the user
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var comp Component

	// Detect file format by extension
	ext := strings.ToLower(filepath.Ext(filename))

	logging.GetLogger("component").Debug().
		Str("filename", filename).
		Str("extension", ext).
		Msg("Parsing component file")

	// Validate raw data against schema BEFORE parsing
	var validationErr error
	if ext == ".toml" {
		validationErr = validator.ValidateTOML(data)
	} else {
		validationErr = validator.ValidateJSON(data)
	}

	if validationErr != nil {
		logging.GetLogger("component").Error().
			Err(validationErr).
			Str("filename", filename).
			Msg("Schema validation failed")
		return nil, fmt.Errorf("configuration validation failed: %w", validationErr)
	}

	if ext == ".toml" {
		// Parse TOML format
		err = toml.Unmarshal(data, &comp)
	} else {
		// Default to JSON format for .json or other extensions
		err = json.Unmarshal(data, &comp)
	}

	if err != nil {
		return nil, err
	}

	// Post-parse validation (double-check struct after unmarshaling)
	if structErr := validator.ValidateComponent(&comp); structErr != nil {
		logging.GetLogger("component").Error().
			Err(structErr).
			Str("filename", filename).
			Msg("Post-parse validation failed")
		return nil, fmt.Errorf("configuration validation failed: %w", structErr)
	}

	logging.GetLogger("component").Debug().
		Str("filename", filename).
		Str("name", comp.Name).
		Msg("Component file parsed and validated successfully")

	return &comp, nil
}

// CreateLocalComponent creates a component representing the current directory.
// This is used when no manifest file exists - the current directory is treated as a local component.
func CreateLocalComponent(name string) *Component {
	logging.GetLogger("component").Debug().Msg("No config file found, using current directory as local component")
	comp := &Component{
		Name: name,
	}
	return comp
}
