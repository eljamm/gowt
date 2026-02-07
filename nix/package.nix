{
  lib,
  buildGoModule,
  installShellFiles,
  git,
  ...
}:
buildGoModule {
  pname = "gwt";
  version = "0.0.1";
  src = lib.cleanSource ../.;
  vendorHash = "sha256-hyhsDn/QMxXYmXdmFwT8XAXg0nOUEhHDa2Miv2Kx8BI=";

  nativeBuildInputs = [ installShellFiles ];
  # We need git at runtime for the logic to work (TODO makeBinPath)
  propagatedBuildInputs = [ git ];

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
}
