# Architecture — State machine, GitCommander, TUI rendering

## Overview

The TUI is built from scratch using `tcell` for rendering and a state machine pattern for input handling. The binary only prints paths; shell wrappers (bash/zsh/fish) auto-`cd` into the selected worktree.

## State Machine

Three modes: Insert (default), Normal (vim-style), ConfirmQuit.

### InsertState (default)
- Type to fuzzy-filter worktrees
- Enter to select
- ESC clears query or switches to Normal mode
- Ctrl+C to quit

### NormalState
- `j/k` or arrow keys: move selection
- `g/G`: jump to top/bottom
- Count prefixes: `3j` moves down 3 lines
- `i`: switch to Insert mode
- `q` or ESC: quit confirmation
- Ctrl+D/Ctrl+U: page down/up (with count prefix)

### ConfirmQuitState
- `y` or Enter: confirm quit
- `n` or ESC: cancel and return to Normal

```go
type State interface {
    HandleKey(e KeyEvent) (State, Action, RenderRequest)
}

type Action int
const (
    ActionNone Action = iota
    ActionDraw
    ActionQuit
    ActionSelect
)
```

## GitCommander Interface

All git operations go through `GitCommander` interface for testability.

```go
type GitCommander interface {
    worktreeList() (string, error)
    revParse(showToplevel bool) (string, error)
    gitCommonDir() (string, error)
    worktreeRemove(path string, force bool) error
    worktreeAdd(path, branch string) error
    worktreeAddNew(path, branch string) error
}

var defaultCommander GitCommander = &realGitCommander{}
```

Use `setCommander(mock)` in tests to inject mock behavior.

## Data Types

```go
type WorktreeInfo struct {
    AbsPath string // absolute path for git commands and cd
    Display string // worktree name for fuzzy finder UI
    Commit  string // commit hash
    Branch  string // branch name (if any)
    IsCwd   bool   // true if this is the current working directory
}

type matchedWorktree struct {
    WorktreeInfo
    MatchPositions [][2]int
}
```

## Fuzzy Matching

Uses `go-fuzzyfinder/matching` for character matching:

```go
displayStrs := make([]string, len(worktrees))
for i, wt := range worktrees {
    displayStrs[i] = wt.Display
}

matched := matching.FindAll(query, displayStrs, matching.WithMode(matching.ModeSmart))
```

## UI Colors

```go
var (
    selectedBg     = tcell.NewRGBColor(73, 77, 100)
    selectedFg     = tcell.NewRGBColor(249, 226, 175)
    colorNormal    = tcell.ColorBlue
    colorInsert    = tcell.ColorGreen
    colorQuit      = tcell.ColorRed
    highlightSelFg = tcell.NewRGBColor(107, 200, 122)
    highlightFg    = tcell.NewRGBColor(166, 227, 161)
)
```

## Output Functions

```go
func printPath(p string) { fmt.Println(p) }  // prints path to stdout
func fail(err error)     { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
```
