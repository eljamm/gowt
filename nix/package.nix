{
  lib,
  buildGoModule,
  installShellFiles,
  git,

  # shell wrappers
  writeText,
  ...
}:
buildGoModule (finalAttrs: {
  pname = "gwt";
  version = "0.0.1";

  src = lib.cleanSourceWith {
    name = "source";
    src = ../.;
    filter =
      name: type:
      let
        base = baseNameOf (toString name);
      in
      (
        (type == "directory" && base != ".wt")
        || lib.hasSuffix ".go" base
        || base == "go.mod"
        || base == "go.sum"
      );
  };

  vendorHash = "sha256-CJLTDn/ib2htab3gpyEYbAFq3iEjvc1nJ+2eqO9pYqE=";

  nativeBuildInputs = [ installShellFiles ];

  # We need git at runtime for the logic to work (TODO makeBinPath)
  propagatedBuildInputs = [ git ];

  ldflags = [ "-s" ];

  meta.mainProgram = "gwt";

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

  passthru = {
    # 2. The Smart Wrapper (Fish)
    fishWrapper = writeText "gwt.fish" ''
      function gwt
          # Run the binary and capture output
          set -l output (${finalAttrs.finalPackage}/bin/gwt $argv)
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
      source ${finalAttrs.finalPackage}/share/fish/vendor_completions.d/gwt.fish
    '';

    # 3. The Smart Wrapper (Bash/Zsh)
    # NOTE: for zsh https://stackoverflow.com/a/74323525/8608146 is required
    shWrapper = writeText "gwt.sh" ''
      gwt() {
        local output
        output=$(${finalAttrs.finalPackage}/bin/gwt "$@")
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
         source ${finalAttrs.finalPackage}/share/zsh/site-functions/_gwt
      elif [ -n "$BASH_VERSION" ]; then
         source ${finalAttrs.finalPackage}/share/bash-completion/completions/gwt.bash
      fi
    '';
  };
})
