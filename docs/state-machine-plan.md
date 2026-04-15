# State Machine Implementation Plan

## Overview

Refactor the TUI key handling to use a state machine pattern for cleaner, more maintainable code.

## Design

### Core Types

```go
// Action - what the state machine should do
type Action int
const (
    ActionNone Action = iota
    ActionDraw
    ActionQuit
    ActionSelect
)

// KeyEvent - wraps tcell event info
type KeyEvent struct {
    Key  tcell.Key
    Rune rune
}

// RenderRequest - what main loop needs to draw
type RenderRequest struct {
    ModeIndicator string     // " N ", " I ", " ? "
    ModeStyle     tcell.Style
    Query         string     // for insert mode
    Items         []WorktreeInfo  // filtered list
    Selected      int
    StatusPrompt  string     // " [Y/Enter] yes / [N/ESC] no "
}

// State interface
type State interface {
    HandleKey(e KeyEvent, items []WorktreeInfo, selected int) (State, Action, RenderRequest)
}
```

### States

#### InsertState

- **Data**: `query string`
- **Behavior**:
  - Printable characters (j, k, i, q, n, p, etc.) → add to query
  - Backspace → remove last char
  - ESC → if query exists, clear it; else switch to Normal
  - Arrow Up/Down, Ctrl-j/k/n/p → navigate (NOT type)
  - Enter → select item

#### NormalState

- **Data**: none (selection passed in)
- **Behavior**:
  - j, Ctrl-j → move down
  - k, Ctrl-k → move up
  - Arrow keys, Ctrl-n/p → navigate
  - i → switch to Insert
  - q, ESC → switch to ConfirmQuit

#### ConfirmQuitState

- **Behavior**:
  - Y, y, Enter → quit
  - N, n, ESC → cancel, return to Normal

## Main Loop

```go
func selectWorktreeTUI(worktrees []WorktreeInfo) (int, error) {
    screen, err := tcell.NewScreen()
    // ... init ...

    state := InsertState{query: ""}
    selected := 0
    worktrees := worktrees

    for {
        // Get render request from current state
        _, action, req := state.HandleKey(KeyEvent{}, worktrees, selected)

        // Draw
        draw(screen, req, worktrees, selected)

        // Handle action
        if action == ActionQuit {
            return -1, fmt.Errorf("cancelled")
        }
        if action == ActionSelect {
            return worktrees[selected].AbsPath, nil
        }

        // Poll and process event
        ev := screen.PollEvent()
        keyEvent := toKeyEvent(ev)
        state, _, req = state.HandleKey(keyEvent, worktrees, selected)
    }
}
```

## Key Bindings Summary

| Key | Insert Mode | Normal Mode | Confirm Mode |
|-----|-------------|-------------|---------------|
| j | type 'j' | move down | - |
| k | type 'k' | move up | - |
| i | type 'i' | switch to insert | - |
| q | type 'q' | confirm quit | - |
| n | type 'n' | cancel confirm | cancel |
| p | type 'p' | - | - |
| y | type 'y' | - | quit |
| Ctrl-j | move down | move down | - |
| Ctrl-k | move up | move up | - |
| Ctrl-n | move down | move down | - |
| Ctrl-p | move up | move up | - |
| Arrow Up | move up | move up | - |
| Arrow Down | move down | move down | - |
| Enter | select | select | quit |
| Backspace | delete char | - | - |
| ESC | clear query / normal | confirm quit | cancel |
| Ctrl-c | quit | quit | quit |

## Implementation Steps

1. Define core types (Action, KeyEvent, RenderRequest, State interface)
2. Implement InsertState
3. Implement NormalState
4. Implement ConfirmQuitState
5. Update main loop to use state machine
6. Test all keybindings

## Colors (Customizable)

```go
var (
    selectedBg  = tcell.ColorWhite
    selectedFg  = tcell.ColorBlack
    colorNormal = tcell.ColorBlue
    colorInsert = tcell.ColorGreen
    colorQuit   = tcell.ColorRed
    filterBg    = tcell.ColorDarkGray
    filterFg    = tcell.ColorWhite
    modeFg      = tcell.ColorBlack
)
```