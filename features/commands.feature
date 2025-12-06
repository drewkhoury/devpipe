Feature: DevPipe Commands
  As a devpipe user
  I want to use various devpipe commands
  So that I can interact with my pipeline configuration

  Scenario: List command shows all task IDs
    Given a config with multiple tasks
    When I run devpipe list
    Then the execution should succeed
    And the output should contain all task IDs
