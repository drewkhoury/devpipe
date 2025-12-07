@wip
Feature: Unknown and Invalid Commands
  As a devpipe user
  I want helpful error messages for invalid commands
  So that I can quickly correct my mistakes

  @wip
  Scenario: Unknown subcommand shows error
    When I run devpipe with subcommand "invalid-command"
    Then the execution should fail
    And the output should show available commands
    And the output should indicate unknown command

  @wip
  Scenario: No subcommand shows usage
    When I run devpipe without any subcommand
    Then the execution should succeed
    And the output should show usage information
    And the output should list available commands

  @wip
  Scenario: Typo in subcommand suggests correction
    When I run devpipe with subcommand "lst"
    Then the execution should fail
    And the output should suggest "list" as correction
