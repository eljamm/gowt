# JK Motion TUI Implementation Plan

## Overview

Replace go-fuzzyfinder with tcell to implement vim-like jk navigation in the worktree selector.

## Iterative Implementation Steps

---

### Phase 1: Baseline TUI with tcell

**Goal**: Get working TUI that displays worktrees, replaces fuzzyfinder.

1. Add `github.com/gdamore/tcell/v2` dependency
2. Create `selectWorktreeTUI()` replacing `selectWorktree()`:
   - Open tcell screen, render worktree list (plain text)
   - Handle window resize, close on window close
   - `Enter` returns selected path
   - `q` exits without selection (cancels)

**Verification**: Run `go run .` - see worktree list, Enter selects, q quits.

---

### Phase 2: Normal mode with j/k navigation

**Goal**: Cursor movement in normal mode.

1. Track `cursorIndex` (0 to len-1)
2. `j` key: cursor down (stop at bottom, no wrap)
3. `k` key: cursor up (stop at top, no wrap)
4. Highlight current selection (simple reverse video)

**Verification**: j/k moves highlight, Enter selects highlighted item.

---

### Phase 3: Insert mode with filter

**Goal**: Filter worktrees by typing.

1. Add `mode` state: `normal` | `insert`
2. `i` in normal: switch to insert mode
3. `i` in insert: insert 'i' character (normal typing)
4. In insert mode:
   - Show input field at bottom of screen
   - Capture printable characters → append to `query`
   - `Backspace` removes last character from query
   - Real-time filter: show only items matching query (case-insensitive substring)
   - Update display on each keystroke
   - If query is empty, show all items

**Verification**: Type to filter list, j/k navigate filtered results.

---

### Phase 4: Mode indicator + ESC behavior

**Goal**: Visual mode indicator and proper ESC handling.

1. Show mode indicator in top-right corner:
   - `NORMAL` in normal mode
   - `INSERT` in insert mode
2. ESC in insert mode:
   - Clear query string
   - Switch to normal mode
   - Show full unfiltered list
3. ESC in normal mode:
   - Show confirmation prompt: "Quit? [ESC] yes / [n] no"
   - Press ESC again → exit (return -1 / cancel)
   - Press 'n' → return to normal mode, clear prompt

**Verification**: Mode shows correctly, ESC clears filter in insert, ESC confirmation works in normal.

---

### Phase 5: Polish

- Ensure all exit paths close screen properly
- Handle edge cases: empty worktree list, no filter matches
- Run `go fmt && go vet` before commit

---

## Key Behaviors Summary

| Key | Normal Mode | Insert Mode |
|-----|-------------|-------------|
| j | cursor down | insert 'j' |
| k | cursor up | insert 'k' |
| i | switch to insert | insert 'i' |
| Enter | select current | select current |
| ESC | quit confirmation | clear filter, switch to normal |
| q | quit (no selection) | quit (no selection) |

---

## Dependencies

- `github.com/gdamore/tcell/v2` - Terminal UI library