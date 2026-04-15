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

	"github.com/gdamore/tcell/v2"
	"github.com/spf13/cobra"
)

// UI Colors
var (
	selectedBg  = tcell.ColorWhite
	selectedFg  = tcell.ColorBlack
	colorNormal = tcell.ColorBlue
	colorInsert = tcell.ColorGreen
	colorQuit   = tcell.ColorRed
	filterBg    = tcell.ColorBlack
	filterFg    = tcell.ColorWhite
	modeFg      = tcell.ColorBlack
)

// State Machine Types

type Action int

const (
	ActionNone Action = iota
	ActionDraw
	ActionQuit
	ActionSelect
)

type KeyEvent struct {
	Key  tcell.Key
	Rune rune
}

type RenderRequest struct {
	ModeIndicator string
	ModeStyle     tcell.Style
	Query         string
	Items         []WorktreeInfo
	Selected      int
	StatusPrompt  string
}

type State interface {
	HandleKey(e KeyEvent) (State, Action, RenderRequest)
}

// InsertState

type InsertState struct {
	query string
}

func (s InsertState) HandleKey(e KeyEvent) (State, Action, RenderRequest) {
	switch e.Key {
	case tcell.KeyUp, tcell.KeyCtrlP, tcell.KeyCtrlK:
		// Navigation - handled in main loop via returning updated selection
		return s, ActionDraw, RenderRequest{Query: s.query, ModeIndicator: " I "}
	case tcell.KeyDown, tcell.KeyCtrlN, tcell.KeyCtrlJ:
		return s, ActionDraw, RenderRequest{Query: s.query, ModeIndicator: " I "}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(s.query) > 0 {
			s.query = s.query[:len(s.query)-1]
		}
		return s, ActionDraw, RenderRequest{Query: s.query, ModeIndicator: " I "}
	case tcell.KeyEsc:
		if len(s.query) > 0 {
			s.query = ""
			return s, ActionDraw, RenderRequest{Query: s.query, ModeIndicator: " I "}
		}
		return NormalState{}, ActionDraw, RenderRequest{ModeIndicator: " N "}
	case tcell.KeyEnter:
		return s, ActionSelect, RenderRequest{Query: s.query, ModeIndicator: " I "}
	case tcell.KeyCtrlC:
		return s, ActionQuit, RenderRequest{}
	default:
		// All printable characters -> type in query
		if e.Rune >= 32 && e.Rune < 127 {
			s.query += string(e.Rune)
		}
		return s, ActionDraw, RenderRequest{Query: s.query, ModeIndicator: " I "}
	}
}

// NormalState

type NormalState struct{}

func (s NormalState) HandleKey(e KeyEvent) (State, Action, RenderRequest) {
	switch e.Key {
	case tcell.KeyUp, tcell.KeyCtrlP, tcell.KeyCtrlK:
		return s, ActionDraw, RenderRequest{ModeIndicator: " N "}
	case tcell.KeyDown, tcell.KeyCtrlN, tcell.KeyCtrlJ:
		return s, ActionDraw, RenderRequest{ModeIndicator: " N "}
	case tcell.KeyEsc:
		return ConfirmQuitState{}, ActionDraw, RenderRequest{ModeIndicator: " ? "}
	case tcell.KeyEnter:
		return s, ActionSelect, RenderRequest{ModeIndicator: " N "}
	case tcell.KeyCtrlC:
		return s, ActionQuit, RenderRequest{}
	}
	switch e.Rune {
	case 'j':
		return s, ActionDraw, RenderRequest{ModeIndicator: " N "}
	case 'k':
		return s, ActionDraw, RenderRequest{ModeIndicator: " N "}
	case 'i':
		return InsertState{query: ""}, ActionDraw, RenderRequest{ModeIndicator: " I ", Query: ""}
	case 'q', 'Q':
		return ConfirmQuitState{}, ActionDraw, RenderRequest{ModeIndicator: " ? "}
	case 'y', 'Y':
		// Only in confirm mode, not here
	case 'n', 'N':
		// Only in confirm mode
	}
	return s, ActionDraw, RenderRequest{ModeIndicator: " N "}
}

// ConfirmQuitState

type ConfirmQuitState struct{}

func (s ConfirmQuitState) HandleKey(e KeyEvent) (State, Action, RenderRequest) {
	switch e.Key {
	case tcell.KeyEsc:
		return NormalState{}, ActionDraw, RenderRequest{ModeIndicator: " N "}
	case tcell.KeyEnter:
		return s, ActionQuit, RenderRequest{ModeIndicator: " ? "}
	case tcell.KeyCtrlC:
		return s, ActionQuit, RenderRequest{}
	}
	switch e.Rune {
	case 'y', 'Y':
		return s, ActionQuit, RenderRequest{ModeIndicator: " ? "}
	case 'n', 'N':
		return NormalState{}, ActionDraw, RenderRequest{ModeIndicator: " N "}
	}
	return s, ActionDraw, RenderRequest{ModeIndicator: " ? "}
}

// TODO: Replace color vars with config-based customization
// - Add UIConfig struct with color fields
// - Support loading from env vars or config file
// - Allow runtime customization via flags

// TODO: Make ESC behavior configurable
// - Option A: ESC immediately quits (fzf-like)
// - Option B: Current confirmation prompt
// - Option C: Different key to quit (e.g., Ctrl+C)

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
	return selectWorktreeTUI(worktrees)
}

func selectWorktreeTUI(worktrees []WorktreeInfo) (int, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return -1, err
	}
	defer func() {
		if screen != nil {
			screen.Fini()
		}
	}()

	if err := screen.Init(); err != nil {
		screen = nil
		return -1, err
	}

	var state State = InsertState{query: ""}
	selected := 0

	displayWorktrees := func(query string) []WorktreeInfo {
		if query == "" {
			return worktrees
		}
		lowerQuery := strings.ToLower(query)
		var filtered []WorktreeInfo
		for _, wt := range worktrees {
			if strings.Contains(strings.ToLower(wt.Display), lowerQuery) {
				filtered = append(filtered, wt)
			}
		}
		if len(filtered) == 0 {
			return worktrees
		}
		return filtered
	}

	draw := func(req RenderRequest) {
		screen.Clear()
		width, height := screen.Size()

		visible := displayWorktrees(req.Query)

		selectedIdx := selected
		if selectedIdx >= len(visible) {
			selectedIdx = len(visible) - 1
		}
		if selectedIdx < 0 {
			selectedIdx = 0
		}

		listHeight := height - 1
		startRow := listHeight - len(visible)
		if startRow < 0 {
			startRow = 0
		}

		for i, wt := range visible {
			row := startRow + i
			if row >= height-1 {
				break
			}
			style := tcell.StyleDefault
			if i == selectedIdx {
				style = style.Background(selectedBg).Foreground(selectedFg)
			}
			for x, r := range wt.Display {
				if x+3 >= width {
					break
				}
				screen.SetContent(x+3, row, r, nil, style)
			}
		}

		modeStyle := tcell.StyleDefault.Foreground(modeFg)
		switch req.ModeIndicator {
		case " N ":
			modeStyle = modeStyle.Background(colorNormal)
		case " I ":
			modeStyle = modeStyle.Background(colorInsert)
		case " ? ":
			modeStyle = modeStyle.Background(colorQuit).Foreground(colorQuit)
		}

		for x, r := range req.ModeIndicator {
			screen.SetContent(x, height-1, r, nil, modeStyle)
		}

		if req.ModeIndicator == " I " {
			for x, r := range req.Query {
				if x+3 >= width {
					break
				}
				screen.SetContent(x+3, height-1, r, nil, tcell.StyleDefault.Background(filterBg).Foreground(filterFg))
			}
		} else if req.ModeIndicator == " ? " {
			prompt := "[Y/Enter] yes / [N/ESC] no"
			for x, r := range prompt {
				if x+3 >= width {
					break
				}
				screen.SetContent(x+3, height-1, r, nil, tcell.StyleDefault.Foreground(colorQuit))
			}
		}

		screen.Show()
	}

	// Initial draw
	var action Action
	var req RenderRequest
	state, action, req = state.HandleKey(KeyEvent{})
	draw(req)

	for {
		ev := screen.PollEvent()

		var keyEvent KeyEvent
		if ev, ok := ev.(*tcell.EventKey); ok {
			keyEvent = KeyEvent{Key: ev.Key(), Rune: ev.Rune()}
		}

		// Handle navigation in main loop (need to update selected)
		switch keyEvent.Key {
		case tcell.KeyUp, tcell.KeyCtrlP, tcell.KeyCtrlK:
			if selected > 0 {
				selected--
			}
		case tcell.KeyDown, tcell.KeyCtrlN, tcell.KeyCtrlJ:
			visible := displayWorktrees("")
			if selected < len(visible)-1 {
				selected++
			}
		}

		// Also handle j/k in normal mode here (states return action for these)
		if keyEvent.Rune == 'j' || keyEvent.Rune == 'k' {
			visible := displayWorktrees("")
			if keyEvent.Rune == 'j' && selected < len(visible)-1 {
				selected++
			}
			if keyEvent.Rune == 'k' && selected > 0 {
				selected--
			}
		}

		var nextState State
		nextState, action, req = state.HandleKey(keyEvent)
		state = nextState
		draw(req)

		if action == ActionQuit {
			return -1, fmt.Errorf("cancelled")
		}
		if action == ActionSelect {
			visible := displayWorktrees("")
			selectedIdx := selected
			if selectedIdx >= len(visible) {
				selectedIdx = len(visible) - 1
			}
			if selectedIdx < 0 {
				selectedIdx = 0
			}
			if selectedIdx >= 0 && selectedIdx < len(visible) {
				for i, wt := range worktrees {
					if wt.AbsPath == visible[selectedIdx].AbsPath {
						return i, nil
					}
				}
			}
			return selected, nil
		}
	}
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
