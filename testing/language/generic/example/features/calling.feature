Feature: Function and method calling patterns
  As a developer
  I want to test function and method calls
  So that I can verify the calling patterns work correctly

  Background:
    Given I have an API client configured in "apiClient"
    And I have test data in "users"

  @calling
  Scenario: Basic API testing with generic steps
    When I call "{apiClient}" with "Get" with parameter "/users"
    Then "{result}" is an object with the following contents
      | status | message |
      |    200 | success |

  @calling
  Scenario: Variable assignment and function calls
    Given "testValue" is a function which returns a value of "hello world"
    When I refer to "{testValue}" as "myFunction"
    And I call "{myFunction}"
    Then "{result}" is "hello world"

  @calling
  Scenario: Counter functionality
    Given "clickHandler" is a invocation counter into "clickCount"
    When I call "{clickHandler}"
    And I call "{clickHandler}"
    And I call "{clickHandler}"
    Then "{clickCount}" is "3"

  @calling
  Scenario: Object method calls
    When I call "{apiClient}" with "Get" with parameter "/status"
    Then "{result}" is not nil
    And "{result}" is an object with the following contents
      | status | message |
      |    200 | success |

  @calling
  Scenario: Function calls with one parameter
    Given "fn1" is a test function with one parameter
    When I call "{fn1}" with parameter "test"
    Then "{result}" is "one-param:test"

  @calling
  Scenario: Function calls with two parameters
    Given "fn2" is a test function with two parameters
    When I call "{fn2}" with parameters "a" and "b"
    Then "{result}" is "two-params:a,b"

  @calling
  Scenario: Function calls with three parameters
    Given "fn3" is a test function with three parameters
    When I call "{fn3}" with parameters "x", "y" and "z"
    Then "{result}" is "three-params:x,y,z"

  @calling
  Scenario: Object method calls with no parameters
    Given I have a test object in "testObj"
    When I call "{testObj}" with "GetValue"
    Then "{result}" is "test-value"

  @calling
  Scenario: Object method calls with two parameters
    Given I have a test object in "testObj"
    When I call "{testObj}" with "CombineStrings" with parameters "hello" and "world"
    Then "{result}" is "hello-world"

  @calling
  Scenario: Object method calls with three parameters
    Given I have a test object in "testObj"
    When I call "{testObj}" with "JoinThree" with parameters "a", "b" and "c"
    Then "{result}" is "a-b-c"
