// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

package commands

import (
	"testing"

	"github.com/cangussu/monhang/internal/components"
)

const (
	testComponentCoreName        = "core"
	testComponentCoreSource      = "git://github.com/monhang/core.git?version=v1.0.0&type=git"
	testComponentCoreDescription = "Core library component"
	testComponentUtilsName       = "utils"
	testComponentUtilsSource     = "git://github.com/monhang/utils.git?version=v2.1.0&type=git"
	testComponentUtilsDesc       = "Utility functions"
	testComponentPluginName      = "plugin"
)

func TestFlattenComponents(t *testing.T) {
	comps := []*components.Component{
		{
			Name:        testComponentCoreName,
			Source:      testComponentCoreSource,
			Description: testComponentCoreDescription,
			Components: []*components.Component{
				{
					Name:        testComponentUtilsName,
					Source:      testComponentUtilsSource,
					Description: testComponentUtilsDesc,
				},
			},
		},
		{
			Name:        testComponentPluginName,
			Source:      "git://github.com/monhang/plugin.git?version=v3.0.1&type=git",
			Description: "Plugin system",
		},
	}

	flattened := flattenComponents(comps)

	// Should have 3 components: core, utils (child), plugin
	if len(flattened) != 3 {
		t.Errorf("Expected 3 flattened components, got %d", len(flattened))
	}

	// Verify order: core, utils, plugin
	if flattened[0].Name != testComponentCoreName {
		t.Errorf("Expected first component to be '%s', got '%s'", testComponentCoreName, flattened[0].Name)
	}
	if flattened[1].Name != testComponentUtilsName {
		t.Errorf("Expected second component to be '%s', got '%s'", testComponentUtilsName, flattened[1].Name)
	}
	if flattened[2].Name != testComponentPluginName {
		t.Errorf("Expected third component to be '%s', got '%s'", testComponentPluginName, flattened[2].Name)
	}
}

func TestParseSourceURL(t *testing.T) {
	tests := []struct {
		source      string
		wantURL     string
		wantVersion string
		wantType    string
	}{
		{
			source:      "git://github.com/monhang/core.git?version=v1.0.0&type=git",
			wantURL:     "https://github.com/monhang/core.git",
			wantVersion: "v1.0.0",
			wantType:    "git",
		},
		{
			source:      "file:///path/to/repo.git?version=v2.0.0&type=git",
			wantURL:     "file:///path/to/repo.git",
			wantVersion: "v2.0.0",
			wantType:    "git",
		},
		{
			source:      "https://github.com/org/repo.git",
			wantURL:     "https://github.com/org/repo.git",
			wantVersion: "",
			wantType:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			gotURL, gotVersion, gotType := parseSourceURL(tt.source)
			if gotURL != tt.wantURL {
				t.Errorf("parseSourceURL() URL = %v, want %v", gotURL, tt.wantURL)
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("parseSourceURL() version = %v, want %v", gotVersion, tt.wantVersion)
			}
			if gotType != tt.wantType {
				t.Errorf("parseSourceURL() type = %v, want %v", gotType, tt.wantType)
			}
		})
	}
}

func TestSyncResultsAdd(t *testing.T) {
	results := &SyncResults{}

	results.Add("comp1", SyncActionCloned, "v1.0.0", nil)
	results.Add("comp2", SyncActionUpdated, "v2.0.0", nil)
	results.Add("comp3", SyncActionFailed, "", nil)

	currentResults := results.GetResults()
	if len(currentResults) != 3 {
		t.Errorf("Expected 3 results, got %d", len(currentResults))
	}

	if currentResults[0].Name != "comp1" || currentResults[0].Action != SyncActionCloned {
		t.Errorf("First result incorrect: %+v", currentResults[0])
	}

	if currentResults[1].Name != "comp2" || currentResults[1].Action != SyncActionUpdated {
		t.Errorf("Second result incorrect: %+v", currentResults[1])
	}

	if currentResults[2].Name != "comp3" || currentResults[2].Action != SyncActionFailed {
		t.Errorf("Third result incorrect: %+v", currentResults[2])
	}
}

func TestSyncResultsInProgress(t *testing.T) {
	results := &SyncResults{}

	// Set component as in-progress
	results.SetInProgress("comp1")

	currentResults := results.GetResults()
	if len(currentResults) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(currentResults))
	}

	if !currentResults[0].InProgress {
		t.Errorf("Expected component to be in progress")
	}

	// Update result
	results.UpdateResult("comp1", SyncActionCloned, "v1.0.0", nil)

	currentResults = results.GetResults()
	if currentResults[0].InProgress {
		t.Errorf("Expected component to no longer be in progress")
	}

	if currentResults[0].Action != SyncActionCloned {
		t.Errorf("Expected action to be '%s', got '%s'", SyncActionCloned, currentResults[0].Action)
	}

	if currentResults[0].Version != "v1.0.0" {
		t.Errorf("Expected version 'v1.0.0', got '%s'", currentResults[0].Version)
	}
}

func validateComponents(t *testing.T, comp *components.Component) {
	t.Helper()

	if len(comp.Components) != 2 {
		t.Fatalf("Expected 2 child components, got %d", len(comp.Components))
	}

	// Test first component
	core := comp.Components[0]
	if core.Name != testComponentCoreName {
		t.Errorf("Expected component name '%s', got '%s'", testComponentCoreName, core.Name)
	}
	if core.Source != testComponentCoreSource {
		t.Errorf("Expected correct source URL, got '%s'", core.Source)
	}
	if core.Description != testComponentCoreDescription {
		t.Errorf("Expected correct description, got '%s'", core.Description)
	}
	if len(core.Components) != 1 {
		t.Errorf("Expected 1 child, got %d", len(core.Components))
	}

	// Test first component's child
	if len(core.Components) > 0 {
		utils := core.Components[0]
		if utils.Name != testComponentUtilsName {
			t.Errorf("Expected child name '%s', got '%s'", testComponentUtilsName, utils.Name)
		}
		if utils.Source != testComponentUtilsSource {
			t.Errorf("Expected correct child source URL, got '%s'", utils.Source)
		}
	}

	// Test second component
	plugin := comp.Components[1]
	if plugin.Name != testComponentPluginName {
		t.Errorf("Expected component name '%s', got '%s'", testComponentPluginName, plugin.Name)
	}
	if len(plugin.Components) != 0 {
		t.Errorf("Expected 0 children for plugin, got %d", len(plugin.Components))
	}
}

func TestParseComponentsFromJSON(t *testing.T) {
	comp, err := components.ParseComponentFile("../testdata/monhang.json")
	if err != nil {
		t.Fatalf("Failed to parse JSON config: %v", err)
	}

	validateComponents(t, comp)
}

func TestParseComponentsFromTOML(t *testing.T) {
	comp, err := components.ParseComponentFile("../testdata/monhang.toml")
	if err != nil {
		t.Fatalf("Failed to parse TOML config: %v", err)
	}

	validateComponents(t, comp)
}

func TestEmptyComponentsList(t *testing.T) {
	// Create a component with no children
	comp := &components.Component{
		Name:       "test-app",
		Components: []*components.Component{},
	}

	// Verify component name is set
	if comp.Name != "test-app" {
		t.Errorf("Expected component name 'test-app', got '%s'", comp.Name)
	}

	if len(comp.Components) != 0 {
		t.Errorf("Expected 0 child components, got %d", len(comp.Components))
	}
}
