# devpipe

Fast, local pipeline runner for development workflows.

## Features

- 🚀 **Single binary** - No dependencies, just download and run
- 🎨 **Beautiful UI** - Animated progress, colored output, grouping
- ⚙️ **Phase-based execution** - Organize tasks into sequential phases
- 📝 **TOML configuration** - Simple, readable config files
- 🔀 **Git integration** - Run checks on staged, unstaged, or ref-based changes
- 📊 **Metrics & Dashboard** - JUnit/artifact parsing, HTML reports
- 🎯 **Flexible** - Run all, skip some, or target specific tasks

## Quick Start

### Install

```bash
# Download latest release
curl -L https://github.com/drewkhoury/devpipe/releases/latest/download/devpipe-darwin-arm64 -o devpipe
chmod +x devpipe

# Or build from source
git clone https://github.com/drewkhoury/devpipe
cd devpipe
make build
```

## Configuration

Expects a `config.toml` in the current directory. If none is found, it will generate a default one.

Example `config.toml`:

```toml
[defaults]
animatedGroupBy = "phase"  # or "type"

[defaults.git]
mode = "staged_unstaged"

[task_defaults]
enabled = true
workdir = "."

[tasks.lint]
name = "Lint"
type = "check"
command = "npm run lint"

[tasks.format]
name = "Format Check"
type = "check"
command = "npm run format:check"
```

## Phases

Tasks can be organized into phases using `[tasks.phase-*]` headers. All tasks under a phase header run in parallel, and phases execute sequentially:

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

## UI Modes

### Dashboard Mode

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
└─────────────────────────────────────────────────────────┘

┌─ Build ───────────────────────────────────────────────┐
│ ✓ build        3s                         
└─────────────────────────────────────────────────────────┘

┌─ Tests ───────────────────────────────────────────────┐
│ ✓ unit-tests   2s                         
│ ✓ e2e-tests    5s                          
└─────────────────────────────────────────────────────────┘
```

### Basic Mode

```bash
./devpipe
```

Simple text output with status symbols:
- ✓ PASS (green)
- ✗ FAIL (red)
- ⊘ SKIPPED (yellow)

## CLI Flags

```bash
# Configuration
--config <path>         # Use specific config file
--since <ref>           # Git ref to compare against

# Execution
--only <task-id>        # Run only specific task
--skip <task-id>        # Skip task (repeatable)
--fast                  # Skip long-running tasks
--fail-fast             # Stop on first failure
--dry-run               # Show what would run

# UI
--dashboard             # Show dashboard with live progress
-ui <mode>              # UI mode: basic, full
--no-color              # Disable colors
--verbose               # Verbose output
```

## Git Modes

Control which files are in scope for changes:

- **`staged`** - Only staged files (`git diff --cached`)
- **`staged_unstaged`** - Staged + unstaged (`git diff HEAD`)
- **`ref`** - Compare against ref (`git diff <ref>`)

```bash
# Check only staged files
./devpipe --config config-staged.toml

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
open .devpipe/dashboard.html
```

## Output Structure

```
.devpipe/
├── dashboard.html          # HTML dashboard
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

## Examples

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
    curl -L https://github.com/drewkhoury/devpipe/releases/latest/download/devpipe-linux-amd64 -o devpipe
    chmod +x devpipe
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
