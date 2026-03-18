# Nix Packaging — Nix package build, shell wrappers, NixOS/Home Manager integration

## Nix Build

```bash
nix build        # Build the package (outputs to ./result)
```

The flake exposes `gwt` (raw binary) and `default` (standalone shell wrapper). See `flake.nix` for details.

## Nix Flake Structure

```nix
{
  # Raw binary
  default = standalone;  # shell wrapper for interactive use
  package = standalone;
  gwt     = gwt-bin;    # raw binary

  # Development
  devShells.default     # Go dev shell (gotestsum, delve, gopls, formatters)
}
```

Dev aliases (available in `nix develop`):
- `td` — run tests with readable output (`gotestdox ./...`)
- `tw` — watch mode (`gotestsum --watch`)
- `tt` — run all tests (`gotestsum`)
- `ff` — format project (treefmt)

## Nix Formatters

```bash
nix fmt          # Format all Nix and Go files (treefmt)
```

Formatters enabled: nixfmt, gofumpt, goimports-reviser, actionlint, zizmor, editorconfig-checker.

## NixOS Module Integration

```nix
{
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-25.11";
  inputs.gowt.url = "github:eljamm/gowt/dev";
  inputs.gowt.inputs.nixpkgs.follows = "nixpkgs";
}
```

### System package:
```nix
{
  nixpkgs.overlays = [
    (final: prev: { gwt = inputs.gowt.default; })
  ];
  environment.systemPackages = [ pkgs.gwt ];
}
```

### Home Manager (Fish/Bash/Zsh wrappers):
```nix
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

## Shell Wrappers

The package provides two shell wrappers that auto-detect whether output is a directory and `cd` into it:

- `shWrapper` — Bash/Zsh compatibility via shell function
- `fishWrapper` — Fish compatibility via shell function

Both wrappers are generated in `nix/package.nix` via `passthru`.
