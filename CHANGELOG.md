# devpipe Changelog

## [v0.2.0] - 2025-11-28 - Iteration 2 Complete

### Added
- **TOML Configuration Support** - Define custom stages per project with `config.toml`
- **Git Modes** - Three modes for changed file detection:
  - `staged` - Only staged files (`git diff --cached`)
  - `staged_unstaged` - Staged + unstaged files (`git diff HEAD`)
  - `ref` - Compare against specific ref (`git diff <ref>`)
- **New CLI Flags**:
  - `--config <path>` - Specify custom config file location
  - `--since <ref>` - Override git ref for comparison
- **Config Schema**:
  - `[defaults]` - Global settings (outputRoot, fastThreshold)
  - `[defaults.git]` - Git configuration (mode, ref)
  - `[stage_defaults]` - Default values for all stages
  - `[stages.<id>]` - Per-stage configuration
- **Package Structure** - Organized code into internal packages:
  - `internal/config` - Configuration loading and merging
  - `internal/git` - Git operations and modes
  - `internal/model` - Data structures
- **Enhanced run.json** - Now includes:
  - `configPath` - Path to config file used
  - `git.mode` - Git mode used
  - `git.ref` - Git ref compared against
  - `flags.config` - Config flag value
  - `flags.since` - Since flag value

### Changed
- Stages now configurable via TOML instead of hardcoded
- Git changed file detection now supports multiple modes
- Stage ordering preserved from built-in order when using config
- Backward compatible - works without config file using built-in stages

### Documentation
- Updated README.md with configuration examples
- Added config.toml.example with comprehensive examples
- Documented all three git modes
- Added configuration section to README

## [Unreleased] - Quick Wins Applied

### Fixed
- **Replaced custom multiWriter with stdlib** - Now using `io.MultiWriter` instead of custom implementation (removed 20 lines of code)
- **Fixed deprecated rand.Seed** - Now using `os.Getpid()` for run ID uniqueness instead of deprecated `rand.Seed()`
- **Fixed git diff empty output bug** - Correctly returns empty array `[]` instead of `[""]` when no files changed

### Validation
- ✅ All builds succeed
- ✅ All demo tests pass (demo, demo-fast, demo-only, demo-dry-run)
- ✅ All failure tests pass (test-fail-fast, test-continue-on-fail)
- ✅ Code reduced from 528 to 508 lines

---

## [v0.1.0] - 2025-11-28 - Iteration 1 Complete

### Added
- Single Go binary pipeline runner (stdlib only, no dependencies)
- Hardcoded 6-stage pipeline: lint, format, type-check, build, unit-tests, e2e-tests
- CLI flags: `--only`, `--skip`, `--fast`, `--fail-fast`, `--dry-run`, `--verbose`
- Git integration: repo detection, changed file tracking from HEAD
- Per-run artifacts: `.devpipe/runs/<run-id>/run.json` + `logs/<stage>.log`
- Plain text console output
- hello-world.sh test harness with failure simulation via `DEVPIPE_TEST_FAIL`
- Comprehensive test suite with automated failure tests
- Makefile with build/test/demo commands
- Documentation: README.md, TESTING.md, ROADMAP.md

### Features
- Detects git repo or falls back to CWD
- Tracks changed files with `git diff --name-only HEAD`
- Sequential stage execution with continue-on-failure by default
- `--fail-fast` stops on first failure
- `--fast` skips stages with estimatedSeconds >= 300
- `--only <stage-id>` runs single stage
- `--skip <stage-id>` skips stages (repeatable)
- `--dry-run` simulates without executing
- `--verbose` shows commands and exit codes
- Structured run.json with full metadata
- Per-stage log files with stdout+stderr

### Testing
- hello-world.sh simulates all stage commands
- DEVPIPE_TEST_FAIL env var for failure simulation
- Automated tests verify --fail-fast behavior
- Automated tests verify continue-on-failure behavior
- All success and failure scenarios covered

### Documentation
- README.md with quick start and examples
- TESTING.md with comprehensive test guide
- ROADMAP.md with 5-iteration development plan
- Makefile help command
