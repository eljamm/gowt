# Code Style — Formatting, imports, error handling, naming conventions

## General Principles

- Follow standard Go conventions (go.dev/doc/effective_go)
- Keep code simple and readable
- Write tests for new functionality
- Single-file application (all code in main.go)

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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
)
```

## Error Handling

- Use `fail(err)` for CLI errors (prints to stderr, exits with code 1)
- Use `fmt.Errorf("message: %w", err)` for wrapped errors
- Return errors from helper functions, handle in callers
- Ignore errors only when truly safe (use `_` sparingly)

## Output Functions

```go
func printPath(p string) { fmt.Println(p) }  // prints path to stdout
func fail(err error)     { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
```
