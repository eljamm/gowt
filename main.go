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
	"github.com/ktr0731/go-fuzzyfinder/matching"
	"github.com/spf13/cobra"
)

// UI Colors
var (
	selectedBg     = tcell.NewRGBColor(73, 77, 100)
	selectedFg     = tcell.NewRGBColor(249, 226, 175)
	colorNormal    = tcell.ColorBlue
	colorInsert    = tcell.ColorGreen
	colorQuit      = tcell.ColorRed
	filterBg       = tcell.ColorBlack
	filterFg       = tcell.ColorWhite
	modeFg         = tcell.ColorBlack
	highlightSelFg = tcell.NewRGBColor(107, 200, 122)
	highlightFg    = tcell.NewRGBColor(166, 227, 161)
	arrowFg        = tcell.ColorRed
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
	NavDelta      int  // -1 for up, +1 for down, 0 for none
	NavTop        bool // g without count
	NavBottom     bool // G without count
	NavTarget     int  // 0 = none, 1-based line number
	NavPage       int  // -1 for ctrl-u (page up), +1 for ctrl-d (page down)
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
		return s, ActionDraw, RenderRequest{Query: s.query, ModeIndicator: " I "}
	case tcell.KeyDown, tcell.KeyCtrlN, tcell.KeyCtrlJ:
		return s, ActionDraw, RenderRequest{Query: s.query, ModeIndicator: " I "}
	case tcell.KeyCtrlD:
		// Page down (half window)
		return s, ActionDraw, RenderRequest{Query: s.query, ModeIndicator: " I ", NavPage: 1}
	case tcell.KeyCtrlU:
		// Page up (half window)
		return s, ActionDraw, RenderRequest{Query: s.query, ModeIndicator: " I ", NavPage: -1}
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

type NormalState struct {
	count int // accumulated count prefix (0 if none)
}

func (s NormalState) HandleKey(e KeyEvent) (State, Action, RenderRequest) {
	switch e.Key {
	case tcell.KeyUp, tcell.KeyCtrlP, tcell.KeyCtrlK:
		return s, ActionDraw, RenderRequest{ModeIndicator: " N "}
	case tcell.KeyDown, tcell.KeyCtrlN, tcell.KeyCtrlJ:
		return s, ActionDraw, RenderRequest{ModeIndicator: " N "}
	case tcell.KeyCtrlD:
		delta := s.count
		if delta == 0 {
			delta = 1
		}
		s.count = 0
		return s, ActionDraw, RenderRequest{ModeIndicator: " N ", NavPage: delta}
	case tcell.KeyCtrlU:
		delta := s.count
		if delta == 0 {
			delta = 1
		}
		s.count = 0
		return s, ActionDraw, RenderRequest{ModeIndicator: " N ", NavPage: -delta}
	case tcell.KeyEsc:
		s.count = 0
		return ConfirmQuitState{}, ActionDraw, RenderRequest{ModeIndicator: " ? "}
	case tcell.KeyEnter:
		s.count = 0
		return s, ActionSelect, RenderRequest{ModeIndicator: " N "}
	case tcell.KeyCtrlC:
		s.count = 0
		return s, ActionQuit, RenderRequest{}
	}
	switch e.Rune {
	case 'j', 'J':
		delta := s.count
		if delta == 0 {
			delta = 1
		}
		s.count = 0
		return s, ActionDraw, RenderRequest{ModeIndicator: " N ", NavDelta: delta}
	case 'k', 'K':
		delta := s.count
		if delta == 0 {
			delta = -1
		} else {
			delta = -delta // count before k flips direction (like vim)
		}
		s.count = 0
		return s, ActionDraw, RenderRequest{ModeIndicator: " N ", NavDelta: delta}
	case 'g':
		if s.count > 0 {
			idx := s.count - 1 // 1-based to 0-based
			s.count = 0
			return s, ActionDraw, RenderRequest{ModeIndicator: " N ", NavTarget: idx}
		}
		s.count = 0
		return s, ActionDraw, RenderRequest{ModeIndicator: " N ", NavTop: true}
	case 'G':
		if s.count > 0 {
			idx := s.count - 1 // 1-based to 0-based
			s.count = 0
			return s, ActionDraw, RenderRequest{ModeIndicator: " N ", NavTarget: idx}
		}
		s.count = 0
		return s, ActionDraw, RenderRequest{ModeIndicator: " N ", NavBottom: true}
	case 'i':
		s.count = 0
		return InsertState{query: ""}, ActionDraw, RenderRequest{ModeIndicator: " I ", Query: ""}
	case 'q', 'Q':
		s.count = 0
		return ConfirmQuitState{}, ActionDraw, RenderRequest{ModeIndicator: " ? "}
	case 'y', 'Y':
		// Only in confirm mode, not here
	case 'n', 'N':
		// Only in confirm mode
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		s.count = s.count*10 + int(e.Rune-'0')
		return s, ActionDraw, RenderRequest{ModeIndicator: " N "}
	}
	s.count = 0
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
	rootCmd := &cobra.Command{
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
	if idx < 0 {
		return
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
	if idx < 0 {
		return
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
	selected := -1
	windowTop := 0
	if gitDir, err := defaultCommander.gitCommonDir(); err == nil && gitDir != "" {
		gitRoot := strings.TrimSuffix(gitDir, "/.git")
		for i, wt := range worktrees {
			if wt.AbsPath == gitRoot || strings.HasPrefix(wt.AbsPath, gitRoot+"/") {
				selected = i
				_, screenHeight := screen.Size()
				listHeight := screenHeight - 1
				windowRatio := 0.75
				windowHeight := int(float64(listHeight) * windowRatio)
				if windowHeight > 1 && selected >= windowHeight-1 {
					windowTop = selected - windowHeight + 1
				}
				break
			}
		}
	}

	numDigits := len(fmt.Sprintf("%d", len(worktrees)))
	numPad := fmt.Sprintf("%%%dd", numDigits)

	displayWorktrees := func(query string) []matchedWorktree {
		if query == "" {
			result := make([]matchedWorktree, len(worktrees))
			for i, wt := range worktrees {
				result[i] = matchedWorktree{WorktreeInfo: wt}
			}
			return result
		}

		displayStrs := make([]string, len(worktrees))
		for i, wt := range worktrees {
			displayStrs[i] = wt.Display
		}

		matched := matching.FindAll(query, displayStrs, matching.WithMode(matching.ModeSmart))
		if len(matched) == 0 {
			return []matchedWorktree{}
		}

		result := make([]matchedWorktree, 0, len(matched))
		for i := len(matched) - 1; i >= 0; i-- {
			m := matched[i]
			wt := worktrees[m.Idx]
			positions := fuzzyMatchPositions(query, wt.Display)
			result = append(result, matchedWorktree{
				WorktreeInfo:   wt,
				MatchPositions: positions,
			})
		}
		return result
	}

	draw := func(req RenderRequest) {
		screen.Clear()
		width, height := screen.Size()

		visible := displayWorktrees(req.Query)

		selectedIdx := selected
		if selectedIdx < 0 {
			selectedIdx = len(visible) - 1
		} else if selectedIdx >= len(visible) {
			selectedIdx = 0
		}

		listHeight := height - 1
		windowRatio := 0.75
		windowHeight := int(float64(listHeight) * windowRatio)
		if windowHeight < 1 {
			windowHeight = 1
		}
		if windowHeight < len(visible) && selectedIdx >= windowTop+windowHeight {
			windowTop = selectedIdx - windowHeight + 1
		}
		if windowTop < 0 {
			windowTop = 0
		}
		if windowTop+windowHeight > len(visible) {
			windowTop = len(visible) - windowHeight
		}
		if windowTop < 0 {
			windowTop = 0
		}

		arrowStyle := tcell.StyleDefault.Foreground(arrowFg)
		numberStyle := tcell.StyleDefault.Foreground(tcell.ColorDarkGray)

		for i := windowTop; i < len(visible); i++ {
			wt := visible[i]
			totalShow := windowHeight
			if totalShow > len(visible) {
				totalShow = len(visible)
			}
			startRow := listHeight - totalShow
			row := startRow + (i - windowTop)
			if row >= height-1 || row < 0 {
				continue
			}
			if i == selectedIdx {
				screen.SetContent(0, row, '>', nil, arrowStyle)
			}
			screen.SetContent(1, row, ' ', nil, tcell.StyleDefault)

			numStr := fmt.Sprintf(numPad, i+1)
			for x, r := range numStr {
				if x+2 >= width {
					break
				}
				screen.SetContent(x+2, row, r, nil, numberStyle)
			}
			numberEnd := 2 + numDigits

			style := tcell.StyleDefault
			if i == selectedIdx {
				style = style.Background(selectedBg).Foreground(selectedFg).Bold(true)
			}
			for x, r := range wt.Display {
				if x+numberEnd+1 >= width {
					break
				}
				charStyle := style
				for _, pos := range wt.MatchPositions {
					if x >= pos[0] && x < pos[1] {
						if i == selectedIdx {
							charStyle = style.Foreground(highlightSelFg).Bold(true)
						} else {
							charStyle = style.Foreground(highlightFg)
						}
						break
					}
				}
				screen.SetContent(x+numberEnd+1, row, r, nil, charStyle)
			}
		}

		modeStyle := tcell.StyleDefault.Foreground(modeFg)
		switch req.ModeIndicator {
		case " N ":
			modeStyle = modeStyle.Background(colorNormal)
		case " I ":
			modeStyle = modeStyle.Background(colorInsert)
		case " ? ":
			modeStyle = modeStyle.Background(colorQuit)
		}

		leftWidth := 2 + numDigits
		offset := (leftWidth + 1) % 2
		modeRunes := []rune(req.ModeIndicator)
		modeStyle = modeStyle.Bold(true)
		for col := 0; col < leftWidth; col++ {
			idx := col - offset
			if idx < 0 || idx >= len(modeRunes) {
				continue
			}
			screen.SetContent(col, height-1, modeRunes[idx], nil, modeStyle)
		}

		if req.ModeIndicator == " I " {
			for x, r := range req.Query {
				if x+numDigits+3 >= width {
					break
				}
				screen.SetContent(
					x+numDigits+3,
					height-1,
					r,
					nil,
					tcell.StyleDefault.Background(filterBg).Foreground(filterFg),
				)
			}
		} else if req.ModeIndicator == " ? " {
			prompt := "[Y/Enter] yes / [N/ESC] no"
			for x, r := range prompt {
				if x+numDigits+3 >= width {
					break
				}
				screen.SetContent(x+numDigits+3, height-1, r, nil, tcell.StyleDefault.Foreground(colorQuit))
			}
		}

		screen.Show()
	}

	// Initial draw
	var action Action
	var req RenderRequest
	state, action, req = state.HandleKey(KeyEvent{})
	currentQuery := ""
	draw(req)
	if selected < 0 {
		selected = len(worktrees) - 1
	}

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
			visible := displayWorktrees(currentQuery)
			if selected < len(visible)-1 {
				selected++
			}
		}

		_, screenHeight := screen.Size()
		listHeight := screenHeight - 1
		windowRatio := 0.75
		windowHeight := int(float64(listHeight) * windowRatio)
		if windowHeight < 1 {
			windowHeight = 1
		}
		if selected < windowTop {
			windowTop = selected
		} else if selected >= windowTop+windowHeight {
			windowTop = selected - windowHeight + 1
		}

		var nextState State
		nextState, action, req = state.HandleKey(keyEvent)

		// Handle navigation from state machine (j/k/g/G in Normal mode, Ctrl-d/Ctrl-u)
		if req.NavDelta != 0 || req.NavTop || req.NavBottom || req.NavTarget > 0 || req.NavPage != 0 {
			visible := displayWorktrees(currentQuery)
			if req.NavTop {
				selected = 0
			} else if req.NavBottom {
				if len(visible) > 0 {
					selected = len(visible) - 1
				} else {
					selected = 0
				}
			} else if req.NavTarget > 0 {
				targetIdx := req.NavTarget
				if targetIdx < 0 {
					targetIdx = 0
				}
				if targetIdx >= len(visible) {
					targetIdx = len(visible) - 1
				}
				selected = targetIdx
			} else if req.NavPage != 0 {
				_, screenHeight := screen.Size()
				listHeight := screenHeight - 1
				windowRatio := 0.75
				windowHeight := int(float64(listHeight) * windowRatio)
				if windowHeight < 1 {
					windowHeight = 1
				}
				halfPage := windowHeight / 2
				if halfPage < 1 {
					halfPage = 1
				}
				pageDelta := halfPage * req.NavPage
				newSelected := selected + pageDelta
				if newSelected >= len(visible) {
					selected = len(visible) - 1
				} else if newSelected < 0 {
					selected = 0
				} else {
					selected = newSelected
				}
			} else {
				newSelected := selected + req.NavDelta
				if req.NavDelta > 0 {
					if newSelected >= len(visible) {
						selected = len(visible) - 1
					} else {
						selected = newSelected
					}
				} else {
					if newSelected < 0 {
						selected = 0
					} else {
						selected = newSelected
					}
				}
			}

			_, screenHeight := screen.Size()
			listHeight := screenHeight - 1
			windowRatio := 0.75
			windowHeight := int(float64(listHeight) * windowRatio)
			if windowHeight < 1 {
				windowHeight = 1
			}
			if selected < windowTop {
				windowTop = selected
			} else if selected >= windowTop+windowHeight {
				windowTop = selected - windowHeight + 1
			}
		}

		// Handle query changes: track the most relevant match or preserve selection on clear
		if currentQuery != req.Query {
			oldVisible := displayWorktrees(currentQuery)
			newVisible := displayWorktrees(req.Query)

			var selectedPath string
			if len(oldVisible) > 0 {
				clamped := selected
				if clamped >= len(oldVisible) {
					clamped = len(oldVisible) - 1
				}
				if clamped < 0 {
					clamped = 0
				}
				selectedPath = oldVisible[clamped].AbsPath
			}

			if len(newVisible) > 0 {
				if req.Query == "" {
					selected = 0
					for i, wt := range newVisible {
						if wt.AbsPath == selectedPath {
							selected = i
							break
						}
					}
				} else {
					selected = len(newVisible) - 1
				}
			} else {
				selected = 0
			}
		}

		// Extract query from RenderRequest to use for navigation bounds
		currentQuery = req.Query
		state = nextState
		draw(req)

		if action == ActionQuit {
			return -1, nil
		}
		if action == ActionSelect {
			visible := displayWorktrees(currentQuery)
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
	gitCommonDir() (string, error)
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

func (r *realGitCommander) gitCommonDir() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--git-common-dir").Output()
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

type matchedWorktree struct {
	WorktreeInfo
	MatchPositions [][2]int
}

func fuzzyMatchPositions(query, target string) [][2]int {
	if query == "" {
		return nil
	}
	queryLower := strings.ToLower(query)
	targetLower := strings.ToLower(target)
	var positions [][2]int
	targetRunes := []rune(targetLower)
	targetIdx := 0
	for _, q := range queryLower {
		found := false
		for ; targetIdx < len(targetRunes); targetIdx++ {
			if targetRunes[targetIdx] == q {
				positions = append(positions, [2]int{targetIdx, targetIdx + 1})
				targetIdx++
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	return positions
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

func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
