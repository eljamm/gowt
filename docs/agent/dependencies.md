# Dependencies — tcell, go-fuzzyfinder/matching, Cobra

## Overview

Three direct dependencies. Use `go mod tidy` after any change.

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/gdamore/tcell/v2` | v2.13.8 | TUI rendering engine |
| `github.com/ktr0731/go-fuzzyfinder` | v0.9.0 | Fuzzy matching (matching subpackage only) |
| `github.com/spf13/cobra` | v1.10.2 | CLI framework |

## tcell

Terminal screen library for building TUIs from scratch. Handles:
- Screen initialization and cleanup
- Key event polling (`screen.PollEvent()`)
- Character rendering (`screen.SetContent()`)
- Style application (`tcell.Style`)

```go
screen, err := tcell.NewScreen()
defer screen.Fini()
screen.Init()
screen.Clear()
screen.Show()
```

## go-fuzzyfinder/matching

Only the `matching` subpackage is used. Provides fuzzy matching algorithms:

```go
import "github.com/ktr0731/go-fuzzyfinder/matching"

matched := matching.FindAll(query, displayStrs, matching.WithMode(matching.ModeSmart))
```

The higher-level `Find()` API is NOT used — the UI is built from scratch with tcell.

## Cobra

CLI command framework. Three commands:

- `gwt` — root: fuzzy-select worktree
- `gwt add <branch>` — create worktree
- `gwt remove [-f]` — remove worktree

```go
rootCmd := &cobra.Command{
    Use:   "gwt",
    Short: "Git Worktree Manager",
    Run:   func(cmd *cobra.Command, args []string) { runJump(cmd, args) },
}
rootCmd.AddCommand(&cobra.Command{...})
```
