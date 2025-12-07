Feature: Phase Execution
  As a devpipe user
  I want tasks to execute in sequential phases
  So that I can control execution order and dependencies

  Scenario: Sequential phase execution
    Given a config with three sequential phases
    When I run devpipe with that config
    Then phase 1 should complete before phase 2 starts
    And phase 2 should complete before phase 3 starts
    And all three phases should execute sequentially

  Scenario: Phase execution stops on failure with fail-fast
    Given a config with three phases where phase 2 fails
    When I run devpipe with --fail-fast
    Then phase 1 tasks should complete
    And phase 2 tasks should fail
    And phase 3 tasks should not execute

  Scenario: Phase with no tasks is skipped
    Given a config with a phase header but no tasks in that phase
    When I run devpipe with that config
    Then the empty phase should be skipped
    And other phases should execute normally
