# CLI Reference

## Commands

| Command | Description |
|---------|-------------|
| `devpipe` | Run the pipeline with default or specified config |
| `devpipe validate [files...]` | Validate one or more config files |
| `devpipe help` | Show help information |


### Run Flags

| Flag | Description | Default |
|------|-------------|---------||
| `--config <path>` | Path to config file | `config.toml` |
| `--since <ref>` | Git ref to compare against (overrides config) | - |
| `--only <task-id>` | Run only a single task by id | - |
| `--skip <task-id>` | Skip a task by id (repeatable) | - |
| `--fix-type <type>` | Fix behavior: `auto`, `helper`, `none` (overrides config) | - |
| `--ui <mode>` | UI mode: `basic`, `full` | `basic` |
| `--dashboard` | Show dashboard with live progress | `false` |
| `--fail-fast` | Stop on first task failure | `false` |
| `--fast` | Skip long-running tasks (> fastThreshold) | `false` |
| `--dry-run` | Do not execute commands, simulate only | `false` |
| `--verbose` | Show verbose output (always logged to pipeline.log) | `false` |
| `--no-color` | Disable colored output | `false` |

### Validate Flags

| Flag | Description | Default |
|------|-------------|---------||
| `--config <path>` | Path to config file to validate (supports multiple files) | `config.toml` |

See [config-validation.md](config-validation.md) for more details.

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

