# Architecture — Why state machine, tcell, Commander interface

## Overview

The TUI is built from scratch using `tcell` for rendering and a state machine pattern for input handling. The binary only prints paths; shell wrappers (bash/zsh/fish) auto-`cd` into the selected worktree.

## Why State Machine

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

**Rationale**: Vim users expect modal input. A state machine makes each mode's behavior explicit and easy to test independently.

## Why tcell

`tcell` provides:
- Cross-platform terminal handling
- Key event polling (`screen.PollEvent()`)
- Character rendering (`screen.SetContent()`)
- Style application (`tcell.Style`)

**Rationale**: Lightweight alternative to Bubble Tea. Full control over rendering for custom fuzzy-matching UI.

## Why Commander Interface

All git operations go through `git.Commander` interface in `internal/git` for testability.

**Rationale**: Testability. Without aCommander interface, tests would either call real git (slow, flaky) or use globals (race conditions). The interface enables dependency injection with `MockCommander`.

## Why go-fuzzyfinder/matching

Uses the `matching` subpackage only (not the higher-level `Find()` API).

**Rationale**: The UI is built from scratch with tcell. The matching package provides fuzzy algorithms; tcell handles rendering.

## Default Selection

On TUI open, the default selection is determined by:

1. **In root worktree** → select last worktree (bottom of list)
2. **In non-root worktree** → select root worktree (detected via `gitCommonDir` path matching)

**Rationale**: Common workflow — start in root, jump to feature branch; or start in feature branch, jump back to root.