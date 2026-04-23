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
  version = "0.3.1";

  src = lib.cleanSourceWith {
    name = "source";
    src = ../.;

    filter =
      name: type:
      let
        base = baseNameOf (toString name);
      in
      (type == "directory" && base != ".wt")
      || lib.hasSuffix ".go" base
      || base == "go.mod"
      || base == "go.sum";
  };

  vendorHash = "sha256-SD0K8oMvs+bDPSBWD3j/NqEHeYm6LBIGmLk1ySC1kXw=";

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
    fishWrapper =
      writeText "gwt.fish"
        # fish
        ''
          function gwt
              set -l target (${finalAttrs.finalPackage}/bin/gwt $argv)
              or return

              if test -d "$target"
                  builtin cd "$target"
                  return
              end

              string join \n $target
          end

          source ${finalAttrs.finalPackage}/share/fish/vendor_completions.d/gwt.fish
        '';

    # NOTE: for zsh https://stackoverflow.com/a/74323525/8608146 is required
    shWrapper =
      writeText "gwt.sh"
        # bash
        ''
          gwt() {
              local target

              target=$(${finalAttrs.finalPackage}/bin/gwt "$@") || return

              if [ -d "$target" ]; then
                  builtin cd "$target"
                  return
              fi

              printf "%s\n" "$target"
          }

          if [ -n "$ZSH_VERSION" ]; then
              source ${finalAttrs.finalPackage}/share/zsh/site-functions/_gwt
          elif [ -n "$BASH_VERSION" ]; then
              source ${finalAttrs.finalPackage}/share/bash-completion/completions/gwt.bash
          fi
        '';
  };
})
