# devpipe - Iteration 2

A local "ready to commit" pipeline runner with TOML configuration support.

## Overview

`devpipe` is a CLI tool that runs a configurable pipeline of checks before committing code. This is **Iteration 2** - now with TOML config support and git modes!

### Current Features (Iteration 2)

- ✅ Single static Go binary
- ✅ **TOML configuration** - Define your own stages per project
- ✅ **Git modes** - `staged`, `staged_unstaged`, or `ref`
- ✅ Git integration (repo detection, changed file tracking)
- ✅ Per-run artifacts under `.devpipe/runs/<run-id>/`
- ✅ Structured `run.json` with stage results
- ✅ Per-stage logs (`logs/<stage-id>.log`)
- ✅ CLI flags: `--config`, `--since`, `--only`, `--skip`, `--fail-fast`, `--dry-run`, `--verbose`, `--fast`
- ✅ Plain text console output
- ✅ Backward compatible - works without config file

## Quick Start

### Build

```bash
make build
```

### Run

```bash
# Run all stages (uses config.toml if present, otherwise built-in stages)
./devpipe

# Run with verbose output
./devpipe --verbose

# Use a specific config file
./devpipe --config my-config.toml

# Compare against a specific git ref
./devpipe --since main

# Run only specific stage
./devpipe --only unit-tests

# Skip stages
./devpipe --skip lint --skip format

# Fast mode (skip stages >= 300s)
./devpipe --fast

# Stop on first failure
./devpipe --fail-fast

# Dry run (don't execute, just show what would run)
./devpipe --dry-run
```

## Configuration

### Creating a config.toml

Copy the example config and customize for your project:

```bash
cp config.toml.example config.toml
```

### Config Structure

```toml
[defaults]
outputRoot = ".devpipe"        # Where to store runs
fastThreshold = 300            # Seconds threshold for --fast

[defaults.git]
mode = "staged_unstaged"       # Git mode: staged, staged_unstaged, or ref
ref = "main"                   # Ref to compare against (when mode = ref)

[stage_defaults]
enabled = true
workdir = "."
estimatedSeconds = 10

[stages.lint]
name = "Lint"
group = "quality"
command = "npm run lint"
estimatedSeconds = 5

[stages.build]
name = "Build"
group = "release"
command = "npm run build"
estimatedSeconds = 30
```

### Git Modes

- **`staged`** - Only staged files (`git diff --cached`)
- **`staged_unstaged`** - Staged + unstaged files (`git diff HEAD`)
- **`ref`** - Compare against specific ref (`git diff <ref>`)

Override with `--since`:
```bash
./devpipe --since origin/main
```

## Makefile Commands

```bash
# Build & Run
make build          # Build the devpipe binary
make run            # Build and run with default settings
make clean          # Remove build artifacts and .devpipe directory

# Demo commands (for testing)
make demo           # Run basic pipeline
make demo-verbose   # Run with verbose output
make demo-fast      # Run with --fast (skip long stages)
make demo-fail-fast # Run with --fail-fast
make demo-only      # Run only unit-tests stage
make demo-skip      # Run pipeline, skip lint and format
make demo-dry-run   # Dry run (don't execute commands)

# Utilities
make show-runs      # List all pipeline runs
make show-latest    # Show latest run.json
make hello-test     # Test hello-world.sh directly
```

## Output Structure

After running `devpipe`, you'll find:

```
.devpipe/
└── runs/
    └── 2025-11-28T21-13-26Z_692593/
        ├── run.json              # Structured run metadata
        └── logs/
            ├── lint.log
            ├── format.log
            ├── type-check.log
            ├── build.log
            ├── unit-tests.log
            └── e2e-tests.log

artifacts/                        # Created by hello-world.sh
├── build/
│   └── app.txt
└── test/
    └── junit.xml
```

## run.json Schema

Each run creates a `run.json` with:

```json
{
  "runId": "2025-11-28T21-13-26Z_692593",
  "timestamp": "2025-11-28T21:13:32Z",
  "repoRoot": "/Users/drew/repos/devpipe",
  "outputRoot": "/Users/drew/repos/devpipe/.devpipe",
  "git": {
    "inGitRepo": true,
    "repoRoot": "/Users/drew/repos/devpipe",
    "diffBase": "HEAD",
    "changedFiles": ["main.go", "hello-world.sh"]
  },
  "flags": {
    "fast": false,
    "failFast": false,
    "dryRun": false,
    "verbose": false
  },
  "stages": [
    {
      "id": "lint",
      "name": "Lint",
      "group": "quality",
      "status": "PASS",
      "exitCode": 0,
      "command": "/path/to/hello-world.sh lint",
      "workdir": "/Users/drew/repos/devpipe",
      "logPath": ".devpipe/runs/.../logs/lint.log",
      "startTime": "2025-11-28T21:13:26Z",
      "endTime": "2025-11-28T21:13:27Z",
      "durationMs": 1014,
      "estimatedSeconds": 5
    }
    // ... more stages
  ]
}
```

## Testing with hello-world.sh

The `hello-world.sh` script simulates real pipeline commands without requiring actual tools:

```bash
# Test individual commands
./hello-world.sh lint
./hello-world.sh build
./hello-world.sh unit-tests

# Or test all at once
make hello-test
```

It creates dummy artifacts:
- `artifacts/build/app.txt` - Simulated build output
- `artifacts/test/junit.xml` - Dummy JUnit XML for testing

### Simulating Failures

You can simulate stage failures using the `DEVPIPE_TEST_FAIL` environment variable:

```bash
# Make the lint stage fail
DEVPIPE_TEST_FAIL=lint ./devpipe --verbose

# Make the format stage fail
DEVPIPE_TEST_FAIL=format ./devpipe --fail-fast

# Test that --fail-fast actually stops on first failure
make test-fail-fast

# Test that pipeline continues without --fail-fast
make test-continue-on-fail

# Run all failure tests
make test-failures
```

## Hardcoded Stages (Iteration 1)

| ID | Name | Group | Est. Time | Command |
|----|------|-------|-----------|---------|
| `lint` | Lint | quality | 5s | `hello-world.sh lint` |
| `format` | Format | quality | 5s | `hello-world.sh format` |
| `type-check` | Type Check | correctness | 10s | `hello-world.sh type-check` |
| `build` | Build | release | 15s | `hello-world.sh build` |
| `unit-tests` | Unit Tests | correctness | 20s | `hello-world.sh unit-tests` |
| `e2e-tests` | E2E Tests | correctness | 600s | `hello-world.sh e2e-tests` |

## Roadmap

### Iteration 2 - Config + TOML + Git modes
- [ ] TOML config support (`config.toml`)
- [ ] Configurable stages
- [ ] Git modes (staged, staged_unstaged, ref)
- [ ] `--config` flag

### Iteration 3 - UI engine + colors
- [ ] TTY detection
- [ ] UI modes: minimal, full
- [ ] Live progress bars
- [ ] Color output

### Iteration 4 - Metrics, summary.json, HTML dashboard
- [ ] Metrics parsing (JUnit, ESLint, SARIF)
- [ ] `summary.json` aggregation
- [ ] HTML dashboard (`report.html`)
- [ ] Historical analytics

### Iteration 5 - Autofix + AI stubs
- [ ] Auto-fix on failure
- [ ] Build artifact contracts
- [ ] AI commit message generation
- [ ] AI PR review stubs

## Development

### Requirements

- Go 1.20+ (tested with Go 1.25.4)
- Git (for repo detection)

### Build from source

```bash
git clone <repo>
cd devpipe
go mod init github.com/drew/devpipe
make build
```

### Testing

```bash
# Run all demo scenarios
make demo
make demo-verbose
make demo-fast
make demo-dry-run
make demo-only
make demo-skip

# View results
make show-runs
make show-latest
```

## License

TBD

## Credits

Built as a local pipeline runner for development workflows.
