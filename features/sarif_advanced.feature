@wip
Feature: SARIF Advanced Features
  As a devpipe user
  I want advanced SARIF analysis capabilities
  So that I can understand complex security findings

  @wip
  Scenario: SARIF dataflow analysis display
    Given a SARIF file with dataflow information
    When I run devpipe sarif with that file
    Then dataflow paths should be displayed
    And the flow from source to sink should be clear

  @wip
  Scenario: SARIF with multiple tools in single file
    Given a SARIF file with results from multiple tools
    When I run devpipe sarif with that file
    Then results should be grouped by tool
    And each tool's findings should be clearly separated

  @wip
  Scenario: SARIF severity filtering
    Given a SARIF file with mixed severity findings
    When I run devpipe sarif with severity filter
    Then only findings matching the filter should be shown
    And the count should reflect filtered results

  @wip
  Scenario: SARIF with very large result sets
    Given a SARIF file with 100+ findings
    When I run devpipe sarif with that file
    Then all findings should be processed
    And performance should remain acceptable
    And the output should be paginated or summarized
