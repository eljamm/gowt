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

        gwt-bin = pkgs.buildGoModule {
          pname = "gwt";
          version = "0.0.1";
          src = pkgs.lib.cleanSource ./.;
          vendorHash = "sha256-hyhsDn/QMxXYmXdmFwT8XAXg0nOUEhHDa2Miv2Kx8BI=";

          nativeBuildInputs = [ pkgs.installShellFiles ];
          # We need git at runtime for the logic to work (TODO makeBinPath)
          propagatedBuildInputs = [ pkgs.git ];

          ldflags = [ "-s" ];

          postInstall = ''
            # Generate completions
            $out/bin/gwt completion bash > gwt.bash
            $out/bin/gwt completion zsh  > _gwt
            $out/bin/gwt completion fish > gwt.fish

            # --- THE CRITICAL FIX ---
            # We replace '$args[1]' (which resolves to 'gwt', the function)
            # with the absolute path to the binary.
            # This ensures TAB completion bypasses the shell function entirely.

            sed -i "s|\$args\[1\] __complete|$out/bin/gwt __complete|g" gwt.fish

            installShellCompletion --cmd gwt \
              --bash gwt.bash \
              --zsh _gwt \
              --fish gwt.fish
          '';
        };

        # 2. The Smart Wrapper (Fish)
        fishWrapper = pkgs.writeText "gwt.fish" ''
          function gwt
              # Run the binary and capture output
              set -l output (${gwt-bin}/bin/gwt $argv)
              set -l exit_code $status

              if test $exit_code -eq 0
                  # CRITICAL: Only cd if it is actually a directory
                  if test -d "$output"
                      builtin cd "$output"
                  else
                      # Otherwise just print the text (Help, Version, etc)
                      printf "%s\n" "$output"
                  end
              else
                  # Print errors
                  printf "%s\n" "$output"
                  return $exit_code
              end
          end

          # Source the patched completions
          source ${gwt-bin}/share/fish/vendor_completions.d/gwt.fish
        '';

        # 3. The Smart Wrapper (Bash/Zsh)
        # NOTE: for zsh https://stackoverflow.com/a/74323525/8608146 is required
        shWrapper = pkgs.writeText "gwt.sh" ''
          gwt() {
            local output
            output=$(${gwt-bin}/bin/gwt "$@")
            local exit_code=$?

            if [ $exit_code -eq 0 ]; then
              if [ -d "$output" ]; then
                builtin cd "$output"
              else
                printf "%s\n" "$output"
              fi
            else
              printf "%s\n" "$output"
              return $exit_code
            fi
          }

          if [ -n "$ZSH_VERSION" ]; then
             source ${gwt-bin}/share/zsh/site-functions/_gwt
          elif [ -n "$BASH_VERSION" ]; then
             source ${gwt-bin}/share/bash-completion/completions/gwt.bash
          fi
        '';
        wrapper = pkgs.stdenv.mkDerivation {
          pname = "gwt-wrapper";
          version = "0.0.1";
          phases = [ "installPhase" ];
          installPhase = ''
            mkdir -p $out/share/gwt
            cp ${shWrapper} $out/share/gwt/gwt.sh
            cp ${fishWrapper} $out/share/gwt/gwt.fish
          '';
        };
      in
      {
        # nix build .#package -L
        package = wrapper;
      }
    )
    // flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        # nix run .#default -L
        packages.default = pkgs.writeShellScriptBin "gwt-demo" ''
          ${pkgs.bashInteractive}/bin/bash -i -c "source ${self.package}/share/gwt/gwt.sh && gwt"
        '';
      }
    );
}
