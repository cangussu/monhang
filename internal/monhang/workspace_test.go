// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

package monhang

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
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
	comp := Component{
		Name:        testComponentCoreName,
		Source:      testComponentCoreSource,
		Description: testComponentCoreDescription,
		Children: []Component{
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

	syncComponent(comp, 0)

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

	// Verify output contains component description
	if !strings.Contains(output, testComponentCoreDescription) {
		t.Errorf("Expected output to contain '%s', got: %s", testComponentCoreDescription, output)
	}

	// Verify output contains component source
	if !strings.Contains(output, testComponentCoreSource) {
		t.Errorf("Expected output to contain source URL, got: %s", output)
	}

	// Verify output contains child component
	if !strings.Contains(output, testComponentUtilsName) {
		t.Errorf("Expected output to contain child '%s', got: %s", testComponentUtilsName, output)
	}

	// Verify output contains child description
	if !strings.Contains(output, testComponentUtilsDesc) {
		t.Errorf("Expected output to contain '%s', got: %s", testComponentUtilsDesc, output)
	}
}

func validateComponents(t *testing.T, proj *Project) {
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
	proj, err := ParseProjectFile("testdata/monhang.json")
	if err != nil {
		t.Fatalf("Failed to parse JSON config: %v", err)
	}

	validateComponents(t, proj)
}

func TestParseComponentsFromTOML(t *testing.T) {
	proj, err := ParseProjectFile("testdata/monhang.toml")
	if err != nil {
		t.Fatalf("Failed to parse TOML config: %v", err)
	}

	validateComponents(t, proj)
}

func TestEmptyComponentsList(t *testing.T) {
	// Create a project with no components
	proj := &Project{
		ComponentRef: ComponentRef{
			Name: "test-app",
		},
		Components: []Component{},
	}

	if len(proj.Components) != 0 {
		t.Errorf("Expected 0 components, got %d", len(proj.Components))
	}
}
