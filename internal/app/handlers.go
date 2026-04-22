package app

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ktr0731/go-fuzzyfinder/matching"
	"github.com/spf13/cobra"

	"gwt/internal/git"
	"gwt/internal/tui"
)

func GetWorktreesSorted(commander git.Commander) ([]tui.WorktreeInfo, error) {
	out, err := commander.WorktreeList()
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

	gitRoot, err := commander.RevParse(true)
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
			branch := ExtractBranch(line)
			lineInfos = append(
				lineInfos,
				lineInfo{
					absPath: absPath,
					name:    name,
					commit:  commit,
					branch:  branch,
					isCwd:   isCwd,
				},
			)
		}
	}

	result := make([]tui.WorktreeInfo, 0, len(lineInfos))
	for _, info := range lineInfos {
		var display string
		if info.branch != "" {
			display = fmt.Sprintf("%-20s %s [%s]", info.name, info.commit, info.branch)
		} else {
			display = fmt.Sprintf("%-20s %s", info.name, info.commit)
		}
		result = append(result, tui.WorktreeInfo{
			Name:    info.name,
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

func FindWorktreePathForBranch(branch string, commander git.Commander) (string, error) {
	wts, err := GetWorktreesSorted(commander)
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

func ExtractBranch(line string) string {
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

var SelectWorktreeTUI func(worktrees []tui.WorktreeInfo, commander git.Commander, colors tui.TUIColors) (int, error)

var TUIColors tui.TUIColors

func RunHome(cmd *cobra.Command, args []string, commander git.Commander) {
	gitCommonDir, err := commander.GitCommonDir()
	if err != nil {
		Fail(fmt.Errorf("failed to get git common dir: %w", err))
	}

	cfg, err := LoadConfig(gitCommonDir)
	if err != nil {
		Fail(fmt.Errorf("failed to load config: %w", err))
	}

	if cfg.Main == "" {
		Fail(fmt.Errorf("main worktree path not configured"))
	}
	PrintPath(cfg.Main)
}

func RunRoot(cmd *cobra.Command, args []string, commander git.Commander) {
	gitCommonDir, err := commander.GitCommonDir()
	if err != nil {
		Fail(fmt.Errorf("failed to get git common dir: %w", err))
	}
	root := strings.TrimSuffix(gitCommonDir, "/.git")
	if root == gitCommonDir {
		Fail(fmt.Errorf("failed to determine git root from common dir"))
	}
	PrintPath(root)
}

func RunConfig(cmd *cobra.Command, args []string, commander git.Commander) {
	gitCommonDir, err := commander.GitCommonDir()
	if err != nil {
		Fail(fmt.Errorf("failed to get git common dir: %w", err))
	}

	cfg, err := LoadConfig(gitCommonDir)
	if err != nil {
		Fail(fmt.Errorf("failed to load config: %w", err))
	}

	if len(args) > 0 {
		absPath, err := filepath.Abs(args[0])
		if err != nil {
			Fail(fmt.Errorf("failed to resolve absolute path: %w", err))
		}
		cfg.Main = absPath
		if err := SaveConfig(gitCommonDir, cfg); err != nil {
			Fail(fmt.Errorf("failed to save config: %w", err))
		}
		PrintPath(cfg.Main)
		return
	}

	if cfg.Main == "" {
		Fail(fmt.Errorf("main worktree path not configured"))
	}
	PrintPath(cfg.Main)
}

func FindWorktreePath(name string, commander git.Commander) (string, error) {
	wts, err := GetWorktreesSorted(commander)
	if err != nil {
		return "", err
	}
	for _, wt := range wts {
		if wt.Name == name || wt.Branch == name {
			return wt.AbsPath, nil
		}
	}
	return "", fmt.Errorf("worktree '%s' not found", name)
}

func FindSimilarWorktrees(name string, worktrees []tui.WorktreeInfo) []tui.WorktreeInfo {
	displayStrs := make([]string, len(worktrees))
	for i, wt := range worktrees {
		displayStrs[i] = wt.Display
	}
	matched := matching.FindAll(name, displayStrs, matching.WithMode(matching.ModeSmart))
	if len(matched) == 0 {
		return nil
	}
	result := make([]tui.WorktreeInfo, 0, len(matched))
	for _, m := range matched {
		result = append(result, worktrees[m.Idx])
	}
	return result
}

func loadConfigForAdd(commander git.Commander) (Config, error) {
	gitCommonDir, err := commander.GitCommonDir()
	if err != nil {
		return Config{}, err
	}
	cfg, err := LoadConfig(gitCommonDir)
	if err != nil {
		return Config{}, err
	}
	if cfg.Main == "" {
		return Config{}, fmt.Errorf(
			"main worktree path not configured. Run 'gwt config main /path' first",
		)
	}
	return cfg, nil
}

func RunJump(cmd *cobra.Command, args []string, commander git.Commander) {
	wtName := ""
	if len(args) > 0 {
		wtName = args[0]
	}

	if wtName != "" {
		path, err := FindWorktreePath(wtName, commander)
		if err == nil {
			PrintPath(path)
			return
		}
		fmt.Fprintln(os.Stderr, err)
		reader := bufio.NewReader(os.Stdin)
		worktrees, listErr := GetWorktreesSorted(commander)
		if listErr != nil {
			Fail(listErr)
		}
		similar := FindSimilarWorktrees(wtName, worktrees)
		if len(similar) > 0 {
			fmt.Fprintln(os.Stderr, "Did you mean:")
			for i, s := range similar {
				fmt.Fprintf(os.Stderr, "  [%d] %s\n", i+1, s.Display)
			}
			fmt.Fprintf(os.Stderr, "  [Enter] Create new worktree '%s'\n", wtName)
			fmt.Fprintf(os.Stderr, "  [q] Quit\n")
			fmt.Fprint(os.Stderr, "Choose: ")
			res, err := reader.ReadString('\n')
			if err != nil {
				Fail(fmt.Errorf("failed to read input: %w", err))
			}
			res = strings.TrimSpace(res)
			if res == "" {
				fmt.Fprintf(os.Stderr, "Creating worktree '%s'...\n", wtName)
				cfg, cfgErr := loadConfigForAdd(commander)
				if cfgErr != nil {
					Fail(cfgErr)
				}
				newPath := filepath.Join(cfg.Main, wtName)
				if err := commander.WorktreeAdd(newPath, "HEAD"); err != nil {
					Fail(fmt.Errorf("failed to create worktree: %w", err))
				}
				PrintPath(newPath)
				return
			}
			if res == "q" || res == "Q" {
				os.Exit(0)
			}
			var chosenIdx int
			_, err = fmt.Sscanf(res, "%d", &chosenIdx)
			if err == nil && chosenIdx > 0 && chosenIdx <= len(similar) {
				PrintPath(similar[chosenIdx-1].AbsPath)
				return
			}
		}
		fmt.Fprintf(os.Stderr, "No similar worktrees found.\n")
		fmt.Fprintf(os.Stderr, "  [Enter] Create new worktree '%s'\n", wtName)
		fmt.Fprintf(os.Stderr, "  [q] Quit\n")
		fmt.Fprint(os.Stderr, "Choose: ")
		res, err := reader.ReadString('\n')
		if err != nil {
			Fail(fmt.Errorf("failed to read input: %w", err))
		}
		res = strings.TrimSpace(res)
		if res == "" {
			cfg, cfgErr := loadConfigForAdd(commander)
			if cfgErr != nil {
				Fail(cfgErr)
			}
			newPath := filepath.Join(cfg.Main, wtName)
			if err := commander.WorktreeAdd(newPath, "HEAD"); err != nil {
				Fail(fmt.Errorf("failed to create worktree: %w", err))
			}
			PrintPath(newPath)
			return
		}
		os.Exit(0)
	}

	worktrees, err := GetWorktreesSorted(commander)
	if err != nil {
		Fail(err)
	}

	idx, err := SelectWorktreeTUI(worktrees, commander, TUIColors)
	if err != nil {
		Fail(err)
	}
	if idx < 0 {
		return
	}

	PrintPath(worktrees[idx].AbsPath)
}

func RunAdd(cmd *cobra.Command, args []string, commander git.Commander) {
	gitCommonDir, err := commander.GitCommonDir()
	if err != nil {
		Fail(fmt.Errorf("failed to get git common dir: %w", err))
	}

	cfg, err := LoadConfig(gitCommonDir)
	if err != nil {
		Fail(fmt.Errorf("failed to load config: %w", err))
	}

	if cfg.Main == "" {
		fmt.Fprintf(os.Stderr, "Main worktree path not configured. Set now? [/path]: ")
		res, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			Fail(fmt.Errorf("failed to read input: %w", err))
		}
		res = strings.TrimSpace(res)
		if res == "" {
			Fail(fmt.Errorf("aborted: no path provided"))
		}
		absPath, err := filepath.Abs(res)
		if err != nil {
			Fail(fmt.Errorf("failed to resolve absolute path: %w", err))
		}
		cfg.Main = absPath
		if err := SaveConfig(gitCommonDir, cfg); err != nil {
			Fail(fmt.Errorf("failed to save config: %w", err))
		}
		fmt.Fprintf(os.Stderr, "Config saved.\n")
	}

	reader := bufio.NewReader(os.Stdin)

	var wtName, branch string

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Worktree name: ")
		wtName, err = reader.ReadString('\n')
		if err != nil {
			Fail(fmt.Errorf("failed to read input: %w", err))
		}
		wtName = strings.TrimSpace(wtName)
		if wtName == "" {
			Fail(fmt.Errorf("aborted: no worktree name"))
		}

		fmt.Fprintf(os.Stderr, "Branch name: ")
		branch, err = reader.ReadString('\n')
		if err != nil {
			Fail(fmt.Errorf("failed to read input: %w", err))
		}
		branch = strings.TrimSpace(branch)
	} else if len(args) == 1 {
		wtName = args[0]

		fmt.Fprintf(os.Stderr, "Branch name: ")
		branch, err = reader.ReadString('\n')
		if err != nil {
			Fail(fmt.Errorf("failed to read input: %w", err))
		}
		branch = strings.TrimSpace(branch)
	} else {
		wtName = args[0]
		branch = args[1]
	}

	newPath := filepath.Join(cfg.Main, wtName)

	if branch == "" {
		fmt.Fprintln(os.Stderr, "No branch specified, creating worktree at HEAD")
		if err := commander.WorktreeAdd(newPath, "HEAD"); err != nil {
			Fail(fmt.Errorf("failed to create worktree: %w", err))
		}
		PrintPath(newPath)
		return
	}

	if err := commander.WorktreeAdd(newPath, branch); err != nil {
		fmt.Fprintf(os.Stderr, "Branch not found. Create '%s'? [y/N] ", branch)
		res, err := reader.ReadString('\n')
		if err != nil {
			Fail(fmt.Errorf("failed to read input: %w", err))
		}
		if strings.EqualFold(strings.TrimSpace(res), "y") {
			if err := commander.WorktreeAddNew(newPath, branch); err != nil {
				Fail(fmt.Errorf("failed to create worktree: %w", err))
			}
		} else {
			os.Exit(1)
		}
	}
	PrintPath(newPath)
}

func RunRemove(cmd *cobra.Command, args []string, commander git.Commander) {
	worktrees, err := GetWorktreesSorted(commander)
	if err != nil {
		Fail(err)
	}

	var toRemove []tui.WorktreeInfo

	if len(args) > 0 {
		nameSet := make(map[string]bool)
		for _, arg := range args {
			nameSet[arg] = true
		}
		for _, wt := range worktrees {
			if nameSet[filepath.Base(wt.AbsPath)] {
				toRemove = append(toRemove, wt)
			}
		}
	} else {
		idx, err := SelectWorktreeTUI(worktrees, commander, TUIColors)
		if err != nil {
			Fail(err)
		}
		if idx < 0 {
			return
		}
		toRemove = append(toRemove, worktrees[idx])
	}

	if len(toRemove) == 0 {
		Fail(fmt.Errorf("no matching worktrees found"))
	}

	fmt.Fprintln(os.Stderr, "About to remove:")
	for _, wt := range toRemove {
		fmt.Fprintf(os.Stderr, "  - %s\n", wt.Display)
	}
	fmt.Fprintf(os.Stderr, "Remove %d worktree(s)? [y/N] ", len(toRemove))

	reader := bufio.NewReader(os.Stdin)
	res, err := reader.ReadString('\n')
	if err != nil {
		Fail(fmt.Errorf("failed to read input: %w", err))
	}
	if !strings.EqualFold(strings.TrimSpace(res), "y") {
		Fail(fmt.Errorf("aborted"))
	}

	force, _ := cmd.Flags().GetBool("force")
	gitRoot, err := commander.RevParse(true)
	if err != nil {
		Fail(fmt.Errorf("failed to get git root: %w", err))
	}

	var failed []string
	for _, wt := range toRemove {
		if err := commander.WorktreeRemove(wt.AbsPath, force); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to remove %s: %v\n", wt.Display, err)
			failed = append(failed, wt.Display)
			continue
		}
		fmt.Fprintf(os.Stderr, "Removed: %s\n", wt.Display)
	}

	if len(failed) > 0 {
		fmt.Fprintf(os.Stderr, "Failed to remove: %s\n", strings.Join(failed, ", "))
	}

	PrintPath(gitRoot)
}

func RunPurge(cmd *cobra.Command, args []string, commander git.Commander) {
	worktrees, err := GetWorktreesSorted(commander)
	if err != nil {
		Fail(err)
	}

	var toRemove []tui.WorktreeInfo

	if len(args) > 0 {
		nameSet := make(map[string]bool)
		for _, arg := range args {
			nameSet[arg] = true
		}
		for _, wt := range worktrees {
			if nameSet[filepath.Base(wt.AbsPath)] {
				toRemove = append(toRemove, wt)
			}
		}
	} else {
		idx, err := SelectWorktreeTUI(worktrees, commander, TUIColors)
		if err != nil {
			Fail(err)
		}
		if idx < 0 {
			return
		}
		toRemove = append(toRemove, worktrees[idx])
	}

	if len(toRemove) == 0 {
		Fail(fmt.Errorf("no matching worktrees found"))
	}

	branches := make(map[string]bool)
	for _, wt := range toRemove {
		if wt.Branch != "" {
			branches[wt.Branch] = true
		}
	}

	fmt.Fprintln(os.Stderr, "About to remove:")
	for _, wt := range toRemove {
		fmt.Fprintf(os.Stderr, "  - %s\n", wt.Display)
	}
	if len(branches) > 0 {
		fmt.Fprintln(os.Stderr, "About to delete branches:")
		for branch := range branches {
			fmt.Fprintf(os.Stderr, "  - %s\n", branch)
		}
	}
	fmt.Fprintf(
		os.Stderr,
		"Remove %d worktree(s) and delete %d branch(es)? [y/N] ",
		len(toRemove),
		len(branches),
	)

	reader := bufio.NewReader(os.Stdin)
	res, err := reader.ReadString('\n')
	if err != nil {
		Fail(fmt.Errorf("failed to read input: %w", err))
	}
	if !strings.EqualFold(strings.TrimSpace(res), "y") {
		Fail(fmt.Errorf("aborted"))
	}

	force, _ := cmd.Flags().GetBool("force")
	gitRoot, err := commander.RevParse(true)
	if err != nil {
		Fail(fmt.Errorf("failed to get git root: %w", err))
	}

	var failed []string
	for _, wt := range toRemove {
		if err := commander.WorktreeRemove(wt.AbsPath, force); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to remove %s: %v\n", wt.Display, err)
			failed = append(failed, wt.Display)
			continue
		}
		fmt.Fprintf(os.Stderr, "Removed: %s\n", wt.Display)

		if wt.Branch != "" {
			if err := commander.BranchDelete(wt.Branch, force); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to delete branch %s: %v\n", wt.Branch, err)
				failed = append(failed, wt.Branch)
				continue
			}
			fmt.Fprintf(os.Stderr, "Deleted branch: %s\n", wt.Branch)
		}
	}

	if len(failed) > 0 {
		fmt.Fprintf(os.Stderr, "Failed: %s\n", strings.Join(failed, ", "))
	}

	PrintPath(gitRoot)
}

func PrintPath(p string) { fmt.Println(p) }

func Fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
