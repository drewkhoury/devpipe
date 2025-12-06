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
