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
	"sync"
	"time"

	"github.com/cangussu/monhang/internal/components"
	"github.com/cangussu/monhang/internal/logging"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
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
	Error      error
	Name       string
	Action     string // SyncActionCloned, SyncActionUpdated, SyncActionFailed
	Version    string
	InProgress bool
}

// SyncResults collects all sync results for final reporting.
type SyncResults struct {
	Results []SyncResult
	mu      sync.RWMutex
}

// Add adds a result to the collection.
func (sr *SyncResults) Add(name, action, version string, err error) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.Results = append(sr.Results, SyncResult{
		Name:       name,
		Action:     action,
		Version:    version,
		Error:      err,
		InProgress: false,
	})
}

// SetInProgress marks a component as in-progress.
func (sr *SyncResults) SetInProgress(name string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.Results = append(sr.Results, SyncResult{
		Name:       name,
		InProgress: true,
	})
}

// UpdateResult updates an in-progress result.
func (sr *SyncResults) UpdateResult(name, action, version string, err error) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	for i := range sr.Results {
		if sr.Results[i].Name == name && sr.Results[i].InProgress {
			sr.Results[i].Action = action
			sr.Results[i].Version = version
			sr.Results[i].Error = err
			sr.Results[i].InProgress = false
			return
		}
	}
}

// GetResults returns a copy of all results.
func (sr *SyncResults) GetResults() []SyncResult {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	resultsCopy := make([]SyncResult, len(sr.Results))
	copy(resultsCopy, sr.Results)
	return resultsCopy
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

	comp, err := components.ParseComponentFile(filename)
	Check(err)

	if len(comp.Components) == 0 {
		logging.GetLogger("workspace").Info().Msg("No components defined in manifest")
		fmt.Println("No components defined in manifest")
		return
	}

	// Collect results
	results := &SyncResults{}

	// Collect all components (flatten tree)
	allComponents := flattenComponents(comp.Components)

	// Run interactive sync
	if err := runInteractiveSync(filename, allComponents, results); err != nil {
		fmt.Printf("Error running interactive sync: %v\n", err)
		os.Exit(1)
	}

	logging.GetLogger("workspace").Info().Msg("Workspace sync completed")
}

// flattenComponents flattens the component tree into a list.
func flattenComponents(comps []*components.Component) []*components.Component {
	// Pre-allocate with estimated capacity
	result := make([]*components.Component, 0, len(comps)*2)
	for _, comp := range comps {
		result = append(result, comp)
		if len(comp.Components) > 0 {
			result = append(result, flattenComponents(comp.Components)...)
		}
	}
	return result
}

// syncModel is the bubbletea model for sync operations.
//
//nolint:govet // fieldalignment: UI model where field order doesn't significantly impact performance
type syncModel struct {
	components []*components.Component
	results    *SyncResults
	filename   string
	allDone    bool
	quitting   bool
	width      int
	height     int
}

type syncTickMsg time.Time

func syncTickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return syncTickMsg(t)
	})
}

// Init initializes the bubbletea model.
func (m syncModel) Init() tea.Cmd {
	return syncTickCmd()
}

// Update handles bubbletea messages and updates the model.
func (m syncModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case KeyCtrlC, "q":
			m.quitting = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case syncTickMsg:
		// Check if all done
		currentResults := m.results.GetResults()
		allDone := true
		for _, r := range currentResults {
			if r.InProgress {
				allDone = false
				break
			}
		}
		m.allDone = allDone

		if m.allDone {
			return m, tea.Quit
		}

		return m, syncTickCmd()
	}

	return m, nil
}

// View renders the bubbletea UI.
//
//nolint:gocyclo // UI rendering function with multiple output states
func (m syncModel) View() string {
	if m.quitting && !m.allDone {
		return ""
	}

	styles := getSyncStyles()
	var s strings.Builder

	s.WriteString(styles.title.Render(fmt.Sprintf(" Syncing %d component(s) from %s ", len(m.components), m.filename)))
	s.WriteString("\n\n")

	// Count results
	currentResults := m.results.GetResults()
	var cloned, updated, failed, inProgress int
	for _, r := range currentResults {
		if r.InProgress {
			inProgress++
			continue
		}
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
	headerLine := fmt.Sprintf("%-30s %-15s %-20s", "Component", "Status", "Version")
	s.WriteString(styles.header.Render(headerLine) + "\n")
	s.WriteString(strings.Repeat("-", 70) + "\n")

	// Print each result
	for _, r := range currentResults {
		var statusStyle lipgloss.Style
		status := r.Action
		version := r.Version

		if r.InProgress {
			status = "syncing..."
			statusStyle = styles.running
			version = ""
		} else {
			switch r.Action {
			case SyncActionCloned:
				statusStyle = styles.success
			case SyncActionUpdated:
				statusStyle = styles.success
			case SyncActionFailed:
				statusStyle = styles.errorStyle
				if r.Error != nil {
					version = r.Error.Error()
					if len(version) > 20 {
						version = version[:17] + "..."
					}
				}
			default:
				statusStyle = styles.normal
			}
		}

		line := fmt.Sprintf("%-30s %-15s %-20s", r.Name, status, version)
		s.WriteString(statusStyle.Render(line) + "\n")
	}

	// Print summary
	s.WriteString("\n")
	if m.allDone {
		s.WriteString(styles.success.Render("✓ Sync completed!") + "\n\n")
		s.WriteString(fmt.Sprintf("Summary: %s cloned, %s updated, %s failed\n",
			styles.success.Render(fmt.Sprintf("%d", cloned)),
			styles.success.Render(fmt.Sprintf("%d", updated)),
			styles.errorStyle.Render(fmt.Sprintf("%d", failed))))

		if failed > 0 {
			s.WriteString("\n")
			s.WriteString(styles.errorStyle.Render("Some components failed to sync. See errors above."))
		}
	} else {
		s.WriteString(fmt.Sprintf("Progress: %d syncing, %d cloned, %d updated, %d failed\n",
			inProgress, cloned, updated, failed))
		s.WriteString("\nPress q to quit")
	}

	return s.String()
}

// runInteractiveSync runs sync operation in interactive mode.
func runInteractiveSync(filename string, comps []*components.Component, results *SyncResults) error {
	// Check if we have a TTY
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		// Fall back to non-interactive mode
		runNonInteractiveSync(filename, comps, results)
		return nil
	}

	// Redirect logging to a file during interactive mode to prevent output interference
	// but still capture logs for debugging
	logFile, previousLogger, err := logging.RedirectLoggingToFile()
	if err != nil {
		logging.GetLogger("workspace").Warn().Err(err).Msg("Failed to redirect logs, falling back to non-interactive")
		runNonInteractiveSync(filename, comps, results)
		return nil
	}
	defer func() {
		logPath := logging.RestoreLogger(logFile, previousLogger)
		if logging.IsDebugEnabled() {
			logging.GetLogger("workspace").Debug().Str("log_file", logPath).Msg("Interactive session logs saved")
		}
	}()

	// Start sync in background
	go func() {
		for _, comp := range comps {
			syncComponentBackground(comp, results)
		}
	}()

	// Start bubbletea UI
	// Note: We don't use WithAltScreen() because it can cause flickering
	// Instead, we suppress logging and let bubbletea manage the display
	p := tea.NewProgram(
		syncModel{
			components: comps,
			results:    results,
			filename:   filename,
		},
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	if _, err := p.Run(); err != nil {
		// Restore logging before falling back
		logPath := logging.RestoreLogger(logFile, previousLogger)
		// If interactive mode fails, fall back to non-interactive
		logging.GetLogger("workspace").Warn().Err(err).Str("log_file", logPath).Msg("Interactive mode failed, falling back to non-interactive")
		runNonInteractiveSync(filename, comps, results)
		return nil
	}

	return nil
}

// runNonInteractiveSync runs sync in non-interactive mode (for CI/non-TTY environments).
func runNonInteractiveSync(filename string, comps []*components.Component, results *SyncResults) {
	styles := getSyncStyles()

	fmt.Printf("Syncing %d component(s) from %s\n\n", len(comps), filename)

	// Sync all components sequentially
	for _, comp := range comps {
		fmt.Printf("- %s... ", comp.Name)
		results.SetInProgress(comp.Name)

		repoURL, version, _ := parseSourceURL(comp.Source)

		if componentExists(comp.Name) {
			action, err := updateComponent(comp.Name, version)
			if err != nil {
				fmt.Printf("%s\n", styles.errorStyle.Render("failed: "+err.Error()))
				results.UpdateResult(comp.Name, SyncActionFailed, "", err)
			} else {
				currentVer := getCurrentVersion(comp.Name)
				fmt.Printf("%s (%s)\n", styles.success.Render(action), currentVer)
				results.UpdateResult(comp.Name, action, currentVer, nil)
			}
		} else {
			err := cloneComponent(comp.Name, repoURL, version)
			if err != nil {
				fmt.Printf("%s\n", styles.errorStyle.Render("failed: "+err.Error()))
				results.UpdateResult(comp.Name, SyncActionFailed, "", err)
			} else {
				currentVer := getCurrentVersion(comp.Name)
				fmt.Printf("%s (%s)\n", styles.success.Render("cloned"), currentVer)
				results.UpdateResult(comp.Name, SyncActionCloned, currentVer, nil)
			}
		}
	}

	// Print final summary
	printNonInteractiveSummary(results)
}

// printNonInteractiveSummary prints a summary in non-interactive mode.
func printNonInteractiveSummary(results *SyncResults) {
	styles := getSyncStyles()

	currentResults := results.GetResults()
	var cloned, updated, failed int
	for _, r := range currentResults {
		if r.InProgress {
			continue
		}
		switch r.Action {
		case SyncActionCloned:
			cloned++
		case SyncActionUpdated:
			updated++
		case SyncActionFailed:
			failed++
		}
	}

	fmt.Println()
	fmt.Println(styles.title.Render(" Sync Summary "))
	fmt.Println()
	fmt.Printf("Summary: %s cloned, %s updated, %s failed\n",
		styles.success.Render(fmt.Sprintf("%d", cloned)),
		styles.success.Render(fmt.Sprintf("%d", updated)),
		styles.errorStyle.Render(fmt.Sprintf("%d", failed)))

	if failed > 0 {
		fmt.Println()
		fmt.Println(styles.errorStyle.Render("Some components failed to sync."))
	}
}

// syncComponentBackground synchronizes a component in the background.
func syncComponentBackground(comp *components.Component, results *SyncResults) {
	logging.GetLogger("workspace").Debug().
		Str("name", comp.Name).
		Str("source", comp.Source).
		Msg("Processing component in background")

	// Mark as in-progress
	results.SetInProgress(comp.Name)

	// Parse source URL
	repoURL, version, _ := parseSourceURL(comp.Source)

	if componentExists(comp.Name) {
		// Component exists - try to update
		action, err := updateComponent(comp.Name, version)
		if err != nil {
			results.UpdateResult(comp.Name, SyncActionFailed, "", err)
		} else {
			currentVer := getCurrentVersion(comp.Name)
			results.UpdateResult(comp.Name, action, currentVer, nil)
		}
	} else {
		// Component missing - clone it
		err := cloneComponent(comp.Name, repoURL, version)
		if err != nil {
			results.UpdateResult(comp.Name, SyncActionFailed, "", err)
		} else {
			currentVer := getCurrentVersion(comp.Name)
			results.UpdateResult(comp.Name, SyncActionCloned, currentVer, nil)
		}
	}
}

// syncStyles holds the lipgloss styles for sync output.
type syncStyles struct {
	title      lipgloss.Style
	header     lipgloss.Style
	success    lipgloss.Style
	errorStyle lipgloss.Style
	running    lipgloss.Style
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
		running: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00")),
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
