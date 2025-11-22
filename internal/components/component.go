// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

// Package monhang provides component management functionality.
package components

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/cangussu/monhang/internal/logging"
)

// ComponentRef is the configuration block that references a component.
type ComponentRef struct {
	Repoconfig *RepoConfig `json:"repoconfig" toml:"repoconfig"`
	Name       string      `json:"name" toml:"name"`
	Version    string      `json:"version" toml:"version"`
	Repo       string      `json:"repo" toml:"repo"`
}

// RepoConfig defines the configuration for a repository.
type RepoConfig struct {
	Type string `json:"type" toml:"type"`
	Base string `json:"base" toml:"base"`
}

// Component represents a component in the workspace.
// The Source URL encodes the version and type (schema).
type Component struct {
	Source      string      `json:"source" toml:"source"`
	Name        string      `json:"name" toml:"name"`
	Description string      `json:"description" toml:"description"`
	Children    []Component `json:"children,omitempty" toml:"children,omitempty"`
}

// Project is the toplevel struct that represents a configuration file.
type Project struct {
	ComponentRef
	Components []Component `json:"components,omitempty" toml:"components,omitempty"`
}

var git = func(args []string) {
	logging.GetLogger("component").Info().Strs("args", args).Msg("Executing git command")
	_, err := exec.Command("git", args...).Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			msg := string(ee.Stderr)
			logging.GetLogger("component").Fatal().Str("stderr", msg).Msg("Error executing git command")
		}

		logging.GetLogger("component").Fatal().Err(err).Msg("Error executing git command")
	}
}

func resolveRepo(comp ComponentRef) string {
	var repo string
	if comp.Repoconfig != nil {
		repo = comp.Repoconfig.Base + comp.Repo
	} else {
		repo = comp.Repo
	}
	return repo
}

// Fetch the specified component.
func (comp ComponentRef) Fetch() {
	repo := resolveRepo(comp)
	args := []string{"clone", repo, comp.Name}
	git(args)
}

// Project methods

// ParseProjectFile parses a project configuration file (JSON or TOML format).
func ParseProjectFile(filename string) (*Project, error) {
	var data []byte
	// #nosec G304 -- filename is a config file path provided by the user
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var proj Project

	// Detect file format by extension
	ext := strings.ToLower(filepath.Ext(filename))

	logging.GetLogger("component").Debug().Str("filename", filename).Str("extension", ext).Msg("Parsing project file")

	if ext == ".toml" {
		// Parse TOML format
		err = toml.Unmarshal(data, &proj)
	} else {
		// Default to JSON format for .json or other extensions
		err = json.Unmarshal(data, &proj)
	}

	return &proj, err
}

// CreateLocalProject creates a project representing the current directory as a git repository.
func CreateLocalProject(name string) *Project {
	logging.GetLogger("git").Debug().Msg("No config file found, using current directory as git repository")
	// Create a minimal project with just the current directory
	// Use "." as the name so getRepos() will find it in the current directory
	proj := &Project{
		ComponentRef: ComponentRef{
			Name: name,
		},
	}
	return proj
}
