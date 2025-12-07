@wip
Feature: Dashboard Flag
  As a devpipe user
  I want to see live progress during execution
  So that I can monitor long-running pipelines

  @wip
  Scenario: Dashboard shows live task progress
    Given a config with multiple tasks
    When I run devpipe with --dashboard
    Then I should see live progress updates
    And the dashboard should update as tasks complete

  @wip
  Scenario: Dashboard displays task failures
    Given a config with tasks where some fail
    When I run devpipe with --dashboard
    Then the dashboard should show failed tasks in red
    And the dashboard should continue showing remaining tasks

  @wip
  Scenario: Dashboard with no-color flag
    Given a config with multiple tasks
    When I run devpipe with --dashboard and --no-color
    Then the dashboard should display without ANSI colors
    And progress should still be visible

  @wip
  Scenario: Dashboard updates on long-running tasks
    Given a config with a task that takes 10 seconds
    When I run devpipe with --dashboard
    Then the dashboard should show progress during execution
    And the elapsed time should update
