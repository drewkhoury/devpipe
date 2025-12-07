@wip
Feature: UI and Output Edge Cases
  As a devpipe user
  I want reliable UI output in all scenarios
  So that I can see task progress clearly

  @wip
  Scenario: UI mode with very long task names
    Given a config with tasks having very long names (100+ characters)
    When I run devpipe with --ui full
    Then task names should be truncated appropriately
    And the UI should remain readable

  @wip
  Scenario: UI mode with special characters
    Given a config with tasks having special characters in names
    When I run devpipe with that config
    Then special characters should be displayed correctly
    And the UI should not break

  @wip
  Scenario: Progress bar edge cases
    Given a config with tasks of varying durations
    When I run devpipe with progress bars enabled
    Then progress bars should update smoothly
    And completion percentages should be accurate

  @wip
  Scenario: Output with mixed success and failure states
    Given a config with some passing and some failing tasks
    When I run devpipe with that config
    Then the output should clearly distinguish success from failure
    And the summary should show accurate counts
