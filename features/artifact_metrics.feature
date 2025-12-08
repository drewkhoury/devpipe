@wip
Feature: Artifact Output
  As a devpipe user
  I want to collect artifact output
  So that I can track build artifacts and outputs

  @wip
  Scenario: Artifact collection with outputType
    Given a config with a task that has outputType = "artifact"
    And the task generates an artifact file
    When I run devpipe with that config
    Then the artifact should be collected
    And the artifact should appear in the report

  @wip
  Scenario: Artifact path validation
    Given a config with invalid artifact path
    When I run devpipe validate with that config
    Then validation should show warning about artifact path
    And the output should suggest correct path format

  @wip
  Scenario: Multiple artifact formats
    Given a config with tasks generating different artifact types
    When I run devpipe with that config
    Then all artifacts should be collected
    And the report should display all artifact types
