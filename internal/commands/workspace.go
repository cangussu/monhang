// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

package commands

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cangussu/monhang/internal/components"
	"github.com/cangussu/monhang/internal/logging"
	"github.com/charmbracelet/lipgloss"
)

// CmdWorkspace is the workspace command for managing workspace components.
var CmdWorkspace = &Command{
	Name:  "workspace",
	Args:  "<subcommand> [args...]",
	Short: "manage workspace components",
	Long: `
Workspace manages components defined in the configuration file.

Subcommands:
	sync    synchronize workspace components from manifest

Examples:
	monhang workspace sync
	monhang ws sync

Options:
	-f <file>    configuration file (default: ./monhang.json)
`,
}

// CmdWs is the alias for CmdWorkspace.
var CmdWs = &Command{
	Name:  "ws",
	Args:  "<subcommand> [args...]",
	Short: "alias for workspace",
	Long: `
Workspace (ws) manages components defined in the configuration file.

Subcommands:
	sync    synchronize workspace components from manifest

Examples:
	monhang ws sync
	monhang workspace sync

Options:
	-f <file>    configuration file (default: ./monhang.json)
`,
}

var (
	workspaceF = CmdWorkspace.Flag.String("f", "./monhang.json", "configuration file")
	wsF        = CmdWs.Flag.String("f", "./monhang.json", "configuration file")
)

// Sync action constants.
const (
	SyncActionCloned  = "cloned"
	SyncActionUpdated = "updated"
	SyncActionFailed  = "failed"
)

// SyncResult represents the result of syncing a single component.
type SyncResult struct {
	Error   error
	Name    string
	Action  string // SyncActionCloned, SyncActionUpdated, SyncActionFailed
	Version string
}

// SyncResults collects all sync results for final reporting.
type SyncResults struct {
	Results []SyncResult
}

// Add adds a result to the collection.
func (sr *SyncResults) Add(name, action, version string, err error) {
	sr.Results = append(sr.Results, SyncResult{
		Name:    name,
		Action:  action,
		Version: version,
		Error:   err,
	})
}

// parseSourceURL extracts the repository URL and version from a component source.
// Source format: git://github.com/org/repo.git?version=v1.0.0&type=git
func parseSourceURL(source string) (repoURL, version, schemaType string) {
	// Parse the URL
	u, err := url.Parse(source)
	if err != nil {
		// If parsing fails, return the source as-is
		return source, "", ""
	}

	// Extract query parameters
	q := u.Query()
	version = q.Get("version")
	schemaType = q.Get("type")

	// Remove query string to get clean repo URL
	u.RawQuery = ""

	// Convert git:// to https:// for cloning if needed
	repoURL = u.String()
	if strings.HasPrefix(repoURL, "git://") {
		repoURL = "https://" + strings.TrimPrefix(repoURL, "git://")
	}
	// Handle file:// URLs (for local testing)
	if strings.HasPrefix(repoURL, "file://") {
		repoURL = u.String()
	}

	return repoURL, version, schemaType
}

// executeGitCmd runs a git command and returns output and error.
func executeGitCmd(dir string, args ...string) (string, error) {
	logging.GetLogger("workspace").Debug().
		Str("dir", dir).
		Strs("args", args).
		Msg("Executing git command")

	// #nosec G204 -- git commands are constructed by the application
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := strings.TrimSpace(stdout.String())
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg != "" {
			output = errMsg
		}
	}

	return output, err
}

// componentExists checks if a component directory exists and is a git repo.
func componentExists(name string) bool {
	path := filepath.Join(".", name)
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// cloneComponent clones a component from source.
func cloneComponent(name, repoURL, version string) error {
	logging.GetLogger("workspace").Info().
		Str("name", name).
		Str("url", repoURL).
		Str("version", version).
		Msg("Cloning component")

	args := []string{"clone", repoURL, name}
	_, err := executeGitCmd("", args...)
	if err != nil {
		return fmt.Errorf("clone failed: %w", err)
	}

	// If version specified, checkout that version
	if version != "" {
		_, err = executeGitCmd(name, "checkout", version)
		if err != nil {
			return fmt.Errorf("checkout %s failed: %w", version, err)
		}
	}

	return nil
}

// updateComponent fetches and checks out the specified version.
func updateComponent(name, version string) (string, error) {
	logging.GetLogger("workspace").Info().
		Str("name", name).
		Str("version", version).
		Msg("Updating component")

	// Fetch latest
	_, err := executeGitCmd(name, "fetch", "--all", "--tags")
	if err != nil {
		return "", fmt.Errorf("fetch failed: %w", err)
	}

	// If version specified, checkout that version
	if version != "" {
		_, err = executeGitCmd(name, "checkout", version)
		if err != nil {
			return "", fmt.Errorf("checkout %s failed: %w", version, err)
		}
		return SyncActionUpdated, nil
	}

	// If no version, just pull current branch
	_, err = executeGitCmd(name, "pull")
	if err != nil {
		return "", fmt.Errorf("pull failed: %w", err)
	}

	return SyncActionUpdated, nil
}

// getCurrentVersion gets the current git ref for a component.
func getCurrentVersion(name string) string {
	// Try to get tag first
	tag, err := executeGitCmd(name, "describe", "--tags", "--exact-match", "HEAD")
	if err == nil && tag != "" {
		return tag
	}

	// Fall back to branch name
	branch, err := executeGitCmd(name, "rev-parse", "--abbrev-ref", "HEAD")
	if err == nil && branch != "" {
		return branch
	}

	// Fall back to short commit hash
	hash, _ := executeGitCmd(name, "rev-parse", "--short", "HEAD")
	return hash
}

// handleWorkspaceSync processes the sync subcommand.
func handleWorkspaceSync(filename string) {
	logging.GetLogger("workspace").Info().Str("file", filename).Msg("Starting workspace sync")

	proj, err := components.ParseProjectFile(filename)
	Check(err)

	if len(proj.Components) == 0 {
		logging.GetLogger("workspace").Info().Msg("No components defined in manifest")
		fmt.Println("No components defined in manifest")
		return
	}

	fmt.Printf("Syncing %d component(s) from %s\n\n", len(proj.Components), filename)

	// Collect results
	results := &SyncResults{}

	// Process each component in the tree
	for _, comp := range proj.Components {
		syncComponent(comp, 0, results)
	}

	// Print final report
	printSyncReport(results)

	logging.GetLogger("workspace").Info().Msg("Workspace sync completed")
}

// syncComponent synchronizes a component and its children.
func syncComponent(comp components.Component, depth int, results *SyncResults) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	logging.GetLogger("workspace").Debug().
		Str("name", comp.Name).
		Str("source", comp.Source).
		Int("depth", depth).
		Msg("Processing component")

	fmt.Printf("%s- %s\n", indent, comp.Name)

	// Parse source URL
	repoURL, version, _ := parseSourceURL(comp.Source)

	if componentExists(comp.Name) {
		// Component exists - try to update
		fmt.Printf("%s  Updating...\n", indent)
		action, err := updateComponent(comp.Name, version)
		if err != nil {
			fmt.Printf("%s  Error: %v\n", indent, err)
			results.Add(comp.Name, SyncActionFailed, "", err)
		} else {
			currentVer := getCurrentVersion(comp.Name)
			fmt.Printf("%s  Version: %s\n", indent, currentVer)
			results.Add(comp.Name, action, currentVer, nil)
		}
	} else {
		// Component missing - clone it
		fmt.Printf("%s  Cloning from %s\n", indent, repoURL)
		err := cloneComponent(comp.Name, repoURL, version)
		if err != nil {
			fmt.Printf("%s  Error: %v\n", indent, err)
			results.Add(comp.Name, SyncActionFailed, "", err)
		} else {
			currentVer := getCurrentVersion(comp.Name)
			fmt.Printf("%s  Version: %s\n", indent, currentVer)
			results.Add(comp.Name, SyncActionCloned, currentVer, nil)
		}
	}

	// Process children recursively
	for _, child := range comp.Children {
		syncComponent(child, depth+1, results)
	}
}

// printSyncReport prints a summary of all sync operations.
func printSyncReport(results *SyncResults) {
	styles := getSyncStyles()

	fmt.Println()
	fmt.Println(styles.title.Render(" Sync Summary "))
	fmt.Println()

	// Count results
	var cloned, updated, failed int
	for _, r := range results.Results {
		switch r.Action {
		case SyncActionCloned:
			cloned++
		case SyncActionUpdated:
			updated++
		case SyncActionFailed:
			failed++
		}
	}

	// Print table header
	headerLine := fmt.Sprintf("%-30s %-15s %-20s", "Component", "Action", "Version")
	fmt.Println(styles.header.Render(headerLine))
	fmt.Println(strings.Repeat("-", 70))

	// Print each result
	for _, r := range results.Results {
		var actionStyle lipgloss.Style
		action := r.Action
		version := r.Version

		switch r.Action {
		case SyncActionCloned:
			actionStyle = styles.success
		case SyncActionUpdated:
			actionStyle = styles.success
		case SyncActionFailed:
			actionStyle = styles.errorStyle
			if r.Error != nil {
				version = r.Error.Error()
				if len(version) > 20 {
					version = version[:17] + "..."
				}
			}
		default:
			actionStyle = styles.normal
		}

		line := fmt.Sprintf("%-30s %-15s %-20s", r.Name, action, version)
		fmt.Println(actionStyle.Render(line))
	}

	// Print summary
	fmt.Println()
	fmt.Printf("Summary: %s cloned, %s updated, %s failed\n",
		styles.success.Render(fmt.Sprintf("%d", cloned)),
		styles.success.Render(fmt.Sprintf("%d", updated)),
		styles.errorStyle.Render(fmt.Sprintf("%d", failed)))

	if failed > 0 {
		fmt.Println()
		fmt.Println(styles.errorStyle.Render("Some components failed to sync. See errors above."))
	}
}

// syncStyles holds the lipgloss styles for sync output.
type syncStyles struct {
	title      lipgloss.Style
	header     lipgloss.Style
	success    lipgloss.Style
	errorStyle lipgloss.Style
	normal     lipgloss.Style
}

// getSyncStyles returns the configured sync UI styles.
func getSyncStyles() syncStyles {
	return syncStyles{
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1),
		header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Underline(true),
		success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")),
		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")),
		normal: lipgloss.NewStyle(),
	}
}

func runWorkspace(_ *Command, args []string) {
	if len(args) == 0 {
		fmt.Println("Error: No subcommand specified")
		fmt.Println("Usage: monhang workspace <subcommand> [args]")
		fmt.Println("\nSubcommands: sync")
		os.Exit(1)
	}

	subcommand := args[0]

	logging.GetLogger("workspace").Info().
		Str("subcommand", subcommand).
		Msg("Starting workspace command")

	// Execute based on subcommand
	switch subcommand {
	case "sync":
		handleWorkspaceSync(*workspaceF)
	default:
		logging.GetLogger("workspace").Error().Str("subcommand", subcommand).Msg("Unknown workspace subcommand")
		fmt.Printf("Error: Unknown subcommand '%s'\n", subcommand)
		fmt.Println("Available subcommands: sync")
		os.Exit(1)
	}

	logging.GetLogger("workspace").Info().Str("subcommand", subcommand).Msg("Workspace command completed")
}

func runWs(_ *Command, args []string) {
	if len(args) == 0 {
		fmt.Println("Error: No subcommand specified")
		fmt.Println("Usage: monhang ws <subcommand> [args]")
		fmt.Println("\nSubcommands: sync")
		os.Exit(1)
	}

	subcommand := args[0]

	logging.GetLogger("workspace").Info().
		Str("subcommand", subcommand).
		Str("alias", "ws").
		Msg("Starting workspace command (via ws alias)")

	// Execute based on subcommand
	switch subcommand {
	case "sync":
		handleWorkspaceSync(*wsF)
	default:
		logging.GetLogger("workspace").Error().Str("subcommand", subcommand).Msg("Unknown workspace subcommand")
		fmt.Printf("Error: Unknown subcommand '%s'\n", subcommand)
		fmt.Println("Available subcommands: sync")
		os.Exit(1)
	}

	logging.GetLogger("workspace").Info().Str("subcommand", subcommand).Msg("Workspace command completed (via ws alias)")
}

func init() {
	CmdWorkspace.Run = runWorkspace
	CmdWs.Run = runWs
}
