@wip
Feature: Concurrency Edge Cases
  As a devpipe user
  I want reliable concurrent task execution
  So that parallel tasks work correctly

  @wip
  Scenario: Maximum parallel tasks in single phase
    Given a config with 20 tasks in the same phase
    When I run devpipe with that config
    Then all tasks should run in parallel
    And all tasks should complete successfully

  @wip
  Scenario: Phase with single task
    Given a config with a phase containing only one task
    When I run devpipe with that config
    Then the single task should execute
    And phase should complete normally

  @wip
  Scenario: Empty phase handling
    Given a config with an empty phase between two populated phases
    When I run devpipe with that config
    Then the empty phase should be skipped
    And execution should proceed to next phase

  @wip
  Scenario: Task spawn failures
    Given a config with a task that fails to spawn
    When I run devpipe with that config
    Then the spawn failure should be caught
    And the error should be reported clearly
