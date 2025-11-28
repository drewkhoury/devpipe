# devpipe Testing Guide

This document describes how to test devpipe Iteration 1.

## Quick Test

Run all tests in one command:

```bash
make test-failures
```

This will verify:
- ✅ `--fail-fast` stops on first failure
- ✅ Pipeline continues on failure without `--fail-fast`

## Manual Testing Scenarios

### 1. Success Scenarios

```bash
# All stages pass
make demo

# Verbose output
make demo-verbose

# Fast mode (skip long stages)
make demo-fast

# Dry run (no execution)
make demo-dry-run

# Run only one stage
make demo-only

# Skip specific stages
make demo-skip
```

### 2. Failure Scenarios

#### Test --fail-fast behavior

```bash
# Automated test
make test-fail-fast

# Manual test
DEVPIPE_TEST_FAIL=format ./devpipe --fail-fast --verbose
```

**Expected behavior:**
- Lint stage runs and passes
- Format stage runs and fails
- Pipeline stops immediately
- Type-check, build, unit-tests, e2e-tests do NOT run
- Exit code is 1
- run.json contains only 2 stages

#### Test continue-on-failure behavior

```bash
# Automated test
make test-continue-on-fail

# Manual test
DEVPIPE_TEST_FAIL=format ./devpipe --verbose
```

**Expected behavior:**
- Lint stage runs and passes
- Format stage runs and fails
- Pipeline continues
- Type-check, build, unit-tests, e2e-tests all run
- Exit code is 1 (because format failed)
- run.json contains all 6 stages

### 3. Failure Simulation

You can make any stage fail using the `DEVPIPE_TEST_FAIL` environment variable:

```bash
# Fail lint
DEVPIPE_TEST_FAIL=lint ./devpipe

# Fail build
DEVPIPE_TEST_FAIL=build ./devpipe

# Fail unit-tests
DEVPIPE_TEST_FAIL=unit-tests ./devpipe --fail-fast
```

## Test Matrix

| Test | Command | Expected Result |
|------|---------|-----------------|
| All pass | `./devpipe` | Exit 0, all stages PASS |
| All pass (verbose) | `./devpipe --verbose` | Exit 0, shows commands |
| Fast mode | `./devpipe --fast` | Exit 0, e2e-tests SKIPPED |
| Dry run | `./devpipe --dry-run` | Exit 0, all stages SKIPPED |
| Only one stage | `./devpipe --only build` | Exit 0, only build runs |
| Skip stages | `./devpipe --skip lint --skip format` | Exit 0, 4 stages run |
| Fail + continue | `DEVPIPE_TEST_FAIL=format ./devpipe` | Exit 1, all 6 stages run |
| Fail + stop | `DEVPIPE_TEST_FAIL=format ./devpipe --fail-fast` | Exit 1, only 2 stages run |

## Verifying Output

### Check run.json

```bash
# View latest run
make show-latest

# Or manually
cat .devpipe/runs/<run-id>/run.json | jq .
```

### Check logs

```bash
# View a specific stage log
cat .devpipe/runs/<run-id>/logs/lint.log

# View all logs
ls -la .devpipe/runs/<run-id>/logs/
```

### Check artifacts

```bash
# Build artifact
cat artifacts/build/app.txt

# JUnit XML
cat artifacts/test/junit.xml
```

## Automated Test Details

### test-fail-fast

This test:
1. Builds devpipe
2. Runs with `DEVPIPE_TEST_FAIL=format` and `--fail-fast`
3. Verifies exit code is 1
4. Parses run.json to confirm only 2 stages ran
5. Passes if pipeline stopped after format failure

### test-continue-on-fail

This test:
1. Builds devpipe
2. Runs with `DEVPIPE_TEST_FAIL=format` (no --fail-fast)
3. Verifies exit code is 1
4. Parses run.json to confirm all 6 stages ran
5. Passes if pipeline continued despite format failure

## Expected run.json Structure

### Success case

```json
{
  "stages": [
    {"id": "lint", "status": "PASS", "exitCode": 0},
    {"id": "format", "status": "PASS", "exitCode": 0},
    {"id": "type-check", "status": "PASS", "exitCode": 0},
    {"id": "build", "status": "PASS", "exitCode": 0},
    {"id": "unit-tests", "status": "PASS", "exitCode": 0},
    {"id": "e2e-tests", "status": "PASS", "exitCode": 0}
  ]
}
```

### Fail-fast case

```json
{
  "flags": {"failFast": true},
  "stages": [
    {"id": "lint", "status": "PASS", "exitCode": 0},
    {"id": "format", "status": "FAIL", "exitCode": 1}
  ]
}
```

### Continue-on-fail case

```json
{
  "flags": {"failFast": false},
  "stages": [
    {"id": "lint", "status": "PASS", "exitCode": 0},
    {"id": "format", "status": "FAIL", "exitCode": 1},
    {"id": "type-check", "status": "PASS", "exitCode": 0},
    {"id": "build", "status": "PASS", "exitCode": 0},
    {"id": "unit-tests", "status": "PASS", "exitCode": 0},
    {"id": "e2e-tests", "status": "PASS", "exitCode": 0}
  ]
}
```

## Regression Testing

Before committing changes, run:

```bash
# Build
make build

# Success scenarios
make demo
make demo-verbose
make demo-fast
make demo-dry-run
make demo-only
make demo-skip

# Failure scenarios
make test-failures

# Verify output
make show-latest
```

All commands should succeed (exit code 0) except the failure tests, which should exit 1 but still pass the test assertions.

## Known Limitations (Iteration 1)

- No actual linting/testing tools (uses hello-world.sh)
- No TOML config (hardcoded stages)
- No TUI (plain text output)
- No metrics parsing
- No HTML dashboard
- No autofix

These will be addressed in future iterations.
