# Documentation Generation

Documentation is **auto-generated** from two sources:
1. **Go struct tags** (for config field tables)
2. **Markdown snippets** (for examples and explanations)

## Generated Files

- `config.example.toml` - Annotated TOML example
- `config.schema.json` - JSON Schema for IDEs
- `docs/configuration.md` - Config reference (tables + snippets)
- `docs/cli-reference.md` - CLI flags (tables + snippets)
- `docs/config-validation.md` - Validation docs (from snippet)

**Never edit these files directly!** They're regenerated from struct tags and snippets.

## How to Add a New Config Field

1. **Add struct tag** in `internal/config/config.go`:
   ```go
   type DefaultsConfig struct {
       LogLevel string `toml:"logLevel" doc:"Logging verbosity" enum:"debug,info,warn"`
   }
   ```

2. **Add default** in `internal/config/defaults.go`:
   ```go
   LogLevel: "info",
   ```

3. **Regenerate**:
   ```bash
   make generate-docs
   ```

That's it! The generator uses reflection to extract everything from struct tags.

## How to Edit Examples/Prose

Edit markdown snippets in `cmd/generate-docs/snippets/`:

- `config-intro.md` - Config documentation intro
- `phase-examples.md` - Phase execution examples
- `cli-intro.md` - CLI commands table
- `cli-examples.md` - CLI usage examples
- `config-validation.md` - Validation documentation

After editing snippets, run `make generate-docs` to regenerate docs.

## Git Hook (Recommended)

Install the pre-commit hook to auto-generate docs:

```bash
make setup-hooks
```

This installs `hooks/pre-commit` which automatically:
1. Detects when you modify `internal/config/` or `cmd/generate-docs/snippets/`
2. Runs `make generate-docs`
3. Adds generated files to your commit

Once installed, you never have to remember to regenerate docs!

## CI Check

The GitHub Actions workflow (`.github/workflows/ci.yml`) includes:

```yaml
- name: Check documentation is up to date
  run: make check-docs
```

This fails CI if someone forgot to regenerate docs, ensuring they're always in sync.
