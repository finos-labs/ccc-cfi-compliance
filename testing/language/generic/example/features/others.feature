Feature: Other test patterns
  As a developer
  I want to test miscellaneous patterns
  So that I can verify error handling and timing work correctly

  @others
  Scenario: Error handling with message
    Given "errorFunction" is a function which throws an error
    When I call "{errorFunction}"
    Then "{result}" is an error with message "something went wrong"

  @others
  Scenario: Error type check
    Given "genericError" is a function which throws an error
    When I call "{genericError}"
    Then "{result}" is an error

  @others
  Scenario: Timing operations
    Given we wait for a period of "100" ms
    Then "{result}" is nil
