package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
)

func main() {
	var preview bool
	var rootCmd = &cobra.Command{
		Use:   "gwt",
		Short: "Git Worktree Manager",
		Run:   func(cmd *cobra.Command, args []string) { runJump(cmd, args, preview) },
	}
	rootCmd.Flags().BoolVar(&preview, "preview", false, "Show git status preview")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "add [branch]",
		Short: "Create a worktree from a branch",
		Args:  cobra.ExactArgs(1),
		Run:   runAdd,
	})

	var removePreview bool
	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Interactively remove a worktree",
		Run:   func(cmd *cobra.Command, args []string) { runRemove(cmd, args, removePreview) },
	}
	removeCmd.Flags().BoolP("force", "f", false, "Force removal")
	removeCmd.Flags().BoolVar(&removePreview, "preview", false, "Show git status preview")
	rootCmd.AddCommand(removeCmd)

	rootCmd.AddCommand(&cobra.Command{
		Use:     "main",
		Aliases: []string{"master"},
		Short:   "Jump to default branch worktree",
		Run:     runJumpDefault,
	})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// --- Handlers ---

func runJump(cmd *cobra.Command, args []string, preview bool) {
	if len(args) > 0 {
		path, err := findWorktreePathForBranch(args[0])
		if err != nil {
			fail(err)
		}
		printPath(path)
		return
	}

	worktrees, err := getWorktreesSorted()
	if err != nil {
		fail(err)
	}

	var opts []fuzzyfinder.Option
	if preview {
		opts = append(opts, fuzzyfinder.WithPreviewWindow(func(i int, w, h int) string {
			if i == -1 {
				return ""
			}
			out, _ := exec.Command("git", "-C", worktrees[i].AbsPath, "status", "--short").Output()
			return string(out)
		}))
	}

	idx, err := fuzzyfinder.Find(
		worktrees,
		func(i int) string { return worktrees[i].Display },
		opts...,
	)
	if err != nil {
		os.Exit(1)
	}

	printPath(worktrees[idx].AbsPath)
}

func runAdd(cmd *cobra.Command, args []string) {
	branch := args[0]
	cwd, _ := os.Getwd()
	repoName := filepath.Base(cwd)
	sanitized := strings.ReplaceAll(branch, "/", "_")
	newPath := filepath.Join("..", fmt.Sprintf("%s_%s", repoName, sanitized))

	gitCmd := exec.Command("git", "worktree", "add", newPath, branch)
	gitCmd.Stderr = os.Stderr
	if err := gitCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Branch not found. Create '%s'? [y/N] ", branch)
		reader := bufio.NewReader(os.Stdin)
		res, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(res)) == "y" {
			gitCmd = exec.Command("git", "worktree", "add", "-b", branch, newPath)
			gitCmd.Stderr = os.Stderr
			if err := gitCmd.Run(); err != nil {
				fail(fmt.Errorf("failed to create worktree"))
			}
		} else {
			os.Exit(1)
		}
	}
	printPath(newPath)
}

func runRemove(cmd *cobra.Command, args []string, preview bool) {
	worktrees, err := getWorktreesSorted()
	if err != nil {
		fail(err)
	}

	var opts []fuzzyfinder.Option
	if preview {
		opts = append(opts, fuzzyfinder.WithPreviewWindow(func(i int, w, h int) string {
			if i == -1 {
				return ""
			}
			out, _ := exec.Command("git", "-C", worktrees[i].AbsPath, "status", "--short").Output()
			return string(out)
		}))
	}

	// Use FindMulti instead of Find
	idxs, err := fuzzyfinder.FindMulti(
		worktrees,
		func(i int) string { return worktrees[i].Display },
		opts...,
	)
	if err != nil {
		os.Exit(1)
	}

	// Collect paths to remove
	var pathsToRemove []string
	for _, i := range idxs {
		pathsToRemove = append(pathsToRemove, worktrees[i].AbsPath)
	}

	// Confirm bulk action
	fmt.Fprintf(os.Stderr, "Remove %d worktree(s)?\n", len(pathsToRemove))
	for _, p := range pathsToRemove {
		fmt.Fprintf(os.Stderr, " - %s\n", p)
	}
	fmt.Fprintf(os.Stderr, "[y/N] ")

	reader := bufio.NewReader(os.Stdin)
	res, _ := reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(res)) != "y" {
		os.Exit(1)
	}

	// Execute removals
	force, _ := cmd.Flags().GetBool("force")
	baseArgs := []string{"worktree", "remove"}
	if force {
		baseArgs = append(baseArgs, "--force")
	}

	for _, path := range pathsToRemove {
		args := append(baseArgs, path)
		c := exec.Command("git", args...)
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to remove %s: %v\n", path, err)
		} else {
			fmt.Fprintf(os.Stderr, "Removed %s\n", path)
		}
	}

	// Finally, jump to repo root (so we don't stay in a deleted dir)
	out, _ := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	printPath(strings.TrimSpace(string(out)))
}

func runJumpDefault(cmd *cobra.Command, args []string) {
	for _, b := range []string{"main", "master"} {
		if p, err := findWorktreePathForBranch(b); err == nil {
			printPath(p)
			return
		}
	}
	out, _ := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	printPath(strings.TrimSpace(string(out)))
}

// --- Helpers ---

type WorktreeInfo struct {
	AbsPath string // absolute path for git commands and cd
	Display string // worktree name for fuzzy finder UI
	Commit  string // commit hash
	Branch  string // branch name (if any)
	IsCwd   bool   // true if this is the current working directory
}

func getWorktreesSorted() ([]WorktreeInfo, error) {
	out, err := exec.Command("git", "worktree", "list").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")

	// Get Current Working Directory
	cwd, _ := os.Getwd()
	// Resolve symlinks just in case
	cwd, _ = filepath.EvalSymlinks(cwd)

	// Get git root directory to make paths relative
	gitRootOut, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git root: %w", err)
	}
	gitRoot := strings.TrimSpace(string(gitRootOut))

	// Parse worktree info and extract names
	type lineInfo struct {
		absPath string
		name    string
		commit  string
		branch  string
		isCwd   bool
	}
	lineInfos := make([]lineInfo, 0, len(lines))

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			absPath := fields[0]
			relPath, err := filepath.Rel(gitRoot, absPath)
			if err != nil {
				return nil, fmt.Errorf("failed to get relative path for %s: %w", absPath, err)
			}
			// Worktree name is the last component of the path
			name := filepath.Base(absPath)
			// Check if this is the current worktree
			isCwd := relPath == "." || relPath == "./"
			// Extract commit hash (second field)
			commit := ""
			if len(fields) > 1 {
				commit = fields[1]
			}
			// Extract branch/ref info from brackets or parentheses
			branch := ""
			if idx := strings.Index(line, "["); idx != -1 {
				end := strings.Index(line[idx:], "]")
				if end != -1 {
					branch = line[idx+1 : idx+end]
				}
			} else if idx := strings.Index(line, "("); idx != -1 {
				end := strings.Index(line[idx:], ")")
				if end != -1 {
					branch = line[idx+1 : idx+end]
				}
			}
			lineInfos = append(lineInfos, lineInfo{absPath: absPath, name: name, commit: commit, branch: branch, isCwd: isCwd})
		}
	}

	// Build result slice
	result := make([]WorktreeInfo, 0, len(lineInfos))
	for _, info := range lineInfos {
		// Build display string with name, commit, and branch/ref info
		var display string
		if info.branch != "" {
			display = fmt.Sprintf("%-20s %s [%s]", info.name, info.commit, info.branch)
		} else {
			display = fmt.Sprintf("%-20s %s", info.name, info.commit)
		}
		result = append(result, WorktreeInfo{
			AbsPath: info.absPath,
			Display: display,
			Commit:  info.commit,
			Branch:  info.branch,
			IsCwd:   info.isCwd,
		})
	}

	// Sort: Current dir first, then alphabetical by name
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].IsCwd {
			return true
		}
		if result[j].IsCwd {
			return false
		}
		return result[i].Display < result[j].Display
	})

	return result, nil
}

func findWorktreePathForBranch(branch string) (string, error) {
	wts, _ := getWorktreesSorted()
	for _, wt := range wts {
		if wt.Branch == branch {
			return wt.AbsPath, nil
		}
	}
	return "", fmt.Errorf("branch not found")
}

func printPath(p string) { fmt.Println(p) }
func fail(err error)     { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
