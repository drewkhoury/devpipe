@wip
Feature: TTY Detection and Output
  As a devpipe user
  I want appropriate output based on terminal type
  So that output works in all environments

  @wip
  Scenario: Non-TTY output piped to file
    Given a config with multiple tasks
    When I run devpipe and pipe output to a file
    Then the output should be plain text without ANSI codes
    And progress indicators should be simplified

  @wip
  Scenario: TTY vs non-TTY behavior differences
    Given a config with multiple tasks
    When I run devpipe in TTY mode
    Then interactive progress should be shown
    When I run devpipe in non-TTY mode
    Then static progress should be shown

  @wip
  Scenario: Color output in non-TTY environments
    Given a config with multiple tasks
    When I run devpipe in non-TTY mode
    Then colors should be disabled automatically
    And output should remain readable

  @wip
  Scenario: Progress bars in non-TTY mode
    Given a config with multiple tasks
    When I run devpipe in non-TTY mode
    Then progress bars should not be rendered
    And progress should be shown as text updates
