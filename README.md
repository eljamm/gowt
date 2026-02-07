## What

Vibe rewrite of <https://github.com/gko/gwt/blob/master/gwt.sh>

Licensed the same as the above project.

## Quickstart

To run this tool without innstalling it:

```shellSession
nix run github:eljamm/gowt
```

## Install

Don't use this, this is a quick vibed tool based on the above, use that if
interested in such a tool.

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
