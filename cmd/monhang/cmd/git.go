// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/cangussu/monhang/internal/commands"
	"github.com/cangussu/monhang/internal/components"
	"github.com/cangussu/monhang/internal/logging"
	"github.com/spf13/cobra"
)

var (
	gitConfigFile  string
	gitInteractive bool
	gitParallel    bool
)

// gitCmd represents the git command.
var gitCmd = &cobra.Command{
	Use:   "git <subcommand> [args...]",
	Short: "Run git operations across all repos",
	Long: `Git runs git operations across all repositories defined in the configuration file.

Available subcommands:
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
  monhang git status -i      # interactive mode
  monhang git pull -p        # parallel execution`,
	Run: runGit,
}

func init() {
	rootCmd.AddCommand(gitCmd)
	gitCmd.Flags().StringVarP(&gitConfigFile, "file", "f", "./monhang.json", "configuration file")
	gitCmd.Flags().BoolVarP(&gitInteractive, "interactive", "i", false, "interactive mode with bubbletea UI")
	gitCmd.Flags().BoolVarP(&gitParallel, "parallel", "p", false, "run operations in parallel")
}

func runGit(cmd *cobra.Command, args []string) {
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
		Bool("interactive", gitInteractive).
		Bool("parallel", gitParallel).
		Msg("Starting git command")

	// Try to parse configuration file
	proj, err := components.ParseProjectFile(gitConfigFile)
	if err != nil {
		// If config file doesn't exist, check if current directory is a git repo
		if commands.IsGitRepository(".") {
			logging.GetLogger("git").Debug().Msg("No config file found, using current directory as git repository")
			// Create a minimal project with just the current directory
			proj = components.CreateLocalProject(".")
		} else {
			// Not a git repo and no config file - fail
			Check(err)
		}
	}

	repos := getGitRepos(proj)

	if len(repos) == 0 {
		logging.GetLogger("git").Warn().Msg("No repositories found")
		fmt.Println("No repositories found (or none have been cloned yet)")
		fmt.Println("Run 'monhang boot' first to clone repositories")
		return
	}

	logging.GetLogger("git").Debug().Int("repo_count", len(repos)).Msg("Repositories loaded")

	executor := commands.NewGitExecutor()

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

// getGitRepos returns list of all repos from project.
func getGitRepos(proj *components.Project) []components.ComponentRef {
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
func executeGitSubcommand(repos []components.ComponentRef, executor *commands.GitExecutor, operation string, execFunc func()) {
	if gitInteractive {
		if err := commands.RunGitInteractive(repos, executor, operation, execFunc); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		execFunc()
		commands.PrintGitTable(repos, executor.GetResults(), operation)
	}
}

// handleGitStatus handles the git status subcommand.
func handleGitStatus(repos []components.ComponentRef, executor *commands.GitExecutor) {
	operation := "status"
	execFunc := func() {
		runGitOperation(repos, gitParallel,
			func(ctx context.Context, repo components.ComponentRef, path string) {
				executor.ExecuteStatus(ctx, repo.Name, path)
			})
	}
	executeGitSubcommand(repos, executor, operation, execFunc)
}

// handleGitPull handles the git pull subcommand.
func handleGitPull(repos []components.ComponentRef, executor *commands.GitExecutor) {
	operation := "pull"
	execFunc := func() {
		runGitOperation(repos, gitParallel,
			func(ctx context.Context, repo components.ComponentRef, path string) {
				executor.ExecutePull(ctx, repo.Name, path)
			})
	}
	executeGitSubcommand(repos, executor, operation, execFunc)
}

// handleGitFetch handles the git fetch subcommand.
func handleGitFetch(repos []components.ComponentRef, executor *commands.GitExecutor) {
	operation := "fetch"
	execFunc := func() {
		runGitOperation(repos, gitParallel,
			func(ctx context.Context, repo components.ComponentRef, path string) {
				executor.ExecuteFetch(ctx, repo.Name, path)
			})
	}
	executeGitSubcommand(repos, executor, operation, execFunc)
}

// handleGitCheckout handles the git checkout subcommand.
func handleGitCheckout(repos []components.ComponentRef, executor *commands.GitExecutor, subArgs []string) {
	if len(subArgs) == 0 {
		fmt.Println("Error: branch name required")
		fmt.Println("Usage: monhang git checkout <branch>")
		os.Exit(1)
	}
	branch := subArgs[0]
	operation := fmt.Sprintf("checkout %s", branch)
	execFunc := func() {
		runGitOperation(repos, gitParallel,
			func(ctx context.Context, repo components.ComponentRef, path string) {
				executor.ExecuteCheckout(ctx, repo.Name, path, branch)
			})
	}
	executeGitSubcommand(repos, executor, operation, execFunc)
}

// handleGitBranch handles the git branch subcommand.
func handleGitBranch(repos []components.ComponentRef, executor *commands.GitExecutor, subArgs []string) {
	if len(subArgs) == 0 {
		fmt.Println("Error: branch name required")
		fmt.Println("Usage: monhang git branch <branch>")
		os.Exit(1)
	}
	branch := subArgs[0]
	operation := fmt.Sprintf("branch %s", branch)
	execFunc := func() {
		runGitOperation(repos, gitParallel,
			func(ctx context.Context, repo components.ComponentRef, path string) {
				executor.ExecuteBranch(ctx, repo.Name, path, branch)
			})
	}
	executeGitSubcommand(repos, executor, operation, execFunc)
}
