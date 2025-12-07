Feature: DevPipe Commands
  As a devpipe user
  I want to use various devpipe commands
  So that I can interact with my pipeline configuration

  Scenario: List command shows all task IDs
    Given a config with multiple tasks
    When I run devpipe list
    Then the execution should succeed
    And the output should contain all task IDs

  Scenario: Validate command with valid config
    Given a valid config file for validation
    When I run devpipe validate command
    Then the execution should succeed
    And the output should indicate validation passed

  Scenario: Validate command with invalid config
    Given an invalid config file with missing required field
    When I run devpipe validate command
    Then the execution should fail
    And the output should show validation errors

  Scenario: Validate command with multiple files
    Given multiple config files with mixed validity
    When I run devpipe validate with multiple files
    Then the execution should fail
    And the output should show which files failed

  Scenario: Validate command with nonexistent file
    Given no config file exists for validation
    When I run devpipe validate with nonexistent file
    Then the execution should fail
    And the output should indicate file not found

  Scenario: Validate command without arguments uses default config
    Given a valid default config file
    When I run devpipe validate without arguments
    Then the execution should succeed
    And the output should indicate validation passed

  Scenario: SARIF command displays security findings
    Given a SARIF file with security findings
    When I run devpipe sarif with the file
    Then the execution should succeed
    And the output should display the findings

  Scenario: SARIF command with summary flag
    Given a SARIF file with multiple findings
    When I run devpipe sarif with summary flag
    Then the execution should succeed
    And the output should show grouped summary

  Scenario: SARIF command with verbose flag
    Given a SARIF file with security findings
    When I run devpipe sarif with verbose flag
    Then the execution should succeed
    And the output should show detailed metadata

  Scenario: SARIF command with nonexistent file
    Given no SARIF file exists
    When I run devpipe sarif with nonexistent file
    Then the execution should fail
    And the output should indicate SARIF file not found

  Scenario: Generate-reports command regenerates all reports
    Given a config with existing run reports
    When I run devpipe generate-reports
    Then the execution should succeed
    And the output should indicate reports were regenerated
    And the output should show the number of runs processed

  Scenario: Generate-reports command with missing config
    Given no config file exists
    When I run devpipe generate-reports
    Then the execution should fail
    And the output should indicate config error

  Scenario: Generate-reports command with no runs directory
    Given a valid config but no runs directory
    When I run devpipe generate-reports
    Then the execution should fail
    And the output should indicate runs directory error
