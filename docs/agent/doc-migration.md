# Doc Migration — How to update AGENTS.md using the Vercel compressed format

## Overview

The AGENTS.md file should stay lean (~8KB max). Bulky content (full API references,
long examples, exhaustive option lists) belongs in reference files under `docs/agent/`.

## Audit Phase

Before updating, audit the current AGENTS.md:

1. **Identify all sections** — list them out
2. **Flag bulky but rarely needed content** — full code examples, exhaustive option lists, long prose sections
3. **Flag always-needed content** — build commands, test commands, code style rules, critical gotchas

## Migration Steps

### Step 1 — Create lean AGENTS.md

Target: under 8KB. Include:
- Project overview (2-3 sentences)
- Build, lint, and test commands
- Code style rules (concise, bullet form)
- Pipe-delimited docs index pointing to reference files
- PR and commit guidelines
- Security gotchas (if any)

Docs index format:
```
## Reference docs (read when relevant)

docs/agent/cli-patterns.md   | Cobra command patterns, WorktreeInfo struct, git execution
docs/agent/code-style.md     | Formatting, imports, error handling, naming conventions
docs/agent/nix-packaging.md | Nix package build, shell wrappers, NixOS/Home Manager integration
docs/agent/testing.md        | Go testing strategy, nix flake check, verifying package outputs
```

### Step 2 — Extract bulky content into reference files

One file per logical domain under `docs/agent/`. Each file must have:
- A 1-line summary header on line 1 (e.g., `# CLI Patterns — Cobra command patterns...`)
- Content organized by topic

### Step 3 — Verify

```bash
nix build               # or your repo's build command
wc -c AGENTS.md         # confirm under 8KB
head -n 1 docs/agent/*.md  # confirm each has a 1-line summary
```

## Decision Guide

| Content type | Stay in AGENTS.md | Move to docs/agent/ |
|---|---|---|
| Build/test/lint commands | ✅ | |
| Code style bullets (no examples) | ✅ | |
| Core architecture (1-2 sentences) | ✅ | |
| Full code examples | | ✅ |
| Exhaustive option lists | | ✅ |
| Long prose explanations | | ✅ |
| Nix/flake internals | | ✅ |
| Testing strategy with examples | | ✅ |
| Dependencies table (brief) | ✅ | |

## When Adding New Content

1. Ask: will an agent need this on every run, or only when working on that specific area?
2. If only for specific area → put in `docs/agent/` reference file
3. If needed almost always → add to AGENTS.md (if size allows)
4. If AGENTS.md exceeds 8KB, extract something else to a reference file
