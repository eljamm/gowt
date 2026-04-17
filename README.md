## What

**gowt** is a fuzzy TUI for managing git worktrees. Select a worktree and instantly jump to it.

Vibe rewrite of [gko/gwt](https://github.com/gko/gwt/blob/master/gwt.sh).

> [!WARNING]
> This project was written by an LLM under heavy human guidance and review.
> If you prefer purely human-written code, use the original project, instead.

## Features

<div align="center">
  <video src="https://github.com/user-attachments/assets/7c878219-5c28-4bc5-9be1-55b0eb7536f2"></video>
</div>

- Fuzzy filtering to quickly find worktrees
- Vim-style navigation (j/k, g/G, Ctrl-d/Ctrl-u)
- Auto-cd to selected worktree with shell wrapper
- Auto-selects root worktree when launching from other worktrees
- Configurable main worktree directory for `gwt add`

## Commands

| Command | Description |
|---------|-------------|
| `gwt` | Jump to a worktree (fuzzy select) |
| `gwt add` | Create a worktree (interactive) |
| `gwt add name` | Create a worktree, prompt for branch |
| `gwt add name branch` | Create a worktree directly |
| `gwt remove` | Remove a worktree (interactive) |
| `gwt config main` | Show main worktree path |
| `gwt config main /path` | Set main worktree path |

### Main Worktree

Set a default directory where new worktrees are created:

```shell
gwt config main ~/worktrees
```

Then creating worktrees uses that path:

```shell
gwt add feature-login    # creates ~/worktrees/feature-login
```

## Usage

### Modes

| Mode | Key to Enter | Description |
|------|-------------|-------------|
| Insert | (default) | Type to fuzzy filter the worktree list |
| Normal | `ESC` | Navigate without filtering |
| Confirm | `q` / `Q` | Confirm quit: `y`/`Enter` to quit, `n`/`ESC` to cancel |

### Insert Mode

Works in Insert mode by default when the TUI opens.

| Key | Action |
|-----|--------|
| Type characters | Fuzzy filter the worktree list |
| `Backspace` | Delete last character |
| `Enter` | Select current worktree |
| `ESC` | Clear query (if non-empty) or switch to Normal |
| `Ctrl-c` | Quit |
| `↓` / `Ctrl-j` / `Ctrl-n` | Move down one line |
| `↑` / `Ctrl-k` / `Ctrl-p` | Move up one line |
| `Ctrl-d` | Page down (half window) |
| `Ctrl-u` | Page up (half window) |

### Normal Mode

Press `ESC` from Insert mode to enter Normal mode.

| Key | Action |
|-----|--------|
| `i` | Switch to Insert mode |
| `q` / `Q` | Enter Confirm mode |
| `Enter` | Select current worktree |
| `ESC` | Enter Confirm mode (quit prompt) |
| `↓` / `Ctrl-j` / `Ctrl-n` | Move down one line |
| `↑` / `Ctrl-k` / `Ctrl-p` | Move up one line |
| `n j` / `n k` | Move by n lines (e.g., `5j`) |
| `g` | Go to first item |
| `G` | Go to last item |
| `n g` | Go to line n (e.g., `5g`) |
| `n G` | Go to line n from bottom (e.g., `5G`) |
| `Ctrl-d` | Page down (half window) |
| `Ctrl-u` | Page up (half window) |
| `n Ctrl-d` / `n Ctrl-u` | Page by n half-windows (e.g., `3Ctrl-d`) |
| `Ctrl-c` | Quit |

### Confirm Mode

Press `q` or `Q` in Normal mode to enter confirm mode.

| Key | Action |
|-----|--------|
| `y` / `Y` / `Enter` | Confirm and quit |
| `n` / `N` / `ESC` | Cancel and return to Normal |

## Quickstart

If you have [Nix](https://nixos.org) installed, you can run this tool without installing it:

```shellSession
cd $(nix run github:eljamm/gowt#gwt)
```

## Install (Nix)

### Flakes

Add the package to your nix flake inputs:

```nix
{
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-25.11";

  inputs.gowt.url = "github:eljamm/gowt/dev";
  inputs.gowt.inputs.nixpkgs.follows = "nixpkgs";
}
```

#### NixOS

Use the following module in your NixOS system:

```nix
{
  inputs,
  pkgs,
  ...
}:
{
  nixpkgs.overlays = [
    (final: prev: {
      gwt = inputs.gowt.default;
    })
  ];

  environment.systemPackages = [
    pkgs.gwt
  ];
}
```

This includes a shell wrapper that automatically `cd`s into the selected worktree.

After rebuilding and switching your system, the tool will be available:

```shellSession
gwt
```

#### Home Manager

Use the following configuration, depending on which shell you want:

```nix
{
  inputs,
  ...
}:
{
  programs.fish.interactiveShellInit = ''
    source ${inputs.gowt.gwt.fishWrapper}
    alias g gwt
  '';

  programs.bash.initExtra = ''
    source ${inputs.gowt.gwt.shWrapper}
  '';

  programs.zsh.initContent = ''
    source ${inputs.gowt.gwt.shWrapper}
    alias g=gwt
  '';
}
```

## Manual Build

Requires Go 1.21+.

```shellSession
git clone https://github.com/eljamm/gowt.git
cd gowt
go install
```

The binary will be installed to `$GOBIN` (defaults to `$GOPATH/bin` or `$HOME/go/bin`).
Make sure that directory is in your `PATH`.

The binary prints the path to the selected worktree, so to switch into it run:

```shellSession
cd $(gwt)
```

Or, to automatically do this, set up a shell wrapper as instructed, below.

### Shell Wrapper

Add the following to your shell config:

<!-- TODO: automate from Nix pacakge -->

**Bash** (add to `~/.bashrc`):

```bash
gwt() {
  output=$(gwt "$@")
  exit_code=$?
  if [ $exit_code -eq 0 ] && [ -d "$output" ]; then
    cd "$output"
  else
    printf "%s\n" "$output"
  fi
}
```

**Zsh** (add to `~/.zshrc`):

```zsh
gwt() {
  output=$(gwt "$@")
  exit_code=$?
  if [ $exit_code -eq 0 ] && [ -d "$output" ]; then
    cd "$output"
  else
    printf "%s\n" "$output"
  fi
}
```

**Fish** (add to `~/.config/fish/config.fish`):

```fish
function gwt
    set -l output (gwt $argv)
    set -l exit_code $status
    if test $exit_code -eq 0 -a -d "$output"
        cd "$output"
    else
        printf "%s\n" "$output"
    end
end
```

Restart your shell, and then switching worktrees will be as easy as:

```shellSession
gwt
```
