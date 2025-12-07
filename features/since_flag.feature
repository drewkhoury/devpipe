Feature: Since Flag
  As a devpipe user
  I want to filter tasks based on git changes since a specific ref
  So that I can run only tasks affected by my changes

  @wip
  Scenario: Since flag overrides config git.ref
    Given a config with git.ref = "main"
    When I run devpipe with --since "develop"
    Then tasks should filter based on develop
    And the config git.ref should be overridden

  @wip
  Scenario: Since with valid git ref filters tasks
    Given a config with multiple tasks
    And git changes since "HEAD~3"
    When I run devpipe with --since "HEAD~3"
    Then only tasks affected by those changes should run
    And unchanged tasks should be skipped

  @wip
  Scenario: Since with invalid ref shows error
    Given a config with multiple tasks
    When I run devpipe with --since "nonexistent-ref"
    Then the execution should fail
    And the output should indicate invalid git ref
