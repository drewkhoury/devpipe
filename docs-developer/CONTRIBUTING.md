# Contributing to devpipe

Thanks for your interest in contributing!

## How to Contribute

1. **Fork the repository**
2. **Make your changes** with clear commit messages
3. **Test your changes** by running `make build` and testing locally
4. **Submit a pull request** with a description of your changes

## Reporting Issues

- Use GitHub Issues to report bugs or suggest features
- Include steps to reproduce for bugs
- Provide example config files when relevant

## Development

```bash
# Build
make build

# Run tests
make test

# Validate configs
make validate

# Generate documentation (after changing config structs)
make generate-docs

# Check if docs are up to date
make check-docs
```

### Documentation Generation

Documentation is auto-generated from Go struct tags and markdown snippets.

**Quick commands:**
```bash
make generate-docs    # Regenerate all docs
make check-docs       # Verify docs are in sync
make setup-hooks      # Install git hook (auto-regenerates)
```

**📖 For complete details, see [DOCUMENTATION-GENERATION.md](DOCUMENTATION-GENERATION.md)**

This covers:
- How the generator works (struct tags + snippets)
- What files are generated
- How to add config fields
- How to edit examples/prose
- Git hook setup
- CI integration

## Code Style

- Follow standard Go conventions
- Keep commits focused and atomic
- Write clear commit messages

## Questions?

Open an issue or reach out via GitHub discussions.
