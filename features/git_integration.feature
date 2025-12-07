@wip
Feature: Git Integration Edge Cases
  As a devpipe user
  I want git integration to handle edge cases
  So that devpipe works in various git scenarios

  @wip
  Scenario: Git mode with no git repo
    Given a directory with no git repository
    And a config with git mode enabled
    When I run devpipe with that config
    Then the execution should fail gracefully
    And the output should indicate no git repo found

  @wip
  Scenario: Git mode with uncommitted changes
    Given a git repo with uncommitted changes
    And a config with git mode = "staged_unstaged"
    When I run devpipe with that config
    Then tasks should run for both staged and unstaged files
    And the output should show which files triggered tasks

  @wip
  Scenario: Git mode with merge conflicts
    Given a git repo with merge conflicts
    And a config with git mode enabled
    When I run devpipe with that config
    Then the execution should handle conflicts gracefully
    And conflicted files should be detected

  @wip
  Scenario: Submodule handling
    Given a git repo with submodules
    And a config with git mode enabled
    When I run devpipe with that config
    Then changes in submodules should be detected
    And tasks should run for submodule changes
