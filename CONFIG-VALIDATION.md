# Config Validation

The `devpipe validate` command validates TOML configuration files to ensure they have valid syntax and correct configuration options.

## Usage

### Validate default config.toml
```bash
devpipe validate
# or
make validate
```

### Validate a specific config file
```bash
devpipe validate config/config-phases.toml
```

### Validate multiple config files
```bash
devpipe validate config/*.toml
# or
make validate-all
```

## What It Validates

### TOML Syntax
- Checks for valid TOML syntax
- Reports line numbers and specific syntax errors

### Configuration Structure
- **Unknown fields**: Detects fields that don't exist in the config schema
- **Unknown sections**: Detects sections that aren't part of the config structure

### Defaults Section (`[defaults]`)
- **uiMode**: Must be one of: `basic`, `animated`, `simple`
- **animatedGroupBy**: Must be one of: `type`, `phase`
- **fastThreshold**: Must be non-negative
- **animationRefreshMs**: Must be non-negative

### Git Configuration (`[defaults.git]`)
- **mode**: Must be one of: `staged`, `staged_unstaged`, `ref`
- **ref**: Warning if mode is `ref` but no ref is specified

### Task Defaults (`[task_defaults]`)
- **estimatedSeconds**: Must be non-negative

### Tasks (`[tasks.*]`)
- **command**: Required for non-phase tasks
- **type**: Warning if not one of the common types: `quality`, `correctness`, `security`, `release`
- **estimatedSeconds**: Must be non-negative
- **metricsFormat**: Must be one of: `junit`, `eslint`, `sarif`, `artifact`
- **metricsPath**: Warning if metricsFormat is set but metricsPath is missing (and vice versa)

### Phase Headers
- Phase headers (tasks starting with `phase-`) should have a `name` but no `command`
- The `desc` field is supported for phase descriptions

## Exit Codes

- **0**: Configuration is valid (may have warnings)
- **1**: Configuration is invalid (has errors)
- **1**: File not found or other error

## Output Format

The validator provides clear, color-coded output:

```
✅ Configuration is valid!
```

Or with errors:
```
❌ Found 3 error(s):
  • [defaults.uiMode] Invalid UI mode 'invalid_mode'. Valid options: basic, animated, simple
  • [tasks.test-task.command] Task must have a command
  • [tasks.test-task.estimatedSeconds] Estimated seconds must be non-negative

⚠️  Found 1 warning(s):
  • [tasks.test-task.type] Unknown task type 'invalid_type'. Common types: quality, correctness, security, release

❌ Configuration is INVALID
```

## Examples

### Valid Configuration
```toml
[defaults]
outputRoot = ".devpipe"
fastThreshold = 300
uiMode = "basic"

[defaults.git]
mode = "staged_unstaged"

[task_defaults]
enabled = true
workdir = "."
estimatedSeconds = 10

[tasks.lint]
name = "Lint"
type = "quality"
command = "./hello-world.sh lint"
estimatedSeconds = 5
```

### Invalid Configuration Examples

**Invalid UI Mode:**
```toml
[defaults]
uiMode = "invalid_mode"  # ERROR: must be basic, animated, or simple
```

**Missing Command:**
```toml
[tasks.my-task]
name = "My Task"
# ERROR: command is required
```

**Invalid Metrics Format:**
```toml
[tasks.test]
name = "Test"
command = "npm test"
metricsFormat = "invalid"  # ERROR: must be junit, eslint, sarif, or artifact
```

## Implementation

The validation logic is implemented in:
- `internal/config/validate.go` - Core validation logic
- `main.go` - CLI command handler
- `Makefile` - Convenience commands

The validator uses the TOML decoder's metadata to detect unknown fields and performs semantic validation on known fields.
