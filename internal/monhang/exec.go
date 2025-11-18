// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

package monhang

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
	Name      string
	Path      string
	Command   string
	Output    string
	Error     error
	ExitCode  int
	StartTime time.Time
	EndTime   time.Time
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
		result.Error = fmt.Errorf("empty command")
		result.Running = false
		result.EndTime = time.Now()
		return
	}

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

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
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

type model struct {
	repos       []ComponentRef
	results     map[string]*RepoResult
	selected    int
	executor    *RepoExecutor
	command     string
	quitting    bool
	width       int
	height      int
	allDone     bool
	viewingRepo string
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
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

func (m model) View() string {
	if m.quitting {
		return ""
	}

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	selectedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		PaddingLeft(2)

	normalStyle := lipgloss.NewStyle().
		PaddingLeft(2)

	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575"))

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000"))

	runningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFAA00"))

	var s strings.Builder

	// If viewing a specific repo's output
	if m.viewingRepo != "" {
		result := m.results[m.viewingRepo]
		if result != nil {
			s.WriteString(titleStyle.Render("Repository: "+m.viewingRepo) + "\n\n")
			s.WriteString(fmt.Sprintf("Command: %s\n", m.command))
			s.WriteString(fmt.Sprintf("Path: %s\n", result.Path))

			if result.Running {
				s.WriteString(runningStyle.Render("Status: Running...") + "\n\n")
			} else if result.Error != nil {
				s.WriteString(errorStyle.Render(fmt.Sprintf("Status: Failed (exit code: %d)", result.ExitCode)) + "\n\n")
			} else {
				duration := result.EndTime.Sub(result.StartTime)
				s.WriteString(successStyle.Render(fmt.Sprintf("Status: Success (%.2fs)", duration.Seconds())) + "\n\n")
			}

			s.WriteString("Output:\n")
			s.WriteString(strings.Repeat("-", 80) + "\n")
			s.WriteString(result.Output)
			s.WriteString("\n" + strings.Repeat("-", 80) + "\n")
			s.WriteString("\nPress ESC to go back, q to quit")
		}
		return s.String()
	}

	// Main list view
	s.WriteString(titleStyle.Render("Executing: "+m.command) + "\n\n")

	for i, repo := range m.repos {
		cursor := " "
		if i == m.selected {
			cursor = ">"
		}

		result := m.results[repo.Name]
		status := "pending"
		statusStyle := normalStyle

		if result != nil {
			if result.Running {
				status = "running..."
				statusStyle = runningStyle
			} else if result.Error != nil {
				status = fmt.Sprintf("failed (exit: %d)", result.ExitCode)
				statusStyle = errorStyle
			} else {
				duration := result.EndTime.Sub(result.StartTime)
				status = fmt.Sprintf("success (%.2fs)", duration.Seconds())
				statusStyle = successStyle
			}
		}

		line := fmt.Sprintf("%s %-30s %s", cursor, repo.Name, status)

		if i == m.selected {
			s.WriteString(selectedStyle.Render(line) + "\n")
		} else {
			s.WriteString(statusStyle.Render(line) + "\n")
		}
	}

	s.WriteString("\n")
	if m.allDone {
		s.WriteString(successStyle.Render("All commands completed!") + "\n")
	}
	s.WriteString("\nUse ↑/↓ or j/k to navigate, Enter to view output, q to quit\n")

	return s.String()
}

func runExec(_ *Command, args []string) {
	if len(args) == 0 {
		fmt.Println("Error: No command specified")
		fmt.Println("Usage: monhang exec [flags] -- <command>")
		os.Exit(1)
	}

	command := strings.Join(args, " ")

	// Parse the configuration file
	proj, err := ParseProjectFile(*execF)
	if err != nil {
		Check(err)
	}

	// Build list of repos
	proj.ProcessDeps()
	repos := []ComponentRef{proj.ComponentRef}
	repos = append(repos, proj.Deps.Build...)
	repos = append(repos, proj.Deps.Runtime...)
	repos = append(repos, proj.Deps.Intall...)

	if len(repos) == 0 {
		fmt.Println("No repositories found in configuration")
		return
	}

	executor := NewRepoExecutor(*execLive && !*execInteractive)

	// Interactive mode with bubbletea
	if *execInteractive {
		ctx := context.Background()

		// Start all executions
		if *execParallel {
			var wg sync.WaitGroup
			for _, repo := range repos {
				wg.Add(1)
				repo := repo // capture loop variable
				go func() {
					defer wg.Done()
					path := filepath.Join(".", repo.Name)
					executor.ExecuteCommand(ctx, repo.Name, path, command)
				}()
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
			fmt.Printf("Error running UI: %v\n", err)
			os.Exit(1)
		}

		return
	}

	// Non-interactive mode
	ctx := context.Background()

	if *execParallel {
		var wg sync.WaitGroup
		for _, repo := range repos {
			wg.Add(1)
			repo := repo // capture loop variable
			go func() {
				defer wg.Done()
				path := filepath.Join(".", repo.Name)
				mglog.Infof("Executing in %s: %s", repo.Name, command)
				executor.ExecuteCommand(ctx, repo.Name, path, command)
			}()
		}
		wg.Wait()
	} else {
		for _, repo := range repos {
			path := filepath.Join(".", repo.Name)
			mglog.Infof("Executing in %s: %s", repo.Name, command)
			executor.ExecuteCommand(ctx, repo.Name, path, command)
		}
	}

	// Print results
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("EXECUTION RESULTS")
	fmt.Println(strings.Repeat("=", 80))

	results := executor.GetResults()
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

	if failCount > 0 {
		os.Exit(1)
	}
}

func init() {
	CmdExec.Run = runExec // break init loop
}
