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

  @others @attach
  Scenario: Attach string to test output
    Given "message" is a function which returns a value of "Hello, World!"
    When I call "{message}"
    And I attach "{result}" to the test output
    Then "{result}" is "Hello, World!"

  @others @attach
  Scenario: Attach JSON data to test output
    Given I have test data in "users"
    When I attach "{users}" to the test output
    Then "{users}" is not nil

  @others @attach
  Scenario: Attach API response to test output
    Given I have an API client configured in "apiClient"
    When I call "{apiClient}" with "Get" with parameter "/users"
    And I attach "{result}" to the test output
    Then "{result}" is not nil
    And "{result.status}" is "200"

  @others @attach
  Scenario: Attach multiple items to test output
    Given "text1" is a function which returns a value of "First attachment"
    And "text2" is a function which returns a value of "Second attachment"
    When I call "{text1}"
    And I attach "{result}" to the test output
    And I call "{text2}"
    And I attach "{result}" to the test output
    Then "{result}" is "Second attachment"
