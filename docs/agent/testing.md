# Testing — Go testing strategy, nix flake check, verifying package outputs

## Go Testing

- Create `*_test.go` files alongside source files
- Use table-driven tests when appropriate
- Test edge cases: empty input, errors, multiple items
- Mock git commands with `exec.Command`

## Nix flake check

```bash
nix flake check    # Run all checks (builds all packages on all systems)
```

## Verifying Package Outputs

```bash
nix build
$(nix-build --no-link -A gwt)/bin/gwt --version   # Verify binary works
```

Checks are defined in `nix/flake/default.nix` and include all packages from `default.packages`.
