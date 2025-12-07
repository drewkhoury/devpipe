@wip
Feature: Logging and Reports
  As a devpipe user
  I want reliable logging and report generation
  So that I can review execution history

  @wip
  Scenario: Report generation failures
    Given a config with tasks that have completed
    And the report generation encounters an error
    When I run devpipe generate-reports
    Then the error should be reported clearly
    And partial reports should still be generated if possible

  @wip
  Scenario: Missing output directory handling
    Given a config with outputRoot pointing to nonexistent directory
    When I run devpipe with that config
    Then the output directory should be created automatically
    And logs should be written successfully

  @wip
  Scenario: Permissions issues on log files
    Given a config with outputRoot in a read-only directory
    When I run devpipe with that config
    Then the permission error should be caught
    And a helpful error message should be shown

  @wip
  Scenario: Report with very large output
    Given a config with tasks generating large output (10MB+)
    When I run devpipe with that config
    Then the report should handle large output gracefully
    And the report should remain performant
