Feature: Advanced DevPipe Features
  As a devpipe user
  I want to use advanced pipeline features
  So that I can build sophisticated CI/CD workflows

  Scenario: JUnit XML output collection
    Given a config with a task that generates JUnit XML
    When I run devpipe with that config
    Then the execution should succeed
    And the JUnit output should be parsed
    And the output should show test counts

  Scenario: SARIF output collection
    Given a config with a task that generates SARIF output
    When I run devpipe with that config
    Then the execution should succeed
    And the SARIF output should be parsed
    And the output should show security findings

  Scenario: Auto-fix workflow for failing task
    Given a config with a fixable task that fails
    And the task has fixType set to "auto"
    When I run devpipe with that config
    Then the fix command should be executed automatically
    And the task should be rechecked
    And the execution should succeed after fix

  Scenario: Helper fix workflow shows fix command
    Given a config with a fixable task that fails
    And the task has fixType set to "helper"
    When I run devpipe with that config
    Then the execution should fail
    And the output should show the fix command suggestion

  # TODO: Phase execution tests need git mode handling fixes
  # Scenario: Sequential phase execution
  #   Given a config with three phases
  #   When I run devpipe with that config
  #   Then phase 1 should complete before phase 2 starts
  #   And phase 2 should complete before phase 3 starts
  #   And all phases should execute in order

  Scenario: Parallel execution within a phase
    Given a config with multiple tasks in the same phase
    When I run devpipe with that config
    Then the execution should succeed
    And tasks should run concurrently within the phase

  # TODO: Phase execution tests need git mode handling fixes
  # Scenario: Phase execution stops on failure with fail-fast
  #   Given a config with three phases where phase 2 fails
  #   When I run devpipe with --fail-fast
  #   Then phase 1 should complete
  #   And phase 2 should fail
  #   And phase 3 should not execute

  Scenario: Task with custom working directory
    Given a config with a task that has a custom workdir
    When I run devpipe with that config
    Then the execution should succeed
    And the task should run in the specified directory

  Scenario: Task with relative workdir path
    Given a config with a task using a relative workdir
    When I run devpipe with that config
    Then the execution should succeed
    And the task should run in the relative directory

  Scenario: Multiple tasks with different working directories
    Given a config with tasks in different workdirs
    When I run devpipe with that config
    Then the execution should succeed
    And each task should run in its own directory

  # TODO: Environment variable support not yet implemented in config schema
  # Scenario: Task with environment variables
  #   Given a config with a task that sets environment variables
  #   When I run devpipe with that config
  #   Then the execution should succeed
  #   And the task should have access to the environment variables

  # TODO: Environment variable support not yet implemented in config schema
  # Scenario: Task inherits default environment variables
  #   Given a config with default environment variables
  #   And a task that uses those variables
  #   When I run devpipe with that config
  #   Then the execution should succeed
  #   And the task should see the default environment variables

  # TODO: Environment variable support not yet implemented in config schema
  # Scenario: Task-specific environment variables override defaults
  #   Given a config with default environment variables
  #   And a task that overrides those variables
  #   When I run devpipe with that config
  #   Then the execution should succeed
  #   And the task should see the overridden values
