# Testing — Go testing strategy, nix flake check, verifying package outputs

## Go Testing

- Create `*_test.go` files alongside source files
- Use table-driven tests when appropriate
- Test edge cases: empty input, errors, multiple items

## Mocking Git Commands

Use the `git.Commander` interface from `internal/git` for testable code:

```go
// internal/git/commander.go
type Commander interface {
    WorktreeList() (string, error)
    RevParse(showToplevel bool) (string, error)
    GitCommonDir() (string, error)
    WorktreeRemove(path string, force bool) error
    WorktreeAdd(path, branch string) error
    WorktreeAddNew(path, branch string) error
}
```

Use `git.MockCommander` from `internal/git/mock.go` in tests:

```go
// internal/git/mock.go
type MockCommander struct {
    WorktreeListOutput string
    WorktreeListErr    error
    // ...
}

func (m *MockCommander) WorktreeList() (string, error) {
    return m.WorktreeListOutput, m.WorktreeListErr
}
```

Inject via function parameters:

```go
// internal/app/handlers.go
func GetWorktreesSorted(commander git.Commander) ([]tui.WorktreeInfo, error) {
    // use commander.WorktreeList(), commander.RevParse(), etc.
}
```

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
