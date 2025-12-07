@wip
Feature: Animated UI Mode
  As a devpipe user
  I want animated progress indicators
  So that I can see live task execution

  @wip
  Scenario: Animated progress with type grouping
    Given a config with animatedGroupBy = "type"
    When I run devpipe with that config
    Then tasks should be grouped by type in the animation
    And progress should update for each group

  @wip
  Scenario: Animated progress with phase grouping
    Given a config with animatedGroupBy = "phase"
    When I run devpipe with that config
    Then tasks should be grouped by phase in the animation
    And phase progress should be animated

  @wip
  Scenario: Animation refresh rate configuration
    Given a config with animationRefreshMs = 100
    When I run devpipe with that config
    Then the animation should refresh every 100ms
    And the animation should be smooth

  @wip
  Scenario: Animation with very fast tasks
    Given a config with tasks completing in < 10ms
    When I run devpipe with animated UI
    Then the animation should handle fast tasks gracefully
    And all task completions should be visible
