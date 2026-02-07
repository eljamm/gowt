{
  lib,
  stdenv,
  writeText,
  gwt-bin,
}:
let
  wrapper = stdenv.mkDerivation {
    pname = "gwt-wrapper";
    version = "0.0.1";
    phases = [ "installPhase" ];
    installPhase = ''
      mkdir -p $out/share/gwt
      cp ${shWrapper} $out/share/gwt/gwt.sh
      cp ${fishWrapper} $out/share/gwt/gwt.fish
    '';
  };

  # 2. The Smart Wrapper (Fish)
  fishWrapper = writeText "gwt.fish" ''
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
  shWrapper = writeText "gwt.sh" ''
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
in
wrapper
