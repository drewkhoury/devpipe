Feature: Config Validation
  As a devpipe user
  I want comprehensive config validation
  So that I can catch configuration errors early

  Scenario: Config with invalid UI mode
    Given a config with invalid UI mode
    When I run devpipe validate with that config
    Then the execution should fail
    And the output should indicate invalid UI mode

  Scenario: Config with invalid animated group-by
    Given a config with invalid animated group-by
    When I run devpipe validate with that config
    Then the execution should fail
    And the output should indicate invalid group-by

  Scenario: Config with negative fast threshold
    Given a config with negative fast threshold
    When I run devpipe validate with that config
    Then the execution should fail
    And the output should indicate negative threshold

  Scenario: Config with negative animation refresh rate
    Given a config with negative animation refresh rate
    When I run devpipe validate with that config
    Then the execution should fail
    And the output should indicate negative refresh rate

  Scenario: Config with invalid git mode
    Given a config with invalid git mode
    When I run devpipe validate with that config
    Then the execution should fail
    And the output should indicate invalid git mode

  Scenario: Config with invalid task-level fix type
    Given a config with invalid task-level fix type
    When I run devpipe validate with that config
    Then the execution should fail
    And the output should indicate invalid fix type

  Scenario: Config with fix type but missing fix command
    Given a config with fix type but missing fix command
    When I run devpipe validate with that config
    Then the execution should fail
    And the output should indicate missing fix command

  Scenario: Config with git ref mode but no ref specified
    Given a config with git ref mode but no ref specified
    When I run devpipe validate with that config
    Then the execution should succeed
    And the output should show validation warning about missing ref

  Scenario: Config with output type but no output path
    Given a config with output type but no output path
    When I run devpipe validate with that config
    Then the execution should succeed
    And the output should show validation warning about missing output path

  Scenario: Config with output path but no output type
    Given a config with output path but no output type
    When I run devpipe validate with that config
    Then the execution should succeed
    And the output should show validation warning about missing output type
