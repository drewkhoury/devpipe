@wip
Feature: Phase Execution
  As a devpipe user
  I want tasks to execute in sequential phases
  So that I can control execution order and dependencies

  @wip
  Scenario: Sequential phase execution
    Given a config with three phases
    When I run devpipe with that config
    Then phase 1 should complete before phase 2 starts
    And phase 2 should complete before phase 3 starts
    And all phases should execute in order

  @wip
  Scenario: Phase execution stops on failure with fail-fast
    Given a config with three phases where phase 2 fails
    When I run devpipe with --fail-fast
    Then phase 1 should complete
    And phase 2 should fail
    And phase 3 should not execute

  @wip
  Scenario: Phase with no tasks is skipped
    Given a config with a phase header but no tasks in that phase
    When I run devpipe with that config
    Then the empty phase should be skipped
    And other phases should execute normally
