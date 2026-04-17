# Code Style — Formatting, imports, error handling, naming conventions

## General Principles

- Follow standard Go conventions (go.dev/doc/effective_go)
- Keep code simple and readable
- Write tests for new functionality
- Modular structure in internal/ packages

## Formatting

- Use tabs for indentation (configured in .editorconfig)
- Run `go fmt` before committing
- Group imports: stdlib first, then external packages (blank line between)
- Keep lines under ~120 characters
- No unnecessary blank lines within functions

## Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Variables | camelCase | `worktreePath`, `removePreview` |
| Constants | camelCase or SCREAMING_SNAKE_CASE | `baseArgs` |
| Functions | snake_case | `runJump`, `getWorktreesSorted` |
| Types | PascalCase | `WorktreeInfo`, `lineInfo` |
| Acronyms | Keep consistent case | `repoName`, not `RepoName` |

## Import Organization

```go
import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/ktr0731/go-fuzzyfinder/matching"
	"github.com/spf13/cobra"
)
```

## Error Handling

- Use `fail(err)` for CLI errors (prints to stderr, exits with code 1)
- Use `fmt.Errorf("message: %w", err)` for wrapped errors
- Return errors from helper functions, handle in callers
- Ignore errors only when truly safe (use `_` sparingly)
- Use `strings.EqualFold()` for case-insensitive string comparison
- Handle all errors from `exec.Command()` - never ignore with `_` unless documented

## Output Functions

```go
func printPath(p string) { fmt.Println(p) }  // prints path to stdout
func fail(err error)     { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
```

## Helper Functions

Extract reusable patterns into helper functions to avoid duplication.

```go
// Branch extraction from git worktree line
// internal/app/handlers.go
func ExtractBranch(line string) string {
	if idx := strings.Index(line, "["); idx != -1 {
		if end := strings.Index(line[idx:], "]"); end != -1 {
			return line[idx+1 : idx+end]
		}
	}
	if idx := strings.Index(line, "("); idx != -1 {
		if end := strings.Index(line[idx:], ")"); end != -1 {
			return line[idx+1 : idx+end]
		}
	}
	return ""
}
```

## Context for Exec Commands

Use `context.Context` with timeout for long-running git commands.

```go
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()

cmd := exec.CommandContext(ctx, "git", "worktree", "add", path, branch)
cmd.Stderr = os.Stderr
if err := cmd.Run(); err != nil {
    fail(fmt.Errorf("failed to create worktree: %w", err))
}
```
