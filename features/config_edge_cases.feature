Feature: Config Edge Cases
  As a devpipe user
  I want robust config handling
  So that edge cases are handled gracefully

  Scenario: Multiple config files validation
    Given multiple config files with different tasks
    When I run devpipe validate on all config files
    Then each config should be validated independently
    And validation results should be shown for each

  Scenario: Config with only phase headers
    Given a config with phase headers but no tasks
    When I run devpipe with that config
    Then the execution should succeed
    And a warning should indicate no tasks to run

  Scenario: Config with very large task count
    Given a config with 100+ tasks
    When I run devpipe list with that config
    Then all tasks should be listed
    And performance should be acceptable

  Scenario: Empty task IDs or special characters
    Given a config with tasks having special characters in IDs
    When I run devpipe with the special characters config
    Then tasks should execute correctly
    And output should display task IDs properly
