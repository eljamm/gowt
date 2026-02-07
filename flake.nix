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
        standalone = pkgs.writeShellScriptBin "gwt" ''
          source ${gwt-bin.shWrapper} && gwt
        '';
      in
      {
        # See README.md for installation instructions
        default = standalone;
        package = standalone;
        gwt = gwt-bin;
      }
    )
    // flake-utils.lib.eachDefaultSystem (system: {
      # nix build .#default -L
      packages.default = self.default;

      # nix run .#gwt -L
      packages.gwt = self.default;
    });
}
