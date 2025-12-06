Feature: Task Execution
  As a devpipe user
  I want to run tasks defined in my configuration
  So that I can execute my development pipeline

  Scenario: Run with --only flag
    Given a config file with multiple tasks
    When I run devpipe with --only flag for one task
    Then the execution should succeed
    And only the specified task should run
    And other tasks should not appear in output
