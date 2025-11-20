// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

// Package monhang provides component management functionality.
package monhang

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/cangussu/monhang/internal/logging"
	"github.com/twmb/algoimpl/go/graph"
)

// ComponentRef is the configuration block that references a component.
type ComponentRef struct {
	Repoconfig *RepoConfig `json:"repoconfig" toml:"repoconfig"`
	node       graph.Node
	Name       string `json:"name" toml:"name"`
	Version    string `json:"version" toml:"version"`
	Repo       string `json:"repo" toml:"repo"`
}

// Dependency is the configuration block that defines a dependency.
// There are three types of dependencies: build, runtime and install.
type Dependency struct {
	Build   []ComponentRef `json:"build" toml:"build"`
	Runtime []ComponentRef `json:"runtime" toml:"runtime"`
	Intall  []ComponentRef `json:"install" toml:"install"`
}

// RepoConfig defines the configuration for a repository.
type RepoConfig struct {
	Type string `json:"type" toml:"type"`
	Base string `json:"base" toml:"base"`
}

// Project is the toplevel struct that represents a configuration file.
type Project struct {
	ComponentRef
	Deps   Dependency
	graph  *graph.Graph
	sorted []graph.Node
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
	if ee, ok := err.(*os.PathError); ok {
		logging.GetLogger("component").Error().Err(ee).Str("filename", filename).Msg("Failed to read project file")
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

	if err != nil {
		logging.GetLogger("component").Error().Err(err).Str("filename", filename).Msg("Failed to parse project file")
	}

	return &proj, err
}

// ProcessDeps builds the dependency graph for the project.
func (proj *Project) ProcessDeps() {
	proj.graph = graph.New(graph.Directed)
	proj.node = proj.graph.MakeNode()
	*proj.node.Value = proj

	logging.GetLogger("component").Debug().Str("project", proj.Name).Msg("Processing project dependencies")

	// Build the dependency graph
	for _, dep := range proj.Deps.Build {
		logging.GetLogger("component").Debug().Str("dependency", dep.Name).Str("type", "build").Msg("Processing dependency")

		// Set repoconfig if not specified
		if dep.Repoconfig == nil {
			logging.GetLogger("component").Debug().
				Str("dependency", dep.Name).
				Str("base", proj.Repoconfig.Base).
				Msg("Adding toplevel repoconfig to dependency")
			dep.Repoconfig = proj.Repoconfig
		}

		// Create dependency edge
		dep.node = proj.graph.MakeNode()
		*dep.node.Value = dep
		if err := proj.graph.MakeEdge(proj.node, dep.node); err != nil {
			logging.GetLogger("component").Error().Err(err).Str("dependency", dep.Name).Msg("Failed to create edge")
		}
	}

	// Process runtime dependencies
	for _, dep := range proj.Deps.Runtime {
		logging.GetLogger("component").Debug().Str("dependency", dep.Name).Str("type", "runtime").Msg("Processing dependency")

		// Set repoconfig if not specified
		if dep.Repoconfig == nil {
			logging.GetLogger("component").Debug().
				Str("dependency", dep.Name).
				Str("base", proj.Repoconfig.Base).
				Msg("Adding toplevel repoconfig to dependency")
			dep.Repoconfig = proj.Repoconfig
		}

		// Create dependency edge
		dep.node = proj.graph.MakeNode()
		*dep.node.Value = dep
		if err := proj.graph.MakeEdge(proj.node, dep.node); err != nil {
			logging.GetLogger("component").Error().Err(err).Str("dependency", dep.Name).Msg("Failed to create edge")
		}
	}

	logging.GetLogger("component").Debug().Str("project", proj.Name).Int("total_deps", len(proj.Deps.Build)+len(proj.Deps.Runtime)).Msg("Finished processing dependencies")
}

// Sort iterates all build dependencies.
func (proj *Project) Sort() {
	logging.GetLogger("component").Debug().Str("project", proj.Name).Msg("Sorting project dependencies")
	proj.sorted = proj.graph.TopologicalSort()
	logging.GetLogger("component").Debug().Str("project", proj.Name).Int("sorted_count", len(proj.sorted)).Msg("Dependencies sorted")
}
