package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "gwt",
		Short: "Git Worktree Manager",
		Run:   func(cmd *cobra.Command, args []string) { runJump(cmd, args) },
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "add [branch]",
		Short: "Create a worktree from a branch",
		Args:  cobra.ExactArgs(1),
		Run:   runAdd,
	})

	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Interactively remove a worktree",
		Run:   func(cmd *cobra.Command, args []string) { runRemove(cmd, args) },
	}
	removeCmd.Flags().BoolP("force", "f", false, "Force removal")
	rootCmd.AddCommand(removeCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// --- Handlers ---

func runJump(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		path, err := findWorktreePathForBranch(args[0], defaultCommander)
		if err != nil {
			fail(err)
		}
		printPath(path)
		return
	}

	worktrees, err := getWorktreesSorted(defaultCommander)
	if err != nil {
		fail(err)
	}

	idx, err := selectWorktree(worktrees)
	if err != nil {
		fail(err)
	}

	printPath(worktrees[idx].AbsPath)
}

func runAdd(cmd *cobra.Command, args []string) {
	branch := args[0]
	cwd, err := os.Getwd()
	if err != nil {
		fail(fmt.Errorf("failed to get current directory: %w", err))
	}
	repoName := filepath.Base(cwd)
	sanitized := strings.ReplaceAll(branch, "/", "_")
	newPath := filepath.Join("..", repoName+"_"+sanitized)

	if err := defaultCommander.worktreeAdd(newPath, branch); err != nil {
		fmt.Fprintf(os.Stderr, "Branch not found. Create '%s'? [y/N] ", branch)
		res, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fail(fmt.Errorf("failed to read input: %w", err))
		}
		if strings.EqualFold(strings.TrimSpace(res), "y") {
			if err := defaultCommander.worktreeAddNew(newPath, branch); err != nil {
				fail(fmt.Errorf("failed to create worktree: %w", err))
			}
		} else {
			os.Exit(1)
		}
	}
	printPath(newPath)
}

func runRemove(cmd *cobra.Command, args []string) {
	worktrees, err := getWorktreesSorted(defaultCommander)
	if err != nil {
		fail(err)
	}

	idx, err := selectWorktree(worktrees)
	if err != nil {
		fail(err)
	}

	path := worktrees[idx].AbsPath
	force, _ := cmd.Flags().GetBool("force")
	if err := defaultCommander.worktreeRemove(path, force); err != nil {
		fail(err)
	}

	gitRoot, err := defaultCommander.revParse(true)
	if err != nil {
		fail(fmt.Errorf("failed to get git root: %w", err))
	}
	printPath(gitRoot)
}

// --- Helpers ---

func selectWorktree(worktrees []WorktreeInfo) (int, error) {
	return fuzzyfinder.Find(worktrees, func(i int) string {
		return worktrees[i].Display
	})
}

func extractBranch(line string) string {
	if idx := strings.Index(line, "["); idx != -1 {
		if end := strings.Index(line[idx:], "]"); end != -1 {
			return line[idx+1 : idx+end]
		}
	}
	if idx := strings.Index(line, "("); idx != -1 {
		if end := strings.Index(line[idx:], ")"); end != -1 {
			return line[idx+1 : idx+end]
		}
	}
	return ""
}

type GitCommander interface {
	worktreeList() (string, error)
	revParse(showToplevel bool) (string, error)
	worktreeRemove(path string, force bool) error
	worktreeAdd(path, branch string) error
	worktreeAddNew(path, branch string) error
}

type realGitCommander struct{}

func (r *realGitCommander) worktreeList() (string, error) {
	out, err := exec.Command("git", "worktree", "list").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (r *realGitCommander) revParse(showToplevel bool) (string, error) {
	args := []string{"rev-parse"}
	if showToplevel {
		args = append(args, "--show-toplevel")
	}
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (r *realGitCommander) worktreeRemove(path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force", path)
	} else {
		args = append(args, path)
	}
	c := exec.Command("git", args...)
	c.Stderr = os.Stderr
	return c.Run()
}

func (r *realGitCommander) worktreeAdd(path, branch string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "worktree", "add", path, branch)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (r *realGitCommander) worktreeAddNew(path, branch string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "worktree", "add", "-b", branch, path)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

var defaultCommander GitCommander = &realGitCommander{}

func setCommander(c GitCommander) {
	defaultCommander = c
}

type WorktreeInfo struct {
	AbsPath string
	Display string
	Commit  string
	Branch  string
	IsCwd   bool
}

func getWorktreesSorted(commander GitCommander) ([]WorktreeInfo, error) {
	out, err := commander.worktreeList()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return nil, fmt.Errorf("no worktrees found")
	}

	lines := strings.Split(trimmed, "\n")

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}
	cwd, err = filepath.EvalSymlinks(cwd)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	gitRoot, err := commander.revParse(true)
	if err != nil {
		return nil, fmt.Errorf("failed to get git root: %w", err)
	}

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
			name := filepath.Base(absPath)
			isCwd := relPath == "." || relPath == "./"
			commit := ""
			if len(fields) > 1 {
				commit = fields[1]
			}
			branch := extractBranch(line)
			lineInfos = append(lineInfos, lineInfo{absPath: absPath, name: name, commit: commit, branch: branch, isCwd: isCwd})
		}
	}

	result := make([]WorktreeInfo, 0, len(lineInfos))
	for _, info := range lineInfos {
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

func findWorktreePathForBranch(branch string, commander GitCommander) (string, error) {
	wts, err := getWorktreesSorted(commander)
	if err != nil {
		return "", fmt.Errorf("failed to get worktrees: %w", err)
	}
	for _, wt := range wts {
		if wt.Branch == branch {
			return wt.AbsPath, nil
		}
	}
	return "", fmt.Errorf("branch not found")
}

func printPath(p string) { fmt.Println(p) }
func fail(err error)     { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
