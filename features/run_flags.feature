Feature: Run Command Flags
  As a devpipe user
  I want to use command-line flags to control execution
  So that I can customize pipeline behavior

  Scenario: Run with --only flag filters tasks
    Given a config with tasks "task-a,task-b,task-c"
    When I run devpipe with --only "task-a"
    Then the execution should succeed
    And only "task-a" should run
    And "task-b,task-c" should not run

  Scenario: Run with --dry-run simulates execution
    Given a config with a task that writes a file
    When I run devpipe with --dry-run
    Then the execution should succeed
    And the file should not exist

  Scenario: Run with --verbose shows detailed output
    Given a config with a simple task
    When I run devpipe with --verbose
    Then the execution should succeed
    And the output should show verbose details

  Scenario: Run with --fast skips slow tasks
    Given a config with fast and slow tasks
    When I run devpipe with --fast
    Then the execution should succeed
    And slow tasks should be skipped
    And fast tasks should run

  Scenario: Run with --ui basic uses basic mode
    Given a config with a simple task
    When I run devpipe with --ui "basic"
    Then the execution should succeed
    And the output should not contain animation characters

  Scenario: Run with --ui full uses full mode
    Given a config with a simple task
    When I run devpipe with --ui "full"
    Then the execution should succeed

  Scenario: Run with --no-color disables colors
    Given a config with a simple task
    When I run devpipe with --no-color
    Then the execution should succeed
    And the output should not contain ANSI color codes

  Scenario: Run with custom --config path
    Given a config file at a custom path
    When I run devpipe with --config pointing to that path
    Then the execution should succeed
    And tasks from the custom config should run

  Scenario: Run with --fix-type auto enables auto-fix
    Given a config with a fixable task
    When I run devpipe with --fix-type "auto"
    Then the execution should succeed

  Scenario: Run with --fix-type none disables fixes
    Given a config with a fixable task
    When I run devpipe with --fix-type "none"
    Then the execution should succeed

  Scenario: Run with multiple --skip flags
    Given a config with tasks "one,two,three,four"
    When I run devpipe with multiple --skip flags for "two,three"
    Then the execution should succeed
    And "one,four" should run
    And "two,three" should not run
