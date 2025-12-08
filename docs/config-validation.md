# Config Validation

The `devpipe validate` command validates TOML configuration files to ensure they have valid syntax and correct configuration options.

## Usage

### Validate default config.toml
```bash
devpipe validate
```

### Validate a specific config file
```bash
devpipe validate config/config-phases.toml
```

### Validate multiple config files
```bash
devpipe validate config/*.toml
```

## What It Validates

### TOML Syntax
- Checks for valid TOML syntax
- Reports line numbers and specific syntax errors

### Configuration Structure
- **Unknown fields**: Detects fields that don't exist in the config schema
- **Unknown sections**: Detects sections that aren't part of the config structure

### Defaults Section (`[defaults]`)
- **uiMode**: Must be one of: `basic`, `full`
- **animatedGroupBy**: Must be one of: `type`, `phase`
- **fastThreshold**: Must be non-negative
- **animationRefreshMs**: Must be between 20-2000 (milliseconds)

### Git Configuration (`[defaults.git]`)
- **mode**: Must be one of: `staged`, `staged_unstaged`, `ref`
- **ref**: Warning if mode is `ref` but no ref is specified

### Task Defaults (`[task_defaults]`)
- **enabled**: Whether tasks are enabled by default
- **workdir**: Default working directory for tasks
- **fixType**: Must be one of: `auto`, `helper`, `none`

### Tasks (`[tasks.*]`)
- **command**: Required for non-phase tasks
- **type**: Warning if not one of the common types: `quality`, `correctness`, `security`, `release`
- **fixType**: Must be one of: `auto`, `helper`, `none`
- **fixCommand**: Required if fixType is set at task level (except when fixType is `none`)
- **outputType**: Must be one of: `junit`, `sarif`, `artifact`
- **outputPath**: Warning if outputType is set but outputPath is missing (and vice versa)

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
❌ Found 2 error(s):
  • [defaults.uiMode] Invalid UI mode 'invalid_mode'. Valid options: basic, full
  • [tasks.test-task.command] Task must have a command

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
fixType = "helper"

[tasks.lint]
name = "Lint"
type = "quality"
command = "./hello-world.sh lint"
fixCommand = "./hello-world.sh lint --fix"
```

### Invalid Configuration Examples

**Invalid UI Mode:**
```toml
[defaults]
uiMode = "invalid_mode"  # ERROR: must be basic or full
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
outputType = "invalid"  # ERROR: must be junit, sarif, or artifact
```

**Invalid Fix Type:**
```toml
[tasks.lint]
name = "Lint"
command = "npm run lint"
fixType = "invalid"  # ERROR: must be auto, helper, or none
```

**Missing Fix Command:**
```toml
[tasks.lint]
name = "Lint"
command = "npm run lint"
fixType = "auto"  # ERROR: fixCommand is required when fixType is set
```

## Implementation

The validation logic is implemented in:
- `internal/config/validate.go` - Core validation logic
- `main.go` - CLI command handler
- `Makefile` - Convenience commands

The validator uses the TOML decoder's metadata to detect unknown fields and performs semantic validation on known fields.
