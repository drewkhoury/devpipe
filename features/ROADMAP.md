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

## Statistics

- **Current coverage:** 75 scenarios
- **Planned (high priority):** ~12-13 scenarios
- **Target coverage:** ~87-90 scenarios
- **Stretch goal:** ~100-110 scenarios

---

## Notes

- Remove `@wip` tag when scenario is fully implemented and passing
- Keep feature files even if unimplemented - they serve as documentation
- Update this roadmap as priorities change
