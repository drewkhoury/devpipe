# devpipe

Fast, local pipeline runner for development workflows.

`devpipe` reads your config.toml and runs the tasks you specify.

## Features

- 🚀 **Single binary** - No dependencies, just download and run
- 🎨 **Beautiful UI** - Animated progress, colored output, grouping
- ⚙️ **Phase-based execution** - Organize tasks into sequential phases, running all tasks in a phase in parallel
- 🔧 **Auto-fix** - Automatically fix formatting, linting, and other fixable issues
- 📝 **TOML configuration** - Simple, readable config files
- 🔀 **Git integration** - Run checks on staged, unstaged, or ref-based changes
- 📊 **Metrics & Dashboard** - JUnit/artifact parsing, HTML reports
- 🎯 **Flexible** - Run all, skip some, or target specific tasks

![devpipe in action](devpipe.png)

## Quick Start

### Install

**Homebrew (macOS/Linux):**

```bash
brew install drewkhoury/tap/devpipe
```

In a rush? Skip to [examples](#examples) once you have `devpipe` installed, or just run it from your project root with no arguments to auto-generate a config.toml with example tasks.

See [cli-reference](#cli-reference) for more details about runtime options.

<details>
<summary>More Install Options</summary>

---

**Direct download:**

```bash
# Set PLATFORM to: darwin-arm64, darwin-amd64, linux-amd64, or linux-arm64
#  darwin-arm64 (macOS Apple Silicon)
#  darwin-amd64 (macOS Intel)
#  linux-amd64 (Linux x86_64)
#  linux-arm64 (Linux ARM64)
PLATFORM=darwin-arm64
curl -L https://github.com/drewkhoury/devpipe/releases/latest/download/devpipe_*_${PLATFORM}.tar.gz | tar xz
chmod +x devpipe
```

**Build from source:**

```bash
git clone https://github.com/drewkhoury/devpipe
cd devpipe
make build
```

</details>

## Configuration

`devpipe` expects `config.toml` from the current directory by default. If no config file is specified and `config.toml` doesn't exist, devpipe will auto-generate one with example tasks.

Example `config.toml`:

```toml
[tasks.lint]
command = "npm run lint"

[tasks.test]
command = "npm test"

[tasks.build]
command = "npm run build"
```

### Configuration Reference

See [config.example.toml](config.example.toml) for a full example.

<details open>
<summary><h4 style="display: inline;">Detailed Configuration Reference</h4></summary>

#### Order of Precedence

All configuration values in devpipe are resolved in this order (highest to lowest priority):

1. **CLI flags** (e.g., `--fix-type`, `--since`, `--ui`)
2. **Task level** (e.g., `[tasks.go-fmt] fixType = "auto"`)
3. **Task defaults** (e.g., `[task_defaults] fixType = "helper"`)
4. **Built-in defaults** (e.g., `helper`)

#### [defaults] Section

Global configuration options.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `outputRoot` | string | No | `.devpipe` | Directory for run outputs and logs |
| `fastThreshold` | int | No | `300` | Tasks longer than this (seconds) are skipped with `--fast` |
| `uiMode` | string | No | `basic` | UI mode: `basic` or `full` |
| `animationRefreshMs` | int | No | `500` | Dashboard refresh rate in milliseconds |
| `animatedGroupBy` | string | No | `phase` | Group tasks by `phase` or `type` in dashboard |

#### [defaults.git] Section

Git integration settings.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `mode` | string | No | `staged_unstaged` | Git mode: `staged`, `staged_unstaged`, or `ref` |
| `ref` | string | No | `HEAD` | Git ref to compare against when mode is `ref` |

#### [task_defaults] Section

Set default values that apply to all tasks unless overridden at the task level.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | bool | No | `true` | Whether tasks are enabled by default |
| `workdir` | string | No | `.` | Default working directory for tasks |
| `fixType` | string | No | `helper` | Default fix behavior: `auto`, `helper`, or `none` |

#### [tasks.\<task-id\>] Section

Individual task configuration. Task ID must be unique.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `command` | string | **Yes** | - | Shell command to execute |
| `name` | string | No | - | Display name for the task |
| `desc` | string | No | - | Description |
| `type` | string | No | - | Task type for grouping (e.g., `check`, `build`, `test`) |
| `workdir` | string | No | `.` | Working directory for this task |
| `enabled` | bool | No | `true` | Whether this task is enabled |
| `fixType` | string | No | (inherited) | Fix behavior: `auto`, `helper`, `none` (overrides task_defaults) |
| `fixCommand` | string | No | - | Command to run to fix issues (required if fixType is set) |
| `metricsFormat` | string | No | - | Metrics format: `junit`, `artifact` |
| `metricsPath` | string | No | - | Path to metrics file (relative to workdir) |

#### Phase Headers

Use `[tasks.phase-<name>]` to create phase headers. Tasks under a phase header run in parallel. Phases execute sequentially.

You can add comments to make it easier to identify phases in the config at a glance (though these are not required).

```toml

[tasks.phase-quality]
##################################

[tasks.lint]
command = "npm run lint"

[tasks.format]
command = "npm run format:check"

[tasks.phase-build]
##################################

[tasks.build]
command = "npm run build"

[tasks.phase-test]
##################################

[tasks.unit-tests]
command = "npm run test:unit"

[tasks.e2e-tests]
command = "npm run test:e2e"
```

Parallel execution:

```
Phase 1: Quality Checks
├─ lint (parallel)
└─ format (parallel)
    ↓ (wait for phase to complete)
Phase 2: Build
└─ build
    ↓ (wait for phase to complete)
Phase 3: Tests
├─ unit-tests (parallel)
└─ e2e-tests (parallel)
```

</details>

### Auto-Fix

`devpipe` can automatically fix issues when tasks fail. This is useful for formatting checks, linting, and other fixable issues.

#### Fix Types

- **`helper`** (default): Show a suggestion on how to fix the issue
- **`auto`**: Automatically run the fix command and re-check
- **`none`**: Don't show fix suggestions (useful to override global defaults)

<details>
<summary>More Info</summary>

#### Configuration Example

```toml
[task_defaults]
fixType = "helper"  # Default: show fix suggestions

[tasks.go-fmt]
command = "make check-fmt"
fixType = "auto"           # Override: auto-fix this task
fixCommand = "make fmt"    # Command to fix formatting

[tasks.go-vet]
command = "go vet ./..."
fixCommand = "make vet-fix"  # Uses fixType="helper" from task_defaults

[tasks.security-scan]
command = "make security"
fixType = "none"  # Don't suggest fixes for security issues
```

#### CLI Override

Override fix behavior for all tasks:

```bash
./devpipe --fix-type auto    # Auto-fix all tasks with fixCommand
./devpipe --fix-type helper  # Show suggestions only
./devpipe --fix-type none    # Disable all fix suggestions
```

#### Example Output

**Helper mode:**
```
[go-fmt         ] ✗ FAIL (80ms)
[go-fmt         ] 💡 To fix run: make fmt
```

**Auto mode:**
```
[go-fmt         ] ✗ FAIL (77ms)

[go-fmt         ] 🔧 Auto-fixing: make fmt (31ms)
[go-fmt         ] ✅ Fix succeeded, re-checking...
[go-fmt         ] ✅ PASS (48ms)
Summary:
  ✓ go-fmt          PASS       0.16s (156ms) [auto-fixed]
```

The total time includes: initial check + fix + re-check.

</details>

## Modes

### UI Modes

**`basic` mode** provides simple text output with status symbols. It's the default UI mode.

```bash
./devpipe -ui basic
```

Simple text output with status symbols:
- ✓ PASS (green)
- ✗ FAIL (red)
- ⊘ SKIPPED (yellow)

**`full` mode** provides borders, and better visuals, and more detailed formatting.

```bash
./devpipe -ui full
```

### Dashboard & Full UI Modes

Dashboard mode provides a live progress view, with animated progress bars and detailed task information.

This can be used with any UI mode (basic or full).

```bash
./devpipe --dashboard -ui full
```

```
╔═════════════════════════════════════════════════════════╗
║ devpipe run 2025-11-29T05-25-25Z_071352                 ║
║ Repo: /Users/you/project                                ║
║ Git: staged_unstaged | Files: 6                         ║
╚═════════════════════════════════════════════════════════╝

Overall: ████████████████████████████████████████ 100%

┌─ Quality Checks ──────────────────────────────────────┐
│ ✓ lint         1s                         
│ ✓ format       1s                         
└───────────────────────────────────────────────────────┘

┌─ Build ───────────────────────────────────────────────┐
│ ✓ build        3s                         
└───────────────────────────────────────────────────────┘

┌─ Tests ───────────────────────────────────────────────┐
│ ✓ unit-tests   2s                         
│ ✓ e2e-tests    5s                          
└───────────────────────────────────────────────────────┘
```

## CLI Reference

### Commands

| Command | Description |
|---------|-------------|
| `devpipe` | Run the pipeline with default or specified config |
| `devpipe validate [files...]` | Validate one or more config files |
| `devpipe help` | Show help information |

### Run Flags

| Flag | Description | Default |
|------|-------------|---------|
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
|------|-------------|---------|
| `--config <path>` | Path to config file to validate (supports multiple files) | `config.toml` |

See [CONFIG-VALIDATION.md](CONFIG-VALIDATION.md) for more details.

### Examples

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
devpipe --dashboard -ui full --verbose

# Validate config files
devpipe validate
devpipe validate config/*.toml
```

## Git Modes

Control which files are in scope for changes:

- **`staged`** - Only staged files (`git diff --cached`)
- **`staged_unstaged`** - Staged + unstaged (`git diff HEAD`)
- **`ref`** - Compare against ref (`git diff <ref>`)

```bash
# Check only staged files
./devpipe --config config/config-staged.toml

# Compare against main branch
./devpipe --since main
```

devpipe will collect this information but doesn't handle the logic within each task.

## Metrics & Dashboard

devpipe can parse test results and generate HTML dashboards:

```toml
[tasks.unit-tests]
metricsFormat = "junit"
metricsPath = "test-results/junit.xml"

[tasks.build]
metricsFormat = "artifact"
metricsPath = "dist/app.js"
```

View the dashboard:
```bash
open .devpipe/report.html
```

## Output Structure

```
.devpipe/
├── report.html             # HTML dashboard
├── summary.json            # Aggregated metrics
└── runs/
    └── 2025-11-29T05-25-25Z_071352/
        ├── run.json        # Run metadata
        ├── pipeline.log    # Verbose output log
        └── logs/
            ├── lint.log
            ├── build.log
            └── unit-tests.log
```

## Where you can use Devpipe

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit
./devpipe --config .devpipe-precommit.toml --fail-fast
```

### CI/CD

```yaml
# .github/workflows/ci.yml
- name: Run devpipe
  run: |
    curl -L https://github.com/drewkhoury/devpipe/releases/latest/download/devpipe_*_linux_amd64.tar.gz | tar xz
    ./devpipe --no-color
```

### Local Development

```bash
# Quick checks before commit
./devpipe --fast

# Full pipeline with dashboard
./devpipe --dashboard -ui full
```

## License

Apache 2.0 - see [LICENSE](LICENSE) for details.

## Contributing

Contributions welcome! Please open an issue or PR.

---

Built by [Andrew Khoury](https://github.com/drewkhoury)
