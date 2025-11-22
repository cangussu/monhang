// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

package commands

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
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

// CmdExec is the exec command for running arbitrary commands in each repo.
var CmdExec = &Command{
	Name:  "exec",
	Args:  "[command...]",
	Short: "run arbitrary commands inside each repo",
	Long: `
Exec runs arbitrary commands inside each repository defined in the configuration file.

Examples:
	monhang exec -- git status
	monhang exec -p -- make build
	monhang exec -i -- npm test

Options:
	-f <file>    configuration file (default: ./monhang.json)
	-p           run commands in parallel
	-i           interactive mode with bubbletea UI
	-l           show live progress (non-interactive)
`,
}

var (
	execF           = CmdExec.Flag.String("f", "./monhang.json", "configuration file")
	execParallel    = CmdExec.Flag.Bool("p", false, "run in parallel")
	execInteractive = CmdExec.Flag.Bool("i", false, "interactive mode")
	execLive        = CmdExec.Flag.Bool("l", false, "show live progress")
)

// RepoResult holds the result of running a command in a repo.
type RepoResult struct {
	StartTime time.Time
	EndTime   time.Time
	Error     error
	Name      string
	Path      string
	Command   string
	Output    string
	ExitCode  int
	Running   bool
}

// RepoExecutor manages command execution across repositories.
type RepoExecutor struct {
	results    map[string]*RepoResult
	mu         sync.RWMutex
	liveOutput bool
}

// NewRepoExecutor creates a new RepoExecutor.
func NewRepoExecutor(liveOutput bool) *RepoExecutor {
	return &RepoExecutor{
		results:    make(map[string]*RepoResult),
		liveOutput: liveOutput,
	}
}

// ExecuteCommand runs a command in a repository directory.
func (re *RepoExecutor) ExecuteCommand(ctx context.Context, name, path, command string) {
	logging.GetLogger("exec").Debug().
		Str("repo", name).
		Str("path", path).
		Str("command", command).
		Msg("Starting command execution")

	result := &RepoResult{
		Name:      name,
		Path:      path,
		Command:   command,
		StartTime: time.Now(),
		Running:   true,
	}

	re.mu.Lock()
	re.results[name] = result
	re.mu.Unlock()

	// Execute command
	cmdArgs := strings.Fields(command)
	if len(cmdArgs) == 0 {
		logging.GetLogger("exec").Error().Str("repo", name).Msg("Empty command provided")
		result.Error = fmt.Errorf("empty command")
		result.Running = false
		result.EndTime = time.Now()
		return
	}

	// #nosec G204 -- command is intentionally provided by the user
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = path

	var outputBuf bytes.Buffer
	var stdoutBuf, stderrBuf bytes.Buffer

	if re.liveOutput {
		// Show live output
		stdout := io.MultiWriter(&stdoutBuf, &outputBuf)
		stderr := io.MultiWriter(&stderrBuf, &outputBuf)
		cmd.Stdout = stdout
		cmd.Stderr = stderr

		// Print live output
		go func() {
			scanner := bufio.NewScanner(&stdoutBuf)
			for scanner.Scan() {
				line := scanner.Text()
				re.mu.Lock()
				result.Output += line + "\n"
				re.mu.Unlock()
				fmt.Printf("[%s] %s\n", name, line)
			}
		}()
	} else {
		cmd.Stdout = &outputBuf
		cmd.Stderr = &outputBuf
	}

	err := cmd.Run()

	re.mu.Lock()
	defer re.mu.Unlock()

	result.Output = outputBuf.String()
	result.Error = err
	result.Running = false
	result.EndTime = time.Now()

	duration := result.EndTime.Sub(result.StartTime)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
		logging.GetLogger("exec").Error().
			Str("repo", name).
			Int("exit_code", result.ExitCode).
			Dur("duration", duration).
			Err(err).
			Msg("Command execution failed")
	} else {
		logging.GetLogger("exec").Debug().
			Str("repo", name).
			Dur("duration", duration).
			Msg("Command execution completed successfully")
	}
}

// GetResults returns a copy of all results.
func (re *RepoExecutor) GetResults() map[string]*RepoResult {
	re.mu.RLock()
	defer re.mu.RUnlock()

	results := make(map[string]*RepoResult)
	for k, v := range re.results {
		// Create a copy
		resultCopy := *v
		results[k] = &resultCopy
	}
	return results
}

// Bubbletea Model for interactive mode
//
//nolint:govet // fieldalignment: UI model where field order doesn't significantly impact performance
type model struct {
	repos       []components.ComponentRef
	results     map[string]*RepoResult
	executor    *RepoExecutor
	command     string
	viewingRepo string
	selected    int
	width       int
	height      int
	quitting    bool
	allDone     bool
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Init initializes the bubbletea model.
func (m model) Init() tea.Cmd {
	return tickCmd()
}

// Update handles bubbletea messages and updates the model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case tickMsg:
		m.results = m.executor.GetResults()

		// Check if all are done
		allDone := true
		for _, result := range m.results {
			if result.Running {
				allDone = false
				break
			}
		}
		m.allDone = allDone

		return m, tickCmd()
	}

	return m, nil
}

// uiStyles holds the lipgloss styles used in the UI.
type uiStyles struct {
	title    lipgloss.Style
	selected lipgloss.Style
	normal   lipgloss.Style
	success  lipgloss.Style
	error    lipgloss.Style
	running  lipgloss.Style
}

// getUIStyles returns the configured UI styles.
func getUIStyles() uiStyles {
	return uiStyles{
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1),
		selected: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			PaddingLeft(2),
		normal: lipgloss.NewStyle().
			PaddingLeft(2),
		success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")),
		error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")),
		running: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00")),
	}
}

// renderRepoDetail renders the detailed view for a specific repository.
func (m model) renderRepoDetail(styles uiStyles) string {
	result := m.results[m.viewingRepo]
	if result == nil {
		return ""
	}

	var s strings.Builder
	s.WriteString(styles.title.Render("Repository: "+m.viewingRepo) + "\n\n")
	s.WriteString(fmt.Sprintf("Command: %s\n", m.command))
	s.WriteString(fmt.Sprintf("Path: %s\n", result.Path))

	switch {
	case result.Running:
		s.WriteString(styles.running.Render("Status: Running...") + "\n\n")
	case result.Error != nil:
		s.WriteString(styles.error.Render(fmt.Sprintf("Status: Failed (exit code: %d)", result.ExitCode)) + "\n\n")
	default:
		duration := result.EndTime.Sub(result.StartTime)
		s.WriteString(styles.success.Render(fmt.Sprintf("Status: Success (%.2fs)", duration.Seconds())) + "\n\n")
	}

	s.WriteString("Output:\n")
	s.WriteString(strings.Repeat("-", 80) + "\n")
	s.WriteString(result.Output)
	s.WriteString("\n" + strings.Repeat("-", 80) + "\n")
	s.WriteString("\nPress ESC to go back, q to quit")

	return s.String()
}

// renderRepoList renders the main list view of all repositories.
func (m model) renderRepoList(styles uiStyles) string {
	var s strings.Builder
	s.WriteString(styles.title.Render("Executing: "+m.command) + "\n\n")

	for i, repo := range m.repos {
		cursor := " "
		if i == m.selected {
			cursor = ">"
		}

		result := m.results[repo.Name]
		status := "pending"
		statusStyle := styles.normal

		if result != nil {
			switch {
			case result.Running:
				status = "running..."
				statusStyle = styles.running
			case result.Error != nil:
				status = fmt.Sprintf("failed (exit: %d)", result.ExitCode)
				statusStyle = styles.error
			default:
				duration := result.EndTime.Sub(result.StartTime)
				status = fmt.Sprintf("success (%.2fs)", duration.Seconds())
				statusStyle = styles.success
			}
		}

		line := fmt.Sprintf("%s %-30s %s", cursor, repo.Name, status)

		if i == m.selected {
			s.WriteString(styles.selected.Render(line) + "\n")
		} else {
			s.WriteString(statusStyle.Render(line) + "\n")
		}
	}

	s.WriteString("\n")
	if m.allDone {
		s.WriteString(styles.success.Render("All commands completed!") + "\n")
	}
	s.WriteString("\nUse ↑/↓ or j/k to navigate, Enter to view output, q to quit\n")

	return s.String()
}

// View renders the bubbletea UI.
func (m model) View() string {
	if m.quitting {
		return ""
	}

	styles := getUIStyles()

	// If viewing a specific repo's output
	if m.viewingRepo != "" {
		return m.renderRepoDetail(styles)
	}

	// Main list view
	return m.renderRepoList(styles)
}

// RunInteractiveMode starts command execution in interactive mode with bubbletea UI.
func RunInteractiveMode(repos []components.ComponentRef, executor *RepoExecutor, command string, parallel bool) error {
	ctx := context.Background()

	// Start all executions
	if parallel {
		var wg sync.WaitGroup
		for _, repo := range repos {
			wg.Add(1)
			go func(r components.ComponentRef) {
				defer wg.Done()
				path := filepath.Join(".", r.Name)
				executor.ExecuteCommand(ctx, r.Name, path, command)
			}(repo)
		}

		// Wait in background
		go func() {
			wg.Wait()
		}()
	} else {
		go func() {
			for _, repo := range repos {
				path := filepath.Join(".", repo.Name)
				executor.ExecuteCommand(ctx, repo.Name, path, command)
			}
		}()
	}

	// Start bubbletea UI
	p := tea.NewProgram(model{
		repos:    repos,
		results:  make(map[string]*RepoResult),
		executor: executor,
		command:  command,
	})

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running UI: %w", err)
	}

	return nil
}

// runNonInteractiveMode executes commands in non-interactive mode.
func runNonInteractiveMode(repos []components.ComponentRef, executor *RepoExecutor, command string, parallel bool) {
	ctx := context.Background()

	if parallel {
		logging.GetLogger("exec").Info().Int("repo_count", len(repos)).Msg("Executing commands in parallel")
		var wg sync.WaitGroup
		for _, repo := range repos {
			wg.Add(1)
			go func(r components.ComponentRef) {
				defer wg.Done()
				path := filepath.Join(".", r.Name)
				logging.GetLogger("exec").Info().Str("repo", r.Name).Str("command", command).Msg("Executing command")
				executor.ExecuteCommand(ctx, r.Name, path, command)
			}(repo)
		}
		wg.Wait()
	} else {
		logging.GetLogger("exec").Info().Int("repo_count", len(repos)).Msg("Executing commands sequentially")
		for _, repo := range repos {
			path := filepath.Join(".", repo.Name)
			logging.GetLogger("exec").Info().Str("repo", repo.Name).Str("command", command).Msg("Executing command")
			executor.ExecuteCommand(ctx, repo.Name, path, command)
		}
	}
}

// PrintResults prints the execution results for all repos.
func PrintResults(repos []components.ComponentRef, results map[string]*RepoResult) int {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("EXECUTION RESULTS")
	fmt.Println(strings.Repeat("=", 80))

	successCount := 0
	failCount := 0

	for _, repo := range repos {
		result := results[repo.Name]
		if result == nil {
			continue
		}

		fmt.Printf("\n[%s]\n", result.Name)
		fmt.Printf("Path: %s\n", result.Path)
		fmt.Printf("Command: %s\n", result.Command)

		if result.Error != nil {
			fmt.Printf("Status: FAILED (exit code: %d)\n", result.ExitCode)
			failCount++
		} else {
			duration := result.EndTime.Sub(result.StartTime)
			fmt.Printf("Status: SUCCESS (%.2fs)\n", duration.Seconds())
			successCount++
		}

		if result.Output != "" {
			fmt.Println("\nOutput:")
			fmt.Println(strings.Repeat("-", 80))
			fmt.Println(result.Output)
			fmt.Println(strings.Repeat("-", 80))
		}
	}

	fmt.Printf("\n\nSummary: %d succeeded, %d failed\n", successCount, failCount)
	return failCount
}

func runExec(_ *Command, args []string) {
	if len(args) == 0 {
		fmt.Println("Error: No command specified")
		fmt.Println("Usage: monhang exec [flags] -- <command>")
		os.Exit(1)
	}

	command := strings.Join(args, " ")
	logging.GetLogger("exec").Info().
		Str("command", command).
		Bool("parallel", *execParallel).
		Bool("interactive", *execInteractive).
		Msg("Starting exec command")

	// Parse the configuration file
	proj, err := components.ParseProjectFile(*execF)
	if err != nil {
		logging.GetLogger("exec").Error().Err(err).Str("filename", *execF).Msg("Failed to parse project file")
		Check(err)
	}

	// Build list of repos
	repos := []components.ComponentRef{proj.ComponentRef}

	if len(repos) == 0 {
		logging.GetLogger("exec").Warn().Msg("No repositories found in configuration")
		fmt.Println("No repositories found in configuration")
		return
	}

	logging.GetLogger("exec").Debug().Int("repo_count", len(repos)).Msg("Repositories loaded")

	executor := NewRepoExecutor(*execLive && !*execInteractive)

	// Interactive mode with bubbletea
	if *execInteractive {
		logging.GetLogger("exec").Debug().Msg("Running in interactive mode")
		if err := RunInteractiveMode(repos, executor, command, *execParallel); err != nil {
			logging.GetLogger("exec").Error().Err(err).Msg("Interactive mode failed")
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Non-interactive mode
	runNonInteractiveMode(repos, executor, command, *execParallel)

	// Print results
	failCount := PrintResults(repos, executor.GetResults())

	logging.GetLogger("exec").Info().
		Int("total", len(repos)).
		Int("failed", failCount).
		Int("succeeded", len(repos)-failCount).
		Msg("Execution summary")

	if failCount > 0 {
		os.Exit(1)
	}
}

func init() {
	CmdExec.Run = runExec // break init loop
}
