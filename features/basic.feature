Feature: Basic DevPipe Functionality
  As a developer
  I want to verify basic devpipe operations
  So that I can ensure the tool works correctly

  Scenario: Simple calculation
    Given I have the number 5
    When I add 3 to it
    Then the result should be 8

  Scenario: String concatenation
    Given I have the string "Hello"
    When I append " World"
    Then the result should be "Hello World"
