# Recording Demo Videos

Use [vhs](https://github.com/charmbracelet/vhs) to record terminal sessions for README demos.

## Prerequisites

```bash
nix develop .
```

## Worktree Setup

Create demo worktrees for recording:

```bash
go run ./scripts/demo-worktrees.go setup   # Create 50 random adjective-animal worktrees
go run ./scripts/demo-worktrees.go clean   # Remove all demo worktrees
```

## Recording

### via devShell

```bash
nix develop . --command record-demo
```

### via nix run

```bash
nix run .#record-demo
```

## Tape Script Format

```
Output docs/demo.gif

Set FontSize 16
Set FontFamily "JetBrains Mono"
Set Width 128
Set Height 48
Set Theme "Dark+"

Type "./gwt"
Enter
Sleep 500ms

Type "j"
Sleep 100ms

# ... more commands
```

### Common Commands

| Command | Description |
|---------|-------------|
| `Type "text"` | Type text |
| `Enter` | Press Enter |
| `Escape` or `Esc` | Press Escape |
| `Sleep Xms` | Wait X milliseconds |
| `Ctrl+C` | Press Ctrl+C |
| `Ctrl+D` | Press Ctrl+D |

### Key Names

- Arrow keys: `Up`, `Down`, `Left`, `Right`
- Control keys: `Ctrl+A`, `Ctrl+C`, etc.
- Special: `Backspace`, `Delete`, `Tab`, `Enter`, `Escape`

## Examples

See `docs/demo.tape` for a working example.
