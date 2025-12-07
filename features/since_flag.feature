Feature: Since Flag
  As a devpipe user
  I want to filter tasks based on git changes since a specific ref
  So that I can run only tasks affected by my changes

  Scenario: Since flag overrides config git.ref
    Given a config with git.ref = "main"
    When I run devpipe with --since "develop"
    Then tasks should filter based on develop
    And the config git.ref should be overridden

  Scenario: Since with valid git ref filters tasks
    Given a config with multiple tasks and git history
    And git changes since "HEAD~3"
    When I run devpipe with --since "HEAD~3"
    Then only tasks affected by those changes should run
    And unchanged tasks should be skipped

  Scenario: Since with invalid ref handles gracefully
    Given a config with multiple tasks and git history
    When I run devpipe with --since "nonexistent-ref"
    Then the execution should succeed
    And the output should show no changed files

  Scenario: Since flag with watchPaths filters correctly
    Given a config with tasks that have watchPaths
    And git changes to "frontend/app.ts" since "HEAD~1"
    When I run devpipe with --since "HEAD~1"
    Then only tasks with matching watchPaths should run
    And tasks with non-matching watchPaths should be skipped

  Scenario: Since flag overrides git mode
    Given a config with git mode "staged"
    And git changes since "HEAD~1"
    When I run devpipe with --since "HEAD~1"
    Then the git mode should be "ref"
    And tasks should filter based on the ref

  Scenario: Since with HEAD shows no changes
    Given a config with multiple tasks and git history
    When I run devpipe with --since "HEAD"
    Then the execution should succeed
    And the output should show no changed files

  Scenario: Since with commit SHA
    Given a config with multiple tasks and git history
    And git changes since a specific commit SHA
    When I run devpipe with --since that commit SHA
    Then only tasks affected by those changes should run

  Scenario: Since with tag reference
    Given a config with multiple tasks and git history
    And a git tag "v1.0.0" exists
    And git changes since "v1.0.0"
    When I run devpipe with --since "v1.0.0"
    Then only tasks affected by those changes should run

  Scenario: Since in brand new repo
    Given a brand new git repo with no commits
    And a config with multiple tasks
    When I run devpipe with --since "HEAD"
    Then the execution should succeed
    And the output should show no changed files
