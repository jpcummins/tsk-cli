package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"github.com/jpcummins/tsk-cli/ui"
	"github.com/jpcummins/tsk-lib/engine"
)

func main() {
	repo := flag.String("repo", "", "Path to a tsk repository (defaults to current directory)")
	me := flag.String("me", "", "Current user identity for me() query resolution (e.g. alice@example.com)")
	flag.Parse()

	repoPath := *repo
	if repoPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: could not determine current directory: %v\n", err)
			os.Exit(1)
		}
		repoPath = cwd
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid path %q: %v\n", repoPath, err)
		os.Exit(1)
	}

	// Verify tasks/ directory exists
	tasksDir := filepath.Join(absPath, "tasks")
	if info, err := os.Stat(tasksDir); err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "error: no tasks/ directory found in %s\n", absPath)
		fmt.Fprintf(os.Stderr, "hint: use --repo to specify the path to a tsk repository\n")
		os.Exit(1)
	}

	// Build engine options
	var opts []engine.Option
	if *me != "" {
		opts = append(opts, engine.WithCurrentUser(*me))
	}

	eng, err := engine.NewDefault(":memory:", opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not initialize engine: %v\n", err)
		os.Exit(1)
	}
	defer eng.Close()

	// Index the repository
	repo_, err := eng.Index(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not index repository: %v\n", err)
		os.Exit(1)
	}

	// Print diagnostics to stderr (non-fatal)
	for _, d := range repo_.Diagnostics.Warnings() {
		fmt.Fprintf(os.Stderr, "warning: %s\n", d.Message)
	}
	for _, d := range repo_.Diagnostics.Errors() {
		fmt.Fprintf(os.Stderr, "error: %s\n", d.Message)
	}

	m := ui.NewModel(eng, absPath)
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
