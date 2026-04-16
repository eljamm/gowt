{
  flake-inputs ? import (fetchTarball {
    url = "https://github.com/fricklerhandwerk/flake-inputs/tarball/4.1.0";
    sha256 = "1j57avx2mqjnhrsgq3xl7ih8v7bdhz1kj3min6364f486ys048bm";
  }),
  self ? flake-inputs.import-flake { src = ./.; },
  inputs ? self.inputs,
  system ? builtins.currentSystem,
  pkgs ? import inputs.nixpkgs {
    config = { };
    overlays = [ ];
    inherit system;
  },
  lib ? import "${inputs.nixpkgs}/lib",
}:
let
  default = lib.makeScope pkgs.newScope (def: {
    inherit
      lib
      pkgs
      self
      system
      inputs
      flake
      default
      ;

    # Custom library. Contains helper functions, builders, ...
    devLib = def.callPackage ./nix/lib.nix { };
    go = def.callPackage ./nix/go.nix { };

    formatter = def.callPackage ./nix/formatter.nix { };
    gwt-bin = pkgs.callPackage ./nix/package.nix { };

    # tool for declaratively recording project demo
    vhs = pkgs.vhs.overrideAttrs (oldAttrs: {
      version = "0.11.0-unstable-2026-04-06";

      src = pkgs.fetchFromGitHub {
        owner = "charmbracelet";
        repo = "vhs";
        rev = "6ec73298e7d3cd39eaa2a6dc8fe08fcc80f97d66";
        hash = "sha256-a7lZzb0PhKDxcmJiJnCpRjz/g1/Czby6LaPN6RhrTRo=";
      };

      vendorHash = "sha256-cgKLYUATtn4hMdIOXZe9JWYNUOrX3S6BDfvS+rIWDfM=";

      patches = oldAttrs.patches or [ ] ++ [
        # Add keystroke captions and overlays for recorded videos
        # https://github.com/charmbracelet/vhs/pull/719
        ./nix/patches/vhs-keystroke-captions.patch
      ];
    });

    record-demo = pkgs.writeShellScriptBin "record-demo" ''
      ${def.vhs}/bin/vhs docs/demo/recording.tape
    '';

    devShells.default = pkgs.mkShellNoCC {
      inputsFrom = [ def.go.shells.default ];
      packages = [
        def.formatter.package
        def.record-demo
        def.vhs
      ];
    };

    overlays.default = final: prev: def.devPkgs;
  });

  flake = default.callPackage ./nix/flake { };

  # return final scope, with computed and non-recursive attributes
  finalScope = default.packages default;
in
finalScope
