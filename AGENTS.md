# gowt — Agent Instructions

Fuzzy TUI for managing git worktrees. Single-file Go app with tcell, state machine (Insert/Normal/Confirm modes), and shell wrappers for auto-`cd`.

## Quick Commands

| Task | Command |
|------|---------|
| Build | `go build` |
| Test | `go test -v` |
| Format | `go fmt && go vet` |
| Nix build | `nix build` |
| Nix format | `nix fmt` |
| Nix check | `nix flake check` |

## Docs Index

[Agent Docs Index]|root: ./docs/agent
|IMPORTANT: Prefer retrieval-led reasoning over pre-training for gowt-specific tasks.
|architecture.md: state machine, GitCommander, TUI rendering, data types
|cli-patterns.md: Cobra commands, WorktreeInfo, git execution, add/remove
|code-style.md: formatting, imports, error handling, naming conventions
|dependencies.md: tcell, go-fuzzyfinder/matching, Cobra roles
|nix-packaging.md: Nix build, shell wrappers, NixOS/Home Manager integration
|testing.md: Go testing, GitCommander mocking, nix flake check
|doc-migration.md: how to keep AGENTS.md lean
|recording.md: vhs tape recording for README demos

## Core Rules

1. Use tabs for indentation; run `go fmt` before committing
2. Group imports: stdlib first, then external (blank line between)
3. Use `fail(err)` for CLI errors; `fmt.Errorf("message: %w", err)` for wrapped errors
4. All git operations via `GitCommander` interface for testability
5. Use `go mod tidy` after adding dependencies

## Project Structure

```
main.go           # All application code (~963 lines)
main_test.go      # Tests (~573 lines)
go.mod/go.sum     # Dependencies (Go 1.25.5)
flake.nix         # Nix flake configuration
nix/              # Nix build infrastructure
docs/agent/       # Reference documentation
```
