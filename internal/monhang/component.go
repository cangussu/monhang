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
	"github.com/op/go-logging"
	"github.com/twmb/algoimpl/go/graph"
)

var mglog = logging.MustGetLogger("monhang")

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
	mglog.Noticef("Executing: git %s\n", args)
	_, err := exec.Command("git", args...).Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			msg := string(ee.Stderr)
			mglog.Fatal("Error executing: ", msg)
		}

		mglog.Fatal(err)
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
		mglog.Error("Error: ", ee)
		return nil, err
	}

	var proj Project

	// Detect file format by extension
	ext := strings.ToLower(filepath.Ext(filename))

	if ext == ".toml" {
		// Parse TOML format
		err = toml.Unmarshal(data, &proj)
	} else {
		// Default to JSON format for .json or other extensions
		err = json.Unmarshal(data, &proj)
	}

	return &proj, err
}

// ProcessDeps builds the dependency graph for the project.
func (proj *Project) ProcessDeps() {
	proj.graph = graph.New(graph.Directed)
	proj.node = proj.graph.MakeNode()
	*proj.node.Value = proj

	// Build the dependency graph
	for _, dep := range proj.Deps.Build {
		mglog.Debug("Processing build dependency ", dep.Name)

		// Create dependency edge
		dep.node = proj.graph.MakeNode()
		*dep.node.Value = dep
		if err := proj.graph.MakeEdge(proj.node, dep.node); err != nil {
			mglog.Error("Failed to create edge: ", err)
		}

		if dep.Repoconfig == nil {
			mglog.Debug("Adding toplevel repoconfig to dep:", *proj.Repoconfig)
			dep.Repoconfig = proj.Repoconfig
		}
	}
}

// Sort iterates all build dependencies.
func (proj *Project) Sort() {
	mglog.Debug("Sorting project ", proj.Name)
	proj.sorted = proj.graph.TopologicalSort()
}
