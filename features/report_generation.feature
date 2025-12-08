@wip
Feature: Report Generation Edge Cases
  As a devpipe user
  I want robust report generation
  So that reports are always available

  @wip
  Scenario: Dashboard with no runs
    Given a fresh devpipe installation with no runs
    When I run devpipe generate-reports
    Then the command should handle no runs gracefully
    And a message should indicate no runs to report

  @wip
  Scenario: Dashboard with corrupted run data
    Given a runs directory with corrupted JSON files
    When I run devpipe generate-reports
    Then corrupted runs should be skipped
    And valid runs should still be processed

  @wip
  Scenario: Dashboard with missing output files
    Given runs that reference missing output files
    When I run devpipe generate-reports
    Then the report should handle missing files gracefully
    And a warning should indicate missing output

  @wip
  Scenario: Dashboard regeneration after config change
    Given existing reports with old config
    And a modified config file
    When I run devpipe generate-reports
    Then reports should be regenerated with new config
    And old reports should be updated
