## Examples

```bash
# Run with default config
devpipe

# Run with custom config
devpipe --config config/custom.toml

# Skip slow tasks and stop on first failure
devpipe --fast --fail-fast

# Run only specific task
devpipe --only lint

# Skip multiple tasks
devpipe --skip e2e-tests --skip integration-tests

# Dry run to see what would execute
devpipe --dry-run

# Full dashboard with verbose output
devpipe --dashboard --ui full --verbose

# Validate config files
devpipe validate
devpipe validate config/*.toml
```
