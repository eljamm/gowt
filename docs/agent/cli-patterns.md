# CLI Patterns — Cobra command patterns, WorktreeInfo struct, git execution

## Adding a New Command

1. Create a `runCmdName` handler function following existing patterns
2. Add cobra.Command to rootCmd in `main()`
3. Update README.md with command documentation

Pattern 1 — Inline command:
```go
rootCmd.AddCommand(&cobra.Command{
    Use:   "command",
    Short: "One-line description",
    Args:  cobra.ExactArgs(1), // or cobra.NoArgs, cobra.RangeArgs(n,m)
    Run:   runCommand,
})
```

Pattern 2 — Predefined command with flags:
```go
var removePreview bool
removeCmd := &cobra.Command{
    Use:   "remove",
    Short: "Interactively remove a worktree",
    Run:   func(cmd *cobra.Command, args []string) { runRemove(cmd, args, removePreview) },
}
removeCmd.Flags().BoolP("force", "f", false, "Force removal")
removeCmd.Flags().BoolVar(&removePreview, "preview", false, "Show git status preview")
rootCmd.AddCommand(removeCmd)
```

## WorktreeInfo Struct

```go
type WorktreeInfo struct {
	AbsPath string // absolute path for git commands and cd
	Display string // worktree name for fuzzy finder UI
	Commit  string // commit hash
	Branch  string // branch name (if any)
	IsCwd   bool   // true if this is the current working directory
}
```

## Worktree Operations

- Parse `git worktree list` output manually (no external library)
- Store both absolute path (`AbsPath`) and display name (`Display`)
- Sort order: current worktree first, then alphabetical by Display name
- Parse both `[branch]` and `(detached)` formats from output

## Git Command Execution

```go
// Simple command with output capture
out, err := exec.Command("git", "worktree", "list").Output()

// Command with stderr to terminal
gitCmd := exec.Command("git", "worktree", "add", newPath, branch)
gitCmd.Stderr = os.Stderr
if err := gitCmd.Run(); err != nil { ... }

// Command with working directory specified
out, _ := exec.Command("git", "-C", worktrees[i].AbsPath, "status", "--short").Output()
```
