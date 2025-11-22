// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

package commands

import (
	"bytes"
	"context"
	"fmt"
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
)

// CmdGit is the git command for running git operations across all repos.
var CmdGit = &Command{
	Name:  "git",
	Args:  "<subcommand> [args...]",
	Short: "run git operations across all repos",
	Long: `
Git runs git operations across all repositories defined in the configuration file.

Subcommands:
	status              show status of all repos
	pull                pull updates for all repos
	fetch               fetch updates for all repos
	checkout <branch>   checkout a branch in all repos
	branch <branch>     create and checkout a new branch in all repos

Examples:
	monhang git status
	monhang git pull
	monhang git checkout main
	monhang git branch feat/new-feature

Options:
	-f <file>    configuration file (default: ./monhang.json)
	-i           interactive mode with bubbletea UI
	-p           run operations in parallel
`,
}

var (
	gitF           = CmdGit.Flag.String("f", "./monhang.json", "configuration file")
	gitInteractive = CmdGit.Flag.Bool("i", false, "interactive mode")
	gitParallel    = CmdGit.Flag.Bool("p", false, "run in parallel")
)

// GitResult holds the result of a git operation in a repo.
type GitResult struct {
	StartTime  time.Time
	EndTime    time.Time
	Error      error
	Name       string
	Path       string
	Operation  string
	Branch     string
	CommitHash string
	Status     string
	Output     string
	Running    bool
}

// GitExecutor manages git operations across repositories.
type GitExecutor struct {
	results map[string]*GitResult
	mu      sync.RWMutex
}

// NewGitExecutor creates a new GitExecutor.
func NewGitExecutor() *GitExecutor {
	return &GitExecutor{
		results: make(map[string]*GitResult),
	}
}

// executeGitCommand runs a git command in a repo directory and captures output.
func executeGitCommand(ctx context.Context, dir string, args ...string) (string, error) {
	logging.GetLogger("git").Debug().
		Str("dir", dir).
		Strs("args", args).
		Msg("Executing git command")

	// #nosec G204 -- git commands are constructed by the application
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	output := strings.TrimSpace(out.String())

	if err != nil {
		logging.GetLogger("git").Error().
			Err(err).
			Str("dir", dir).
			Strs("args", args).
			Str("output", output).
			Msg("Git command failed")
	}

	return output, err
}

// getRepoInfo gets current branch and commit hash for a repo.
func getRepoInfo(ctx context.Context, dir string) (branch, commit string) {
	branch, _ = executeGitCommand(ctx, dir, "rev-parse", "--abbrev-ref", "HEAD")
	commit, _ = executeGitCommand(ctx, dir, "rev-parse", "--short", "HEAD")
	return
}

// getRepoStatus gets the git status for a repo.
func getRepoStatus(ctx context.Context, dir string) string {
	status, _ := executeGitCommand(ctx, dir, "status", "--short")
	if status == "" {
		return "clean"
	}
	lines := strings.Split(status, "\n")
	return fmt.Sprintf("%d changes", len(lines))
}

// isGitRepository checks if the given directory is a git repository.
func isGitRepository(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// ExecuteStatus runs git status on a repository.
func (ge *GitExecutor) ExecuteStatus(ctx context.Context, name, path string) {
	logging.GetLogger("git").Debug().Str("repo", name).Str("operation", "status").Msg("Starting git status")

	result := &GitResult{
		Name:      name,
		Path:      path,
		Operation: "status",
		StartTime: time.Now(),
		Running:   true,
	}

	ge.mu.Lock()
	ge.results[name] = result
	ge.mu.Unlock()

	// Get repo info
	branch, commit := getRepoInfo(ctx, path)
	status := getRepoStatus(ctx, path)

	output, err := executeGitCommand(ctx, path, "status", "--short")

	ge.mu.Lock()
	defer ge.mu.Unlock()
	result.Branch = branch
	result.CommitHash = commit
	result.Status = status
	result.Output = output
	result.Error = err
	result.Running = false
	result.EndTime = time.Now()

	duration := result.EndTime.Sub(result.StartTime)
	if err != nil {
		logging.GetLogger("git").Error().Err(err).Str("repo", name).Dur("duration", duration).Msg("Git status failed")
	} else {
		logging.GetLogger("git").Debug().Str("repo", name).Str("branch", branch).Str("status", status).Dur("duration", duration).Msg("Git status completed")
	}
}

// ExecutePull runs git pull on a repository.
func (ge *GitExecutor) ExecutePull(ctx context.Context, name, path string) {
	result := &GitResult{
		Name:      name,
		Path:      path,
		Operation: "pull",
		StartTime: time.Now(),
		Running:   true,
	}

	ge.mu.Lock()
	ge.results[name] = result
	ge.mu.Unlock()

	// Get initial info
	branch, commit := getRepoInfo(ctx, path)

	// Run pull
	output, err := executeGitCommand(ctx, path, "pull")

	// Get updated info
	_, newCommit := getRepoInfo(ctx, path)

	ge.mu.Lock()
	defer ge.mu.Unlock()
	result.Branch = branch
	result.CommitHash = newCommit
	result.Output = output
	result.Error = err
	result.Running = false
	result.EndTime = time.Now()

	if commit != newCommit {
		result.Status = fmt.Sprintf("updated: %s -> %s", commit, newCommit)
	} else {
		result.Status = "up to date"
	}
}

// ExecuteFetch runs git fetch on a repository.
func (ge *GitExecutor) ExecuteFetch(ctx context.Context, name, path string) {
	result := &GitResult{
		Name:      name,
		Path:      path,
		Operation: "fetch",
		StartTime: time.Now(),
		Running:   true,
	}

	ge.mu.Lock()
	ge.results[name] = result
	ge.mu.Unlock()

	branch, commit := getRepoInfo(ctx, path)
	output, err := executeGitCommand(ctx, path, "fetch")

	ge.mu.Lock()
	defer ge.mu.Unlock()
	result.Branch = branch
	result.CommitHash = commit
	result.Output = output
	result.Error = err
	result.Running = false
	result.EndTime = time.Now()
	result.Status = "fetched"
}

// ExecuteCheckout runs git checkout on a repository.
func (ge *GitExecutor) ExecuteCheckout(ctx context.Context, name, path, branch string) {
	result := &GitResult{
		Name:      name,
		Path:      path,
		Operation: fmt.Sprintf("checkout %s", branch),
		StartTime: time.Now(),
		Running:   true,
	}

	ge.mu.Lock()
	ge.results[name] = result
	ge.mu.Unlock()

	output, err := executeGitCommand(ctx, path, "checkout", branch)

	// Get updated info
	newBranch, commit := getRepoInfo(ctx, path)

	ge.mu.Lock()
	defer ge.mu.Unlock()
	result.Branch = newBranch
	result.CommitHash = commit
	result.Output = output
	result.Error = err
	result.Running = false
	result.EndTime = time.Now()
	result.Status = fmt.Sprintf("on %s", newBranch)
}

// ExecuteBranch creates and checks out a new branch in a repository.
func (ge *GitExecutor) ExecuteBranch(ctx context.Context, name, path, branch string) {
	result := &GitResult{
		Name:      name,
		Path:      path,
		Operation: fmt.Sprintf("branch %s", branch),
		StartTime: time.Now(),
		Running:   true,
	}

	ge.mu.Lock()
	ge.results[name] = result
	ge.mu.Unlock()

	output, err := executeGitCommand(ctx, path, "checkout", "-b", branch)

	// Get updated info
	newBranch, commit := getRepoInfo(ctx, path)

	ge.mu.Lock()
	defer ge.mu.Unlock()
	result.Branch = newBranch
	result.CommitHash = commit
	result.Output = output
	result.Error = err
	result.Running = false
	result.EndTime = time.Now()
	result.Status = fmt.Sprintf("created %s", newBranch)
}

// GetResults returns a copy of all results.
func (ge *GitExecutor) GetResults() map[string]*GitResult {
	ge.mu.RLock()
	defer ge.mu.RUnlock()

	results := make(map[string]*GitResult)
	for k, v := range ge.results {
		resultCopy := *v
		results[k] = &resultCopy
	}
	return results
}

// gitModel is the bubbletea model for git operations.
//
//nolint:govet // fieldalignment: UI model where field order doesn't significantly impact performance
type gitModel struct {
	repos       []components.ComponentRef
	results     map[string]*GitResult
	executor    *GitExecutor
	operation   string
	viewingRepo string
	selected    int
	width       int
	height      int
	quitting    bool
	allDone     bool
}

type gitTickMsg time.Time

func gitTickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return gitTickMsg(t)
	})
}

// Init initializes the bubbletea model.
func (m gitModel) Init() tea.Cmd {
	return gitTickCmd()
}

// Update handles bubbletea messages and updates the model.
func (m gitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case KeyCtrlC, "q":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.repos)-1 {
				m.selected++
			}
		case "enter":
			if m.selected < len(m.repos) {
				m.viewingRepo = m.repos[m.selected].Name
			}
		case "esc":
			m.viewingRepo = ""
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case gitTickMsg:
		m.results = m.executor.GetResults()

		allDone := true
		for _, result := range m.results {
			if result.Running {
				allDone = false
				break
			}
		}
		m.allDone = allDone

		if m.allDone {
			return m, tea.Quit
		}

		return m, gitTickCmd()
	}

	return m, nil
}

// gitUIStyles holds the lipgloss styles for git UI.
type gitUIStyles struct {
	title      lipgloss.Style
	header     lipgloss.Style
	selected   lipgloss.Style
	normal     lipgloss.Style
	success    lipgloss.Style
	error      lipgloss.Style
	running    lipgloss.Style
	tableCell  lipgloss.Style
	tableValue lipgloss.Style
}

// getGitUIStyles returns the configured git UI styles.
func getGitUIStyles() gitUIStyles {
	return gitUIStyles{
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1),
		header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Underline(true),
		selected: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")),
		normal: lipgloss.NewStyle(),
		success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")),
		error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")),
		running: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00")),
		tableCell: lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1),
		tableValue: lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			Foreground(lipgloss.Color("#00D7FF")),
	}
}

// renderGitRepoDetail renders detailed view for a specific repository.
func (m gitModel) renderGitRepoDetail(styles gitUIStyles) string {
	result := m.results[m.viewingRepo]
	if result == nil {
		return ""
	}

	var s strings.Builder
	s.WriteString(styles.title.Render("Repository: "+m.viewingRepo) + "\n\n")
	s.WriteString(fmt.Sprintf("Operation: %s\n", m.operation))
	s.WriteString(fmt.Sprintf("Path: %s\n", result.Path))
	s.WriteString(fmt.Sprintf("Branch: %s\n", result.Branch))
	s.WriteString(fmt.Sprintf("Commit: %s\n", result.CommitHash))

	switch {
	case result.Running:
		s.WriteString(styles.running.Render("Status: Running...") + "\n\n")
	case result.Error != nil:
		s.WriteString(styles.error.Render("Status: Failed") + "\n\n")
	default:
		duration := result.EndTime.Sub(result.StartTime)
		s.WriteString(styles.success.Render(fmt.Sprintf("Status: %s (%.2fs)", result.Status, duration.Seconds())) + "\n\n")
	}

	if result.Output != "" {
		s.WriteString("Output:\n")
		s.WriteString(strings.Repeat("-", 80) + "\n")
		s.WriteString(result.Output + "\n")
		s.WriteString(strings.Repeat("-", 80) + "\n")
	}

	s.WriteString("\nPress ESC to go back, q to quit")
	return s.String()
}

// renderGitRepoList renders the main table view of all repositories.
func (m gitModel) renderGitRepoList(styles gitUIStyles) string {
	var s strings.Builder
	s.WriteString(styles.title.Render("Git: "+m.operation) + "\n\n")

	// Table header
	headerLine := fmt.Sprintf("%-3s %-30s %-20s %-10s %-20s",
		"", "Repository", "Branch", "Commit", "Status")
	s.WriteString(styles.header.Render(headerLine) + "\n")
	s.WriteString(strings.Repeat("-", 85) + "\n")

	for i, repo := range m.repos {
		cursor := " "
		if i == m.selected {
			cursor = ">"
		}

		result := m.results[repo.Name]
		branch := "-"
		commit := "-"
		status := "pending"
		statusStyle := styles.normal

		if result != nil {
			if result.Branch != "" {
				branch = result.Branch
			}
			if result.CommitHash != "" {
				commit = result.CommitHash
			}

			switch {
			case result.Running:
				status = "running..."
				statusStyle = styles.running
			case result.Error != nil:
				status = "failed"
				statusStyle = styles.error
			default:
				status = result.Status
				statusStyle = styles.success
			}
		}

		line := fmt.Sprintf("%-3s %-30s %-20s %-10s %-20s",
			cursor, repo.Name, branch, commit, status)

		if i == m.selected {
			s.WriteString(styles.selected.Render(line) + "\n")
		} else {
			s.WriteString(statusStyle.Render(line) + "\n")
		}
	}

	s.WriteString("\n")
	if m.allDone {
		s.WriteString(styles.success.Render("All operations completed!") + "\n")
	}
	s.WriteString("\nUse ↑/↓ or j/k to navigate, Enter to view details, q to quit\n")

	return s.String()
}

// View renders the bubbletea UI.
func (m gitModel) View() string {
	if m.quitting {
		return ""
	}

	styles := getGitUIStyles()

	if m.viewingRepo != "" {
		return m.renderGitRepoDetail(styles)
	}

	return m.renderGitRepoList(styles)
}

// runGitInteractive runs git operation in interactive mode.
func runGitInteractive(repos []components.ComponentRef, executor *GitExecutor, operation string, execFunc func()) error {
	// Start execution in background
	go execFunc()

	// Start bubbletea UI
	p := tea.NewProgram(gitModel{
		repos:     repos,
		results:   make(map[string]*GitResult),
		executor:  executor,
		operation: operation,
	})

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running UI: %w", err)
	}

	return nil
}

// printGitTable prints the git results in a nice table format.
func printGitTable(repos []components.ComponentRef, results map[string]*GitResult, operation string) {
	styles := getGitUIStyles()

	fmt.Println(styles.title.Render(" Git: " + operation + " "))
	fmt.Println()

	// Print table header
	headerLine := fmt.Sprintf("%-30s %-20s %-10s %-20s",
		"Repository", "Branch", "Commit", "Status")
	fmt.Println(styles.header.Render(headerLine))
	fmt.Println(strings.Repeat("-", 85))

	// Print results
	successCount := 0
	failCount := 0

	for _, repo := range repos {
		result := results[repo.Name]
		if result == nil {
			continue
		}

		branch := result.Branch
		if branch == "" {
			branch = "-"
		}
		commit := result.CommitHash
		if commit == "" {
			commit = "-"
		}

		var statusStyle lipgloss.Style
		status := result.Status
		if result.Error != nil {
			statusStyle = styles.error
			status = "failed"
			failCount++
		} else {
			statusStyle = styles.success
			successCount++
		}

		line := fmt.Sprintf("%-30s %-20s %-10s %-20s",
			repo.Name, branch, commit, status)
		fmt.Println(statusStyle.Render(line))
	}

	fmt.Println()
	fmt.Printf("Summary: %s succeeded, %s failed\n",
		styles.success.Render(fmt.Sprintf("%d", successCount)),
		styles.error.Render(fmt.Sprintf("%d", failCount)))
}

// getRepos returns list of all repos from project.
func getRepos(proj *components.Project) []components.ComponentRef {
	repos := []components.ComponentRef{proj.ComponentRef}

	// Filter out repos that don't exist
	var existing []components.ComponentRef
	for _, repo := range repos {
		path := filepath.Join(".", repo.Name)
		if _, err := os.Stat(path); err == nil {
			existing = append(existing, repo)
		}
	}

	return existing
}

// runGitOperation is a helper to run git operations across repos.
func runGitOperation(repos []components.ComponentRef, parallel bool, opFunc func(context.Context, components.ComponentRef, string)) {
	ctx := context.Background()

	if parallel {
		var wg sync.WaitGroup
		for _, repo := range repos {
			wg.Add(1)
			go func(r components.ComponentRef) {
				defer wg.Done()
				path := filepath.Join(".", r.Name)
				opFunc(ctx, r, path)
			}(repo)
		}
		wg.Wait()
	} else {
		for _, repo := range repos {
			path := filepath.Join(".", repo.Name)
			opFunc(ctx, repo, path)
		}
	}
}

// executeGitSubcommand runs a git subcommand either interactively or non-interactively.
func executeGitSubcommand(repos []components.ComponentRef, executor *GitExecutor, operation string, execFunc func()) {
	if *gitInteractive {
		if err := runGitInteractive(repos, executor, operation, execFunc); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		execFunc()
		printGitTable(repos, executor.GetResults(), operation)
	}
}

// handleGitStatus handles the git status subcommand.
func handleGitStatus(repos []components.ComponentRef, executor *GitExecutor) {
	operation := "status"
	execFunc := func() {
		runGitOperation(repos, *gitParallel,
			func(ctx context.Context, repo components.ComponentRef, path string) {
				executor.ExecuteStatus(ctx, repo.Name, path)
			})
	}
	executeGitSubcommand(repos, executor, operation, execFunc)
}

// handleGitPull handles the git pull subcommand.
func handleGitPull(repos []components.ComponentRef, executor *GitExecutor) {
	operation := "pull"
	execFunc := func() {
		runGitOperation(repos, *gitParallel,
			func(ctx context.Context, repo components.ComponentRef, path string) {
				executor.ExecutePull(ctx, repo.Name, path)
			})
	}
	executeGitSubcommand(repos, executor, operation, execFunc)
}

// handleGitFetch handles the git fetch subcommand.
func handleGitFetch(repos []components.ComponentRef, executor *GitExecutor) {
	operation := "fetch"
	execFunc := func() {
		runGitOperation(repos, *gitParallel,
			func(ctx context.Context, repo components.ComponentRef, path string) {
				executor.ExecuteFetch(ctx, repo.Name, path)
			})
	}
	executeGitSubcommand(repos, executor, operation, execFunc)
}

// handleGitCheckout handles the git checkout subcommand.
func handleGitCheckout(repos []components.ComponentRef, executor *GitExecutor, subArgs []string) {
	if len(subArgs) == 0 {
		fmt.Println("Error: branch name required")
		fmt.Println("Usage: monhang git checkout <branch>")
		os.Exit(1)
	}
	branch := subArgs[0]
	operation := fmt.Sprintf("checkout %s", branch)
	execFunc := func() {
		runGitOperation(repos, *gitParallel,
			func(ctx context.Context, repo components.ComponentRef, path string) {
				executor.ExecuteCheckout(ctx, repo.Name, path, branch)
			})
	}
	executeGitSubcommand(repos, executor, operation, execFunc)
}

// handleGitBranch handles the git branch subcommand.
func handleGitBranch(repos []components.ComponentRef, executor *GitExecutor, subArgs []string) {
	if len(subArgs) == 0 {
		fmt.Println("Error: branch name required")
		fmt.Println("Usage: monhang git branch <branch>")
		os.Exit(1)
	}
	branch := subArgs[0]
	operation := fmt.Sprintf("branch %s", branch)
	execFunc := func() {
		runGitOperation(repos, *gitParallel,
			func(ctx context.Context, repo components.ComponentRef, path string) {
				executor.ExecuteBranch(ctx, repo.Name, path, branch)
			})
	}
	executeGitSubcommand(repos, executor, operation, execFunc)
}

func runGit(_ *Command, args []string) {
	if len(args) == 0 {
		fmt.Println("Error: No subcommand specified")
		fmt.Println("Usage: monhang git <subcommand> [args]")
		fmt.Println("\nSubcommands: status, pull, fetch, checkout, branch")
		os.Exit(1)
	}

	subcommand := args[0]
	subArgs := args[1:]

	logging.GetLogger("git").Info().
		Str("subcommand", subcommand).
		Strs("args", subArgs).
		Bool("interactive", *gitInteractive).
		Bool("parallel", *gitParallel).
		Msg("Starting git command")

	// Try to parse configuration file
	proj, err := components.ParseProjectFile(*gitF)
	if err != nil {
		// If config file doesn't exist, check if current directory is a git repo
		if isGitRepository(".") {
			logging.GetLogger("git").Debug().Msg("No config file found, using current directory as git repository")
			// Create a minimal project with just the current directory
			// Use "." as the name so getRepos() will find it in the current directory
			proj = components.CreateLocalProject(".")
		} else {
			// Not a git repo and no config file - fail
			Check(err)
		}
	}

	repos := getRepos(proj)

	if len(repos) == 0 {
		logging.GetLogger("git").Warn().Msg("No repositories found")
		fmt.Println("No repositories found (or none have been cloned yet)")
		fmt.Println("Run 'monhang boot' first to clone repositories")
		return
	}

	logging.GetLogger("git").Debug().Int("repo_count", len(repos)).Msg("Repositories loaded")

	executor := NewGitExecutor()

	// Execute based on subcommand
	switch subcommand {
	case "status":
		handleGitStatus(repos, executor)
	case "pull":
		handleGitPull(repos, executor)
	case "fetch":
		handleGitFetch(repos, executor)
	case "checkout":
		handleGitCheckout(repos, executor, subArgs)
	case "branch":
		handleGitBranch(repos, executor, subArgs)
	default:
		logging.GetLogger("git").Error().Str("subcommand", subcommand).Msg("Unknown git subcommand")
		fmt.Printf("Error: Unknown subcommand '%s'\n", subcommand)
		fmt.Println("Available subcommands: status, pull, fetch, checkout, branch")
		os.Exit(1)
	}

	logging.GetLogger("git").Info().Str("subcommand", subcommand).Msg("Git command completed")
}

func init() {
	CmdGit.Run = runGit
}
