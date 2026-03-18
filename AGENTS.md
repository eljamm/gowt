# AGENTS.md

CLI tool for managing git worktrees via a fuzzy TUI. Single-file Go application using Cobra and go-fuzzyfinder.

## Build Commands

```bash
# Go
go build              # Build binary (outputs to ./gwt)
go run .              # Run without building
go install            # Install to $GOPATH/bin
go test               # Run all tests
go test -v            # Run with verbose output
go test -run TestName # Run specific test by name
go fmt && go vet      # Format and analyze
go mod tidy           # Clean up dependencies

# Nix
nix build             # Build package (outputs to ./result)
nix develop           # Enter dev shell
nix fmt               # Format all files (treefmt)
nix flake check       # Run all checks
```

## Project Structure

```
docs           # Project documentation
main.go        # All application code
go.mod         # Module definition and dependencies
go.sum         # Locked dependency versions
flake.nix      # Nix flake configuration
```

## Code Style

- Follow standard Go conventions (go.dev/doc/effective_go)
- Single-file application (all code in main.go)
- Use tabs for indentation; run `go fmt` before committing
- Group imports: stdlib first, then external (blank line between)
- Use `fail(err)` for CLI errors; `fmt.Errorf("message: %w", err)` for wrapped errors
- Return errors from helpers, handle in callers
- See docs/agent/code-style.md for naming conventions and code examples

## Reference docs (read when relevant)

docs/agent/cli-patterns.md   | Cobra command patterns, WorktreeInfo struct, git execution
docs/agent/code-style.md     | Formatting, imports, error handling, naming conventions
docs/agent/nix-packaging.md  | Nix package build, shell wrappers, NixOS/Home Manager integration
docs/agent/testing.md        | Go testing strategy, nix flake check, verifying package outputs
docs/agent/doc-migration.md  | How to migrate/update AGENTS.md to the Vercel compressed format

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/ktr0731/go-fuzzyfinder` | Fuzzy TUI selection |
| `github.com/spf13/cobra` | CLI framework |

Use `go mod tidy` after adding new dependencies.

## Commit & PR Guidelines

- Use conventional commits: `feat:`, `fix:`, `refactor:`, `docs:`
- Keep commits focused: one logical change per commit
- Write concise commit messages (subject line under 72 chars)
- Run `go fmt` and `go vet` before committing
- Create PRs via `gh pr create` with a clear title and summary

## Quick Reference

| Task | Command |
|------|---------|
| Run tests | `go test -v` |
| Build | `go build` |
| Format | `go fmt && go vet` |
| Nix build | `nix build` |
| Nix format | `nix fmt` |
| Nix check | `nix flake check` |
