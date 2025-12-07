Feature: Unknown and Invalid Commands
  As a devpipe user
  I want helpful error messages for invalid commands
  So that I can quickly correct my mistakes

  Scenario: Unknown subcommand shows error
    When I run devpipe with subcommand "invalid-command"
    Then the execution should fail
    And the output should show available commands
    And the output should indicate unknown command

  Scenario: No subcommand runs default pipeline
    Given a config file with a simple echo task
    When I run devpipe without any subcommand
    Then the execution should succeed
    And the default pipeline should run
    And the task should execute

  Scenario: Typo in subcommand suggests correction
    When I run devpipe with subcommand "lst"
    Then the execution should fail
    And the output should suggest "list" as correction
