@wip
Feature: Flag Combinations
  As a devpipe user
  I want to combine multiple flags
  So that I can customize execution behavior

  @wip
  Scenario: Fail-fast with dashboard
    Given a config with multiple tasks where one fails
    When I run devpipe with --fail-fast and --dashboard
    Then the dashboard should show live progress
    And execution should stop on first failure
    And the dashboard should show the failure

  @wip
  Scenario: Since with only flag
    Given a config with multiple tasks
    And git changes affecting task-a
    When I run devpipe with --since "HEAD~1" and --only "task-a,task-b"
    Then only task-a should run
    And task-b should be skipped (not affected by changes)

  @wip
  Scenario: Dashboard with no-color
    Given a config with multiple tasks
    When I run devpipe with --dashboard and --no-color
    Then the dashboard should display without ANSI colors
    And progress should still be visible

  @wip
  Scenario: Fail-fast with verbose
    Given a config with multiple tasks where one fails
    When I run devpipe with --fail-fast and --verbose
    Then verbose logging should be enabled
    And execution should stop on first failure
    And the output should show detailed failure information

  @wip
  Scenario: Since with fast flag
    Given a config with fast and slow tasks
    And git changes affecting both types
    When I run devpipe with --since "HEAD~1" and --fast
    Then only fast tasks affected by changes should run
    And slow tasks should be skipped
