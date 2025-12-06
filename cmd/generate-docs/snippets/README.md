# Documentation Snippets

This directory contains markdown snippets that are included in generated documentation.

## Files

| File | Used In | Purpose |
|------|---------|---------|
| `config-intro.md` | `docs/configuration.md` | Introduction and quick start |
| `phase-examples.md` | `docs/configuration.md` | Phase execution examples |
| `cli-intro.md` | `docs/cli-reference.md` | CLI commands table |
| `cli-examples.md` | `docs/cli-reference.md` | Usage examples |
| `config-validation.md` | `docs/config-validation.md` | Full validation documentation |

## How It Works

The generator (`../main.go`) reads these snippets and combines them with auto-generated content:

1. **Auto-generated** - Config tables from struct tags, CLI flags from code
2. **Snippets** - Examples, explanations, prose

## Editing

1. Edit the snippet file
2. Run `make generate-docs`
3. Commit both the snippet and generated files

The git hook will auto-regenerate if you modify files in this directory.
