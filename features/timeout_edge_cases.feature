@wip
Feature: Timeout Edge Cases
  As a devpipe user
  I want robust timeout handling
  So that tasks don't hang indefinitely

  @wip
  Scenario: Zero timeout behavior
    Given a config with a task that has timeout = 0
    When I run devpipe with that config
    Then the task should run without timeout
    And the task should complete normally

  @wip
  Scenario: Timeout inheritance from defaults
    Given a config with defaults.timeout = 30
    And a task with no explicit timeout
    When I run devpipe with that config
    Then the task should inherit the default timeout
    And timeout should be enforced

  @wip
  Scenario: Timeout with very long-running tasks
    Given a config with a task that takes 5 minutes
    And the task has timeout = 600
    When I run devpipe with that config
    Then the task should complete successfully
    And no timeout error should occur
