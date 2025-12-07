Feature: WatchPaths Task Filtering
  As a devpipe user
  I want tasks to run only when their watched files change
  So that I can optimize my pipeline execution

  Scenario: Task with watchPaths skipped when no files changed
    Given a git repo with no changes
    And a task with watchPaths for "src/**/*.ts"
    When I run devpipe
    Then the task should be skipped
    And the skip reason should mention "no changed files"

  Scenario: Task with watchPaths runs when matching file changes
    Given a git repo with changes to "src/app.ts"
    And a task with watchPaths for "src/**/*.ts"
    When I run devpipe
    Then the task should run

  Scenario: Task with watchPaths skipped when non-matching file changes
    Given a git repo with changes to "README.md"
    And a task with watchPaths for "src/**/*.ts"
    When I run devpipe
    Then the task should be skipped
    And the skip reason should mention "no matching changes"

  Scenario: Task without watchPaths always runs
    Given a git repo with no changes
    And a task without watchPaths
    When I run devpipe
    Then the task should run

  Scenario: Multiple watchPaths patterns
    Given a git repo with changes to "package.json"
    And a task with watchPaths for "src/**/*.ts" and "package.json"
    When I run devpipe
    Then the task should run

  Scenario: Glob pattern with double star
    Given a git repo with changes to "src/components/Button.tsx"
    And a task with watchPaths for "src/**/*.tsx"
    When I run devpipe
    Then the task should run

  Scenario: Ignore watch paths flag
    Given a git repo with changes to "README.md"
    And a task with watchPaths for "src/**/*.ts"
    When I run devpipe with --ignore-watch-paths
    Then the task should run

  Scenario: WatchPaths relative to workdir
    Given a git repo with changes to "frontend/src/app.ts"
    And a task with workdir "frontend" and watchPaths for "src/**/*.ts"
    When I run devpipe
    Then the task should run

  Scenario: Multiple tasks with different watchPaths
    Given a git repo with changes to "frontend/app.ts"
    And a task "frontend-test" with watchPaths for "frontend/**"
    And a task "backend-test" with watchPaths for "backend/**"
    When I run devpipe
    Then task "frontend-test" should run
    And task "backend-test" should be skipped
