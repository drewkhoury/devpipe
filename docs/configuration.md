# Configuration Reference

This document provides a complete reference for all configuration options in devpipe.

## Quick Start

Minimal configuration:

```toml
[tasks.my-task]
command = "npm test"
```

## Configuration Sections


### `[defaults]`

Global configuration options

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `projectRoot` | string | No | `-` | Repo/project root directory (optional override, auto-detected from git or config location if not set) |
| `outputRoot` | string | No | `.devpipe` | Directory for run outputs and logs |
| `fastThreshold` | int | No | `300` | Tasks longer than this (seconds) are skipped with --fast |
| `uiMode` | string | No | `basic` | UI mode: basic or full (valid: `basic`, `full`) |
| `animationRefreshMs` | int | No | `500` | Dashboard refresh rate in milliseconds |
| `animatedGroupBy` | string | No | `phase` | Group tasks by phase or type in dashboard (valid: `phase`, `type`) |

### `[defaults.git]`

Git integration settings

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `mode` | string | No | `staged_unstaged` | Git mode: staged, staged_unstaged, or ref (valid: `staged`, `staged_unstaged`, `ref`) |
| `ref` | string | No | `HEAD` | Git ref to compare against when mode is ref |

### `[task_defaults]`

Default values that apply to all tasks unless overridden at the task level

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | bool | No | `true` | Whether tasks are enabled by default |
| `workdir` | string | No | `.` | Default working directory for tasks |
| `fixType` | string | No | `-` | Default fix behavior: auto, helper, or none (valid: `auto`, `helper`, `none`) |

### `[tasks.<task-id>]`

Individual task configuration. Task ID must be unique.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `command` | string | **Yes** | `-` | Shell command to execute |
| `name` | string | No | `-` | Display name for the task |
| `desc` | string | No | `-` | Description |
| `type` | string | No | `-` | Task type for grouping (e.g., check, build, test) |
| `workdir` | string | No | `-` | Working directory for this task |
| `enabled` | bool | No | `-` | Whether this task is enabled |
| `metricsFormat` | string | No | `-` | Metrics format: junit, sarif, artifact (valid: `junit`, `sarif`, `artifact`) |
| `metricsPath` | string | No | `-` | Path to metrics file (relative to workdir) |
| `fixType` | string | No | `-` | Fix behavior: auto, helper, none (overrides task_defaults) (valid: `auto`, `helper`, `none`) |
| `fixCommand` | string | No | `-` | Command to run to fix issues (required if fixType is set) |
| `watchPaths` | []string | No | `-` | File patterns to watch (glob patterns relative to workdir). Task runs only if matching files changed. |

## Phase-Based Execution

Use `[tasks.phase-<name>]` to create phase headers. Tasks under a phase header run in parallel. Phases execute sequentially.

```toml
[tasks.phase-quality]
name = "Quality Checks"

[tasks.lint]
command = "npm run lint"

[tasks.format]
command = "npm run format:check"

[tasks.phase-build]
name = "Build"

[tasks.build]
command = "npm run build"
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
```

## Examples

See [config.example.toml](../config.example.toml) for a complete annotated example.

