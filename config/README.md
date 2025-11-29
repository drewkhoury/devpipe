# Example Configuration Files

This directory contains specialized example configuration files demonstrating specific devpipe features and use cases.

## Configuration Files

| File | Purpose |
|------|---------|
| `config-phases.toml` | Demonstrates phase-based task organization with parallel execution |
| `config-staged.example.toml` | Minimal config for staged-only git mode |
| `config-animation.example.toml` | Shows animation refresh rate options and performance tuning |

## Main Configuration Reference

**The canonical, comprehensive configuration example is [`../config.example.toml`](../config.example.toml) in the project root.**

That file includes:
- All available configuration options with detailed documentation
- Complete field reference
- Best practices and examples
- Phase headers, metrics, and git integration

To get started, copy the example to `config.toml` or your own custom config file name.

## Notes

- All files in this directory are validated by `make validate-all`
- These are reference examples - `make run` and `make demo` use `/config.toml` in the project root
- These examples focus on specific features, not complete configurations