{
  description = "gwt: Git Worktree Manager (Go edition)";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";

    treefmt-nix.url = "github:numtide/treefmt-nix";
    treefmt-nix.inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      ...
    }:
    flake-utils.lib.eachDefaultSystemPassThrough (
      system:
      let
        pkgs = import nixpkgs { inherit system; };

        gwt-bin = pkgs.callPackage ./nix/package.nix { };
        wrapper = pkgs.callPackage ./nix/wrapper.nix { inherit gwt-bin; };
        package = pkgs.writeShellScriptBin "gwt" ''
          ${pkgs.bashInteractive}/bin/bash -i -c \
          "source ${wrapper}/share/gwt/gwt.sh && gwt"
        '';
      in
      {
        # See README.md
        default = package;
        gwt = package;
        shell = wrapper;
      }
    )
    // flake-utils.lib.eachDefaultSystem (system: {
      # nix build .#default -L
      packages.default = self.default;

      # nix run .#gwt -L
      packages.gwt = self.default;
    });
}
