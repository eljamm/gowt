# gowt — Agent Instructions

Fuzzy TUI for managing git worktrees. Modular Go app with tcell, state machine (Insert/Normal/Confirm modes), and shell wrappers for auto-`cd`.

## Quick Commands

| Task | Command |
|------|---------|
| Build | `go build` |
| Test | `go test -v` |
| Format | `go fmt && go vet` |
| Nix build | `nix build` |
| Nix format | `nix fmt` |
| Nix check | `nix flake check` |
| Record demo | `nix run .#record-demo` |

## Docs Index

[Reference Docs]|root: ./docs/agent
|architecture.md: Architecture decisions (why state machine, tcell, Commander interface)

For Nix installation, shell wrappers, and NixOS/Home Manager: see README.md

## Core Rules

1. Use tabs for indentation; run `go fmt` before committing
1. Group imports: stdlib first, then external (blank line between)
1. Use `fail(err)` for CLI errors; `fmt.Errorf("message: %w", err)` for wrapped errors
1. All git operations via `git.Commander` interface in internal/git for testability
1. Use `go mod tidy` after adding dependencies
1. **Keep AGENTS.md lean** — migrate bulky content (full code examples, exhaustive lists) to docs/agent/ reference files; AGENTS.md should stay under 8KB

## Project Structure

```
.
├── main.go                 # CLI wiring only (~43 lines)
├── main_test.go            # Tests (~540 lines)
├── go.mod / go.sum         # Dependencies
├── flake.nix               # Nix flake configuration
├── nix/                    # Nix build infrastructure
├── internal/
│   ├── app/
│   │   ├── config.go       # Config (main worktree path)
│   │   ├── config_test.go  # Config tests
│   │   └── handlers.go     # RunJump, RunAdd, RunRemove, RunConfig
│   ├── git/
│   │   ├── commander.go    # Commander interface + RealCommander
│   │   └── mock.go         # MockCommander for tests
│   └── tui/
│       ├── state.go        # State types (Action, KeyEvent, InsertState, etc.)
│       └── tui.go          # SelectWorktreeTUI, TUIColors
└── docs/agent/             # Reference documentation
```
