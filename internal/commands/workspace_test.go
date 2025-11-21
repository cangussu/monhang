// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

package commands

import (
	"bytes"
	"io"
	"os"
	"strings"
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

func TestSyncComponentTree(t *testing.T) {
	comp := components.Component{
		Name:        testComponentCoreName,
		Source:      testComponentCoreSource,
		Description: testComponentCoreDescription,
		Children: []components.Component{
			{
				Name:        testComponentUtilsName,
				Source:      testComponentUtilsSource,
				Description: testComponentUtilsDesc,
			},
		},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	results := &SyncResults{}
	syncComponent(comp, 0, results)

	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close pipe writer: %v", err)
	}
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("Failed to copy pipe output: %v", err)
	}
	output := buf.String()

	// Verify output contains component name
	if !strings.Contains(output, testComponentCoreName) {
		t.Errorf("Expected output to contain '%s', got: %s", testComponentCoreName, output)
	}

	// Verify output contains child component
	if !strings.Contains(output, testComponentUtilsName) {
		t.Errorf("Expected output to contain child '%s', got: %s", testComponentUtilsName, output)
	}

	// Verify results were collected (should have 2 results - parent and child)
	if len(results.Results) != 2 {
		t.Errorf("Expected 2 sync results, got %d", len(results.Results))
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

	if len(results.Results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results.Results))
	}

	if results.Results[0].Name != "comp1" || results.Results[0].Action != SyncActionCloned {
		t.Errorf("First result incorrect: %+v", results.Results[0])
	}

	if results.Results[1].Name != "comp2" || results.Results[1].Action != SyncActionUpdated {
		t.Errorf("Second result incorrect: %+v", results.Results[1])
	}

	if results.Results[2].Name != "comp3" || results.Results[2].Action != SyncActionFailed {
		t.Errorf("Third result incorrect: %+v", results.Results[2])
	}
}

func validateComponents(t *testing.T, proj *components.Project) {
	t.Helper()

	if len(proj.Components) != 2 {
		t.Fatalf("Expected 2 components, got %d", len(proj.Components))
	}

	// Test first component
	core := proj.Components[0]
	if core.Name != testComponentCoreName {
		t.Errorf("Expected component name '%s', got '%s'", testComponentCoreName, core.Name)
	}
	if core.Source != testComponentCoreSource {
		t.Errorf("Expected correct source URL, got '%s'", core.Source)
	}
	if core.Description != testComponentCoreDescription {
		t.Errorf("Expected correct description, got '%s'", core.Description)
	}
	if len(core.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(core.Children))
	}

	// Test first component's child
	if len(core.Children) > 0 {
		utils := core.Children[0]
		if utils.Name != testComponentUtilsName {
			t.Errorf("Expected child name '%s', got '%s'", testComponentUtilsName, utils.Name)
		}
		if utils.Source != testComponentUtilsSource {
			t.Errorf("Expected correct child source URL, got '%s'", utils.Source)
		}
	}

	// Test second component
	plugin := proj.Components[1]
	if plugin.Name != testComponentPluginName {
		t.Errorf("Expected component name '%s', got '%s'", testComponentPluginName, plugin.Name)
	}
	if len(plugin.Children) != 0 {
		t.Errorf("Expected 0 children for plugin, got %d", len(plugin.Children))
	}
}

func TestParseComponentsFromJSON(t *testing.T) {
	proj, err := components.ParseProjectFile("../testdata/monhang.json")
	if err != nil {
		t.Fatalf("Failed to parse JSON config: %v", err)
	}

	validateComponents(t, proj)
}

func TestParseComponentsFromTOML(t *testing.T) {
	proj, err := components.ParseProjectFile("../testdata/monhang.toml")
	if err != nil {
		t.Fatalf("Failed to parse TOML config: %v", err)
	}

	validateComponents(t, proj)
}

func TestEmptyComponentsList(t *testing.T) {
	// Create a project with no components
	proj := &components.Project{
		ComponentRef: components.ComponentRef{
			Name: "test-app",
		},
		Components: []components.Component{},
	}

	if len(proj.Components) != 0 {
		t.Errorf("Expected 0 components, got %d", len(proj.Components))
	}
}
