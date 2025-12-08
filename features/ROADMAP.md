# BDD Test Roadmap

This document tracks planned but unimplemented BDD test scenarios.

## How It Works

We use the `@wip` (work-in-progress) tag to mark scenarios that are:
- Planned but not yet implemented
- Being actively developed
- Waiting for feature implementation

### Tag Usage

```gherkin
@wip
Feature: Dashboard Flag
  
  @wip
  Scenario: Dashboard shows live progress
    Given a config with multiple tasks
    When I run devpipe with --dashboard
    Then I should see live progress updates
```

### Running Tests

**Normal test run (excludes @wip):**
```bash
./devpipe --only bdd-tests
# or
make test-bdd
```

**View @wip scenarios (they'll show as undefined):**
```bash
go test -v ./features -run TestFeatures
# Look for "undefined" steps
```

### Workflow

1. **Planning:** Write feature file with `@wip` tag
2. **Implementation:** Create `*_test.go` file with step definitions
3. **Testing:** Scenarios will show as "undefined" until steps are implemented
4. **Completion:** Remove `@wip` tag when all steps pass

### Creating a New Roadmap Feature

**Step 1:** Create the feature file
```bash
# Example: features/since_flag_roadmap.feature
```

**Step 2:** Add @wip tags
```gherkin
@wip
Feature: Since Flag
  As a devpipe user
  I want to filter tasks by git changes
  
  @wip
  Scenario: Since overrides config
    Given a config with git.ref = "main"
    When I run devpipe with --since "develop"
    Then tasks should filter based on develop
```

**Step 3:** Update this roadmap
- Add to appropriate priority section
- Update statistics
- Mark as planned

**Step 4:** When ready to implement
- Create `features/since_flag_test.go`
- Implement step definitions
- Register in `basic_test.go`: `InitializeSinceFlagScenario(sc, shared)`
- Run tests, remove `@wip` tags when passing
- Update roadmap to mark as complete

### Reference: Existing Feature Files

- `basic.feature` - Basic pipeline execution
- `commands.feature` - CLI commands (list, validate, sarif, etc.)
- `config_validation.feature` - Config validation errors/warnings
- `error_scenarios.feature` - Error handling
- `run_flags.feature` - Run command flags
- `fail_fast.feature` - Fail-fast execution control
- `advanced_features.feature` - JUnit, SARIF, auto-fix, workdir
- `task_execution.feature` - Task execution and phases

---

## Current Roadmap

### High Priority (0 scenarios implemented)

#### 1. Dashboard Flag - 4 scenarios
**File:** `dashboard_roadmap.feature`
- [ ] Dashboard shows live task progress
- [ ] Dashboard displays task failures  
- [ ] Dashboard with no-color flag
- [ ] Dashboard updates on long-running tasks

**Estimated effort:** ~150-180 lines

---

#### 2. Since Flag - 3 scenarios (not yet created)
**Planned scenarios:**
- [ ] Since flag overrides config git.ref
- [ ] Since with valid git ref filters tasks
- [ ] Since with invalid ref shows error

**Estimated effort:** ~100-120 lines

---

#### 3. Unknown/Invalid Commands - 3 scenarios (not yet created)
**Planned scenarios:**
- [ ] Unknown subcommand shows error
- [ ] No subcommand shows usage
- [ ] Typo in subcommand (if suggestion implemented)

**Estimated effort:** ~80-100 lines

---

#### 4. Phase Execution - 2 scenarios (commented in advanced_features.feature)
**Existing but disabled:**
```gherkin
# TODO: Phase execution tests need git mode handling fixes
# Scenario: Sequential phase execution
# Scenario: Phase execution stops on failure with fail-fast
```

**Estimated effort:** ~100-150 lines (fix + enhance)

---

### Medium Priority - Edge Cases & Combinations

#### 5. Flag Combinations - 4-5 scenarios
**Missing combinations:**
- [ ] `--fail-fast` + `--dashboard`
- [ ] `--since` + `--only`
- [ ] `--dashboard` + `--no-color`
- [ ] `--fail-fast` + `--verbose`
- [ ] `--since` + `--fast`

**Estimated effort:** ~120-150 lines

---

#### 6. Config Edge Cases - 3-4 scenarios
**Planned scenarios:**
- [ ] Multiple config files validation
- [ ] Config with only phase headers (no tasks)
- [ ] Config with very large task count (100+ tasks)
- [ ] Empty task IDs or special characters in names

**Estimated effort:** ~100-120 lines

---

#### 7. Git Integration Edge Cases - 3-4 scenarios
**Planned scenarios:**
- [ ] Git mode with no git repo
- [ ] Git mode with uncommitted changes
- [ ] Git mode with merge conflicts
- [ ] Submodule handling

**Estimated effort:** ~100-120 lines

---

#### 8. Artifact Metrics - 2-3 scenarios
**Planned scenarios:**
- [ ] Artifact collection with outputType="artifact"
- [ ] Artifact path validation
- [ ] Multiple artifact formats

**Estimated effort:** ~80-100 lines

---

### Lower Priority - Advanced Features

#### 9. Timeout Edge Cases - 2-3 scenarios
**Currently tested:** Task exceeds timeout, very short timeout

**Missing:**
- [ ] Zero timeout behavior
- [ ] Timeout inheritance from defaults
- [ ] Timeout with very long-running tasks (minutes)

**Estimated effort:** ~80-100 lines

---

#### 10. Concurrency Edge Cases - 2-3 scenarios
**Planned scenarios:**
- [ ] Maximum parallel tasks in single phase
- [ ] Phase with single task
- [ ] Empty phase handling
- [ ] Task spawn failures

**Estimated effort:** ~80-100 lines

---

#### 11. UI & Output Edge Cases - 2-3 scenarios
**Currently tested:** Basic UI modes, no-color

**Missing:**
- [ ] UI mode with very long task names (truncation)
- [ ] UI mode with special characters
- [ ] Progress bar edge cases
- [ ] Output with mixed success/failure states

**Estimated effort:** ~80-100 lines

---

#### 12. Fix Workflow Edge Cases - 3-4 scenarios
**Currently tested:** Auto-fix success, helper-fix suggestion

**Missing:**
- [ ] Fix command fails
- [ ] Fix command times out
- [ ] Multiple fix attempts
- [ ] Fix with `--dry-run` flag
- [ ] Fix type inheritance from task_defaults

**Estimated effort:** ~100-120 lines

---

#### 13. Logging & Reports - 2-3 scenarios
**Planned scenarios:**
- [ ] Report generation failures
- [ ] Missing output directory handling
- [ ] Permissions issues on log files
- [ ] Report with very large output

**Estimated effort:** ~80-100 lines

---

#### 14. Dashboard Advanced Features - 2-3 scenarios
**Beyond basic dashboard tests:**
- [ ] Dashboard with parallel tasks showing concurrent progress
- [ ] Dashboard refresh rate behavior
- [ ] Dashboard output truncation for long logs
- [ ] Dashboard with phases

**Estimated effort:** ~80-100 lines

---

#### 15. IDE Integration - 2-3 scenarios
**Planned scenarios:**
- [ ] IDE link generation in reports
- [ ] Click-to-file functionality
- [ ] IDE-specific output formats

**Estimated effort:** ~60-80 lines

---

#### 16. Animated UI Mode - 2-3 scenarios
**Planned scenarios:**
- [ ] Animated progress with type grouping
- [ ] Animated progress with phase grouping
- [ ] Animation refresh rate configuration
- [ ] Animation with very fast tasks

**Estimated effort:** ~80-100 lines

---

#### 17. TTY Detection & Output - 2-3 scenarios
**Planned scenarios:**
- [ ] Non-TTY output (piped to file)
- [ ] TTY vs non-TTY behavior differences
- [ ] Color output in non-TTY environments
- [ ] Progress bars in non-TTY mode

**Estimated effort:** ~80-100 lines

---

#### 18. Report Generation Edge Cases - 2-3 scenarios
**Planned scenarios:**
- [ ] Dashboard with no runs
- [ ] Dashboard with corrupted run data
- [ ] Dashboard with missing metrics files
- [ ] Dashboard regeneration after config change

**Estimated effort:** ~80-100 lines

---

#### 19. SARIF Advanced Features - 2-3 scenarios
**Currently tested:** Basic SARIF display, directory scan

**Missing:**
- [ ] SARIF dataflow analysis display
- [ ] SARIF with multiple tools in single file
- [ ] SARIF severity filtering
- [ ] SARIF with very large result sets (100+ findings)

**Estimated effort:** ~80-100 lines

---

## Statistics

- **Current coverage:** 75 scenarios (311 steps)
- **High priority planned:** ~12-13 scenarios
- **Medium priority planned:** ~20-25 scenarios
- **Lower priority planned:** ~25-30 scenarios (expanded from 15-20)
- **Total potential:** ~120-135 scenarios (expanded from 110-120)
- **Target (high priority):** ~87-90 scenarios
- **Realistic goal (high + medium):** ~100-110 scenarios
- **Comprehensive coverage (all priorities):** ~120-135 scenarios

---

## Implementation Priority

### Phase 1: High Priority (Next 4-6 sessions)
Focus on major features with zero coverage:
1. Dashboard (4 scenarios)
2. Since flag (3 scenarios)
3. Unknown commands (3 scenarios)
4. Phase execution fixes (2-3 scenarios)

**Target:** 87-90 scenarios

### Phase 2: Medium Priority (Optional)
Edge cases and combinations:
5. Flag combinations (4-5 scenarios)
6. Config edge cases (3-4 scenarios)
7. Git integration edge cases (3-4 scenarios)
8. Artifact metrics (2-3 scenarios)

**Target:** 100-105 scenarios

### Phase 3: Lower Priority (Nice to have)
Advanced features and polish:
9-19. Various edge cases and advanced features including:
- Timeout, concurrency, UI/output edge cases
- Fix workflow edge cases
- Logging & reports
- Dashboard advanced features
- IDE integration
- Animated UI mode
- TTY detection
- Report generation edge cases
- SARIF advanced features

**Target:** 120-135 scenarios

---

## Diminishing Returns

**Stop point recommendation:** ~100 scenarios
- Excellent coverage for a CLI tool
- All major features tested
- Most edge cases covered
- Reasonable maintenance burden

Beyond 100 scenarios, testing becomes increasingly niche with diminishing value.

---

## Notes

- Remove `@wip` tag when scenario is fully implemented and passing
- Keep feature files even if unimplemented - they serve as documentation
- Update this roadmap as priorities change
- Focus on Phase 1 (high priority) first before expanding to edge cases
- Each scenario should test real user workflows, not just code paths
