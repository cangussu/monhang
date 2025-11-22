// Copyright 2016 Thiago Cangussu de Castro Gomes. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// version 3 that can be found in the LICENSE file.

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cangussu/monhang/internal/commands"
	"github.com/cangussu/monhang/internal/components"
	"github.com/cangussu/monhang/internal/logging"
	"github.com/spf13/cobra"
)

var (
	execConfigFile  string
	execParallel    bool
	execInteractive bool
	execLive        bool
)

// execCmd represents the exec command.
var execCmd = &cobra.Command{
	Use:   "exec [flags] -- <command>",
	Short: "Run arbitrary commands inside each repo",
	Long: `Exec runs arbitrary commands inside each repository defined in the configuration file.

The command to execute must be provided after the -- separator to prevent flag parsing issues.

Examples:
  monhang exec -- git status
  monhang exec -p -- make build
  monhang exec -i -- npm test
  monhang exec -l -- go test ./...`,
	Run: runExec,
}

func init() {
	rootCmd.AddCommand(execCmd)
	execCmd.Flags().StringVarP(&execConfigFile, "file", "f", "./monhang.json", "configuration file")
	execCmd.Flags().BoolVarP(&execParallel, "parallel", "p", false, "run commands in parallel")
	execCmd.Flags().BoolVarP(&execInteractive, "interactive", "i", false, "interactive mode with bubbletea UI")
	execCmd.Flags().BoolVarP(&execLive, "live", "l", false, "show live progress (non-interactive)")
}

func runExec(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("Error: No command specified")
		fmt.Println("Usage: monhang exec [flags] -- <command>")
		os.Exit(1)
	}

	command := strings.Join(args, " ")
	logging.GetLogger("exec").Info().
		Str("command", command).
		Bool("parallel", execParallel).
		Bool("interactive", execInteractive).
		Msg("Starting exec command")

	// Parse the configuration file
	proj, err := components.ParseProjectFile(execConfigFile)
	if err != nil {
		logging.GetLogger("exec").Error().Err(err).Str("filename", execConfigFile).Msg("Failed to parse project file")
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

	executor := commands.NewRepoExecutor(execLive && !execInteractive)

	// Interactive mode with bubbletea
	if execInteractive {
		logging.GetLogger("exec").Debug().Msg("Running in interactive mode")
		if err := runInteractiveMode(repos, executor, command, execParallel); err != nil {
			logging.GetLogger("exec").Error().Err(err).Msg("Interactive mode failed")
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Non-interactive mode
	runNonInteractiveMode(repos, executor, command, execParallel)

	// Print results
	failCount := commands.PrintResults(repos, executor.GetResults())

	logging.GetLogger("exec").Info().
		Int("total", len(repos)).
		Int("failed", failCount).
		Int("succeeded", len(repos)-failCount).
		Msg("Execution summary")

	if failCount > 0 {
		os.Exit(1)
	}
}

// runInteractiveMode starts command execution in interactive mode with bubbletea UI.
func runInteractiveMode(repos []components.ComponentRef, executor *commands.RepoExecutor, command string, parallel bool) error {
	return commands.RunInteractiveMode(repos, executor, command, parallel)
}

// runNonInteractiveMode executes commands in non-interactive mode.
func runNonInteractiveMode(repos []components.ComponentRef, executor *commands.RepoExecutor, command string, parallel bool) {
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
