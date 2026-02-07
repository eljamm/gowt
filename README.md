## What

Vibe rewrite of https://github.com/gko/gwt/blob/master/gwt.sh

Licensed the same as the above project.

## Install

Don't use this, this is a quick vibed tool based on the above, use that if
interested in such a tool.

Only flakes + home-manager setup supported at this point.

```nix
# add it to your nix flake inputs
# add an overlay
(_: _: {
  gwt = inputs.gowt.package;
})
```

Then you need to do this in your home-manager config

```nix
  programs.fish.interactiveShellInit = ''
    source ${pkgs.gwt}/share/gwt/gwt.fish
    alias g gwt
  '';

  programs.bash.initExtra = ''
    source ${pkgs.gwt}/share/gwt/gwt.sh
  '';

  programs.zsh.initContent = ''
    source ${pkgs.gwt}/share/gwt/gwt.sh
    alias g=gwt
  '';
```
