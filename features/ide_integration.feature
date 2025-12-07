@wip
Feature: IDE Integration
  As a devpipe user
  I want IDE integration features
  So that I can navigate from reports to code

  @wip
  Scenario: IDE link generation in reports
    Given a config with tasks that generate reports
    When I run devpipe with that config
    Then the report should contain IDE links
    And links should point to correct file locations

  @wip
  Scenario: Click-to-file functionality
    Given a report with IDE links
    When I click on a file link in the report
    Then the file should open in my IDE
    And the cursor should be at the correct line

  @wip
  Scenario: IDE-specific output formats
    Given a config with IDE integration enabled
    When I run devpipe with that config
    Then the output should include IDE-compatible paths
    And error messages should be IDE-parseable
