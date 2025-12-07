@wip
Feature: Dashboard Advanced Features
  As a devpipe user
  I want advanced dashboard capabilities
  So that I can monitor complex pipelines

  @wip
  Scenario: Dashboard with parallel tasks showing concurrent progress
    Given a config with 10 tasks running in parallel
    When I run devpipe with --dashboard
    Then the dashboard should show all tasks concurrently
    And progress should update for each task independently

  @wip
  Scenario: Dashboard refresh rate behavior
    Given a config with tasks of varying durations
    When I run devpipe with --dashboard
    Then the dashboard should refresh at configured rate
    And updates should be smooth without flickering

  @wip
  Scenario: Dashboard output truncation for long logs
    Given a config with a task generating very long output
    When I run devpipe with --dashboard
    Then the dashboard should truncate long output
    And the full output should be available in log files

  @wip
  Scenario: Dashboard with phases
    Given a config with multiple phases
    When I run devpipe with --dashboard
    Then the dashboard should show current phase
    And phase progress should be displayed
