## What

Vibe rewrite of <https://github.com/gko/gwt/blob/master/gwt.sh>

Licensed the same as the above project.

> [!WARNING]
> This project is mainly for experimenting with LLMs and to scratch a personal itch.
> If you're interested in a tool like this, use the project mentioned above, instead.

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
| `↓` / `↑` or `Ctrl-n` / `Ctrl-p` | Move up/down one line |
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
| `j` / `k` or `↓` / `↑` | Move up/down one line |
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

To run this tool without innstalling it:

```shellSession
nix run github:eljamm/gowt
```

## Install

### Flakes

Add the package to your nix flake inputs:

```nix
{
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-25.11";

  inputs.gowt.url = "github:eljamm/gowt/dev";
  inputs.gowt.inputs.nixpkgs.follows = "nixpkgs";
}
```

#### standalone

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

After rebuilding and switching your system, the tool will be available:

```shellSession
gwt
```

#### home-manager

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
