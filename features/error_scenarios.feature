Feature: Error Scenarios
  As a devpipe user
  I want proper error handling for various failure conditions
  So that I can understand and fix issues quickly

  Scenario: Malformed TOML config file
    Given a config file with malformed TOML syntax
    When I run devpipe with that config
    Then the execution should fail
    And the output should contain "error"
    And the output should contain "TOML"

  Scenario: Config with missing required field
    Given a config file with a task missing the command field
    When I run devpipe with that config
    Then the execution should fail
    And the output should contain "command"

  Scenario: Config with missing task name uses ID as fallback
    Given a config file with a task missing the name field
    When I run devpipe with that config
    Then the execution should succeed
    And the output should contain "broken-task"

  Scenario: Task command exits with error code
    Given a config with a task that exits with code 1
    When I run devpipe with that config
    Then the execution should fail
    And the output should indicate task failure

  Scenario: Task command exits with specific error code
    Given a config with a task that exits with code 42
    When I run devpipe with that config
    Then the execution should fail
    And the exit code should be non-zero

  Scenario: Long-running task exceeds timeout
    Given a config with a task that has a 2 second timeout
    And the task sleeps for 5 seconds
    When I run devpipe with that config
    Then the execution should fail
    And the output should contain "timeout"

  Scenario: Task with very short timeout
    Given a config with a task that has a 1 second timeout
    And the task sleeps for 3 seconds
    When I run devpipe with that config
    Then the execution should fail
    And the output should contain "timeout"

  Scenario: Command binary does not exist
    Given a config with a task using a non-existent binary
    When I run devpipe with that config
    Then the execution should fail
    And the output should indicate command not found

  Scenario: Command binary not in PATH
    Given a config with a task using an invalid command path
    When I run devpipe with that config
    Then the execution should fail
    And the output should indicate command not found

  Scenario: Multiple tasks fail in sequence
    Given a config with three tasks where the second and third fail
    When I run devpipe with that config
    Then the execution should fail
    And the output should show the first task succeeded
    And the output should show the second task failed
    And the output should show the third task failed

  Scenario: Task with invalid working directory
    Given a config with a task that has an invalid working directory
    When I run devpipe with that config
    Then the execution should fail
    And the output should indicate directory error

  Scenario: Config file does not exist
    Given a non-existent config file path
    When I run devpipe with that config path
    Then the execution should fail
    And the output should indicate config file not found

  Scenario: Empty config file
    Given an empty config file
    When I run devpipe with that config
    Then the execution should fail
    And the output should indicate configuration error

  Scenario: Task with empty command
    Given a config with a task that has an empty command
    When I run devpipe with that config
    Then the execution should fail
    And the output should indicate invalid command
