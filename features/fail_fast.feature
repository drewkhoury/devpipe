Feature: Fail-Fast Execution
  As a devpipe user
  I want to stop execution on first failure
  So that I can save time and get quick feedback

  Scenario: Fail-fast stops on first task failure
    Given a config with three tasks where the second fails
    When I run devpipe with --fail-fast
    Then the execution should fail
    And the first task should have run
    And the second task should have failed
    And the third task should not have run

  Scenario: Fail-fast with multiple failing tasks stops at first
    Given a config with three tasks where second and third fail
    When I run devpipe with --fail-fast
    Then the execution should fail
    And only the first two tasks should have run
    And the third task should not have run

  Scenario: Fail-fast with all passing tasks completes normally
    Given a config with three passing tasks
    When I run devpipe with --fail-fast
    Then the execution should succeed
    And all three tasks should have run

  Scenario: Without fail-fast all tasks run despite failures
    Given a config with three tasks where the second fails
    When I run devpipe without fail-fast
    Then the execution should fail
    And all three tasks should have run
