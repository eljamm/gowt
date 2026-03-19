# Testing — Go testing strategy, nix flake check, verifying package outputs

## Go Testing

- Create `*_test.go` files alongside source files
- Use table-driven tests when appropriate
- Test edge cases: empty input, errors, multiple items

## Mocking Git Commands

Use the `GitCommander` interface for testable code:

```go
type GitCommander interface {
    worktreeList() (string, error)
    revParse(showToplevel bool) (string, error)
    worktreeRemove(path string, force bool) error
    worktreeAdd(path, branch string) error
    worktreeAddNew(path, branch string) error
}
```

Create a mock that returns canned responses:

```go
type mockGitCommander struct {
    worktreeListOutput string
    worktreeListErr    error
    // ...
}

func (m *mockGitCommander) worktreeList() (string, error) {
    return m.worktreeListOutput, m.worktreeListErr
}
```

Inject via function parameters:

```go
func getWorktreesSorted(commander GitCommander) ([]WorktreeInfo, error) {
    // use commander.worktreeList(), commander.revParse(), etc.
}
```

Use `setCommander(mock)` in tests to inject mock behavior.

## Test Structure

| Test | What It Covers |
|------|----------------|
| `TestExtractBranch` | Branch parsing from `[...]` and `(...)` formats |
| `TestGetWorktreesSorted` | Parsing `git worktree list` output |
| `TestFindWorktreePathForBranch` | Branch lookup and "not found" error |

## Run Tests

```bash
go mod tidy && go test -v
```

## Nix flake check

```bash
nix flake check    # Run all checks (builds all packages on all systems)
```

## Verifying Package Outputs

```bash
nix build
$(nix-build --no-link -A gwt)/bin/gwt --version   # Verify binary works
```

Checks are defined in `nix/flake/default.nix` and include all packages from `default.packages`.
