// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

package monhang

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// CmdWorkspace is the workspace command for managing workspace components.
var CmdWorkspace = &Command{
	Name:  "workspace",
	Args:  "sync [options]",
	Short: "manage workspace components",
	Long: `
Workspace manages components defined in the project configuration file.

Subcommands:
  sync    synchronize workspace components from the manifest

Usage:
  monhang workspace sync [-f configfile]
  monhang ws sync [-f configfile]
`,
}

// CmdWs is an alias for CmdWorkspace.
var CmdWs = &Command{
	Name:  "ws",
	Args:  "sync [options]",
	Short: "alias for workspace command",
	Long:  CmdWorkspace.Long,
}

var workspaceF = CmdWorkspace.Flag.String("f", "<defaultconfig>", "configuration file")
var wsF = CmdWs.Flag.String("f", "<defaultconfig>", "configuration file")

func getWorkspaceFilename(configFlag *string) string {
	if *configFlag != "<defaultconfig>" {
		return *configFlag
	}
	// Try TOML first, then fall back to JSON
	if _, err := os.Stat("./monhang.toml"); err == nil {
		return "./monhang.toml"
	}
	return "./monhang.json"
}

// parseComponentSource parses a component source URL to extract type, location, and version.
// Expected formats:
//   - git+https://github.com/user/repo@v1.0.0
//   - https://example.com/component.tar.gz@1.2.3
//
// Returns: (type, location, version, error)
func parseComponentSource(source string) (string, string, string, error) {
	var sourceType, location, version string

	// Split by @ to get version
	parts := strings.Split(source, "@")
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("invalid source format: missing version (expected format: <url>@<version>)")
	}

	location = parts[0]
	version = parts[1]

	// Check for explicit type prefix (e.g., git+https://)
	if strings.Contains(location, "+") {
		typeParts := strings.SplitN(location, "+", 2)
		sourceType = typeParts[0]
		location = typeParts[1]
	} else {
		// Infer type from URL
		parsedURL, err := url.Parse(location)
		if err != nil {
			return "", "", "", fmt.Errorf("invalid URL in source: %w", err)
		}

		// Infer type based on URL scheme and extension
		if parsedURL.Scheme == "http" || parsedURL.Scheme == "https" {
			if strings.HasSuffix(parsedURL.Path, ".git") || strings.Contains(parsedURL.Host, "github") {
				sourceType = "git"
			} else {
				sourceType = "http"
			}
		} else if parsedURL.Scheme == "git" {
			sourceType = "git"
		} else {
			sourceType = "unknown"
		}
	}

	return sourceType, location, version, nil
}

// syncComponent fetches and sets up a single component.
func syncComponent(comp Component, workspaceDir string) error {
	mglog.Noticef("Syncing component: %s", comp.Name)
	mglog.Debugf("  Description: %s", comp.Description)
	mglog.Debugf("  Source: %s", comp.Source)

	// Parse the component source
	sourceType, location, version, err := parseComponentSource(comp.Source)
	if err != nil {
		return fmt.Errorf("failed to parse source for component %s: %w", comp.Name, err)
	}

	mglog.Debugf("  Type: %s, Location: %s, Version: %s", sourceType, location, version)

	// Create component directory
	componentPath := filepath.Join(workspaceDir, comp.Name)

	// Check if component already exists
	if _, err := os.Stat(componentPath); err == nil {
		mglog.Noticef("Component %s already exists at %s", comp.Name, componentPath)
		return nil
	}

	// Handle different source types
	switch sourceType {
	case "git":
		return syncGitComponent(location, version, componentPath)
	case "http", "https":
		mglog.Warningf("HTTP/HTTPS sources not yet implemented for component %s", comp.Name)
		return fmt.Errorf("HTTP/HTTPS sources not yet implemented")
	default:
		return fmt.Errorf("unsupported source type: %s", sourceType)
	}
}

// syncGitComponent clones a git repository at a specific version.
func syncGitComponent(repoURL, version, destPath string) error {
	mglog.Noticef("Cloning git repository: %s@%s", repoURL, version)

	// Clone the repository
	args := []string{"clone", "--branch", version, repoURL, destPath}
	git(args)

	return nil
}

// runWorkspaceSync executes the workspace sync command.
func runWorkspaceSync(configFlag *string) error {
	filename := getWorkspaceFilename(configFlag)

	// Parse the project file
	proj, err := ParseProjectFile(filename)
	if err != nil {
		return fmt.Errorf("failed to parse project file: %w", err)
	}

	if len(proj.Components) == 0 {
		mglog.Notice("No components defined in the project configuration")
		return nil
	}

	mglog.Noticef("Found %d component(s) to sync", len(proj.Components))

	// Create workspace directory (default: current directory)
	workspaceDir := "."

	// Sync each component
	for _, comp := range proj.Components {
		if err := syncComponent(comp, workspaceDir); err != nil {
			mglog.Errorf("Failed to sync component %s: %v", comp.Name, err)
			return err
		}
	}

	mglog.Notice("Workspace sync completed successfully")
	return nil
}

// runWorkspace handles the workspace command and its subcommands.
func runWorkspace(cmd *Command, args []string) {
	// Check if we have any arguments at all
	// Note: flags have already been parsed by main.go, so args only contains non-flag arguments
	// If the user wants to use flags, they should come before the subcommand:
	// monhang workspace -f config.toml sync
	// or we check if workspaceF was set

	if len(args) < 1 {
		// No subcommand provided, check if there are any args in the original command
		fmt.Println("Error: workspace command requires a subcommand")
		fmt.Println(cmd.Long)
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "sync":
		if err := runWorkspaceSync(workspaceF); err != nil {
			Check(err)
		}
	default:
		fmt.Printf("Error: unknown subcommand '%s'\n", subcommand)
		fmt.Println(cmd.Long)
		os.Exit(1)
	}
}

// runWs handles the ws (workspace alias) command.
func runWs(cmd *Command, args []string) {
	if len(args) < 1 {
		fmt.Println("Error: ws command requires a subcommand")
		fmt.Println(cmd.Long)
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "sync":
		if err := runWorkspaceSync(wsF); err != nil {
			Check(err)
		}
	default:
		fmt.Printf("Error: unknown subcommand '%s'\n", subcommand)
		fmt.Println(cmd.Long)
		os.Exit(1)
	}
}

func init() {
	CmdWorkspace.Run = runWorkspace // break init loop
	CmdWs.Run = runWs               // break init loop
}
