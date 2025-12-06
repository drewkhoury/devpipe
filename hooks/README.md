# Git Hooks

This directory contains git hooks that can be installed to automate development tasks.

## Installation

```bash
make setup-hooks
```

This copies hooks from this directory to `.git/hooks/`.

## Available Hooks

### `pre-commit`

Automatically regenerates documentation when config files change.

**What it does:**
- Detects changes to `internal/config/` files
- Runs `make generate-docs`
- Adds generated files (`config.example.toml`, `config.schema.json`, `docs/configuration.md`) to your commit

**Why?**
Ensures documentation is always in sync with code changes. You never have to remember to run `make generate-docs` manually.

## Manual Installation

If you prefer not to use `make setup-hooks`:

```bash
cp hooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

## Disabling Hooks

To temporarily skip hooks:

```bash
git commit --no-verify
```

To permanently remove:

```bash
rm .git/hooks/pre-commit
```
