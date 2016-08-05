// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"github.com/twmb/algoimpl/go/graph"
	"io/ioutil"
	"os"
	"os/exec"
)

// ComponentRef is the configuration block that references a component.
type ComponentRef struct {
	Name       string      `json:"name"`
	Version    string      `json:"version"`
	Repo       string      `json:"repo"`
	Repoconfig *RepoConfig `json:"repoconfig"`
	node       graph.Node
}

// Dependency is the configuration block that defines a dependency.
// There are three types of dependencies: build, runtime and intall
type Dependency struct {
	Build   []ComponentRef `json:"build"`
	Runtime []ComponentRef `json:"runtime"`
	Intall  []ComponentRef `json:"install"`
}

// RepoConfig defines the configuration for a repository
type RepoConfig struct {
	Type string `json:"type"`
	Base string `json:"base"`
}

// Project is the toplevel struct that represents a configuration file
type Project struct {
	ComponentRef
	Deps   Dependency
	graph  *graph.Graph
	sorted []graph.Node
}

func git(args []string) {
	log.Noticef("Executing: git %s\n", args)
	_, err := exec.Command("git", args...).Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			msg := string(ee.Stderr[:])
			log.Fatal("Error executing: ", msg)
		}

		log.Fatal(err)
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

// Fetch the specified component
func (comp ComponentRef) Fetch() {
	repo := resolveRepo(comp)
	args := []string{"clone", repo, comp.Name}
	git(args)
}

// Project methods

func parseProjectFile(filename string) (*Project, error) {
	var data []byte
	data, err := ioutil.ReadFile(filename)
	if ee, ok := err.(*os.PathError); ok {
		log.Error("Error: ", ee)
		return nil, err
	}

	var proj Project
	err = json.Unmarshal(data, &proj)
	return &proj, nil
}

func (proj *Project) processDeps() {
	proj.graph = graph.New(graph.Directed)
	proj.node = proj.graph.MakeNode()
	*proj.node.Value = proj

	// Build the dependency graph
	for _, dep := range proj.Deps.Build {
		log.Debug("Processing build dependency ", dep.Name)

		// Create dependency edge
		dep.node = proj.graph.MakeNode()
		*dep.node.Value = dep
		proj.graph.MakeEdge(proj.node, dep.node)

		if dep.Repoconfig == nil {
			log.Debug("Adding toplevel repoconfig to dep:", *proj.Repoconfig)
			dep.Repoconfig = proj.Repoconfig
		}
	}
}

// CompHandler is a callback to process a component handler
// type CompHandler func(*ComponentRef)

// Sort iterates all build dependencies
func (proj Project) Sort() {
	log.Debug("Sorting project ", proj.Name)
	proj.sorted = proj.graph.TopologicalSort()
}

// func (proj Project) (handler CompHandler) {
// 	for _, n := range proj.sorted {
// 		if cref, ok := (*n.Value).(ComponentRef); ok {
// 			log.Debug("Processing node :", cref.Name)
// 			handler(&cref)
// 		}
// 	}
// }
