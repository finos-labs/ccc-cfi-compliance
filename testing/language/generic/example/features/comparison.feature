Feature: Data comparison and validation patterns
  As a developer
  I want to test data comparison operations
  So that I can verify validation patterns work correctly

  Background:
    Given I have test data in "users"

  @comparison
  Scenario: Array validation with data tables
    Then "{users}" is an slice of objects with the following contents
      | name     | active | profile.email    |
      | John Doe | true   | john@example.com |
      | Jane Doe | false  | jane@example.com |

  @comparison
  Scenario: Boolean and nil checks
    Given "trueValue" is a function which returns a value of "{true}"
    And "falseValue" is a function which returns a value of "{false}"
    And "nilValue" is a function which returns a value of "{nil}"
    When I call "{trueValue}"
    And I refer to "{result}" as "boolResult"
    Then "{boolResult}" is true
    When I call "{falseValue}"
    And I refer to "{result}" as "boolResult2"
    Then "{boolResult2}" is false
    When I call "{nilValue}"
    And I refer to "{result}" as "nilResult"
    Then "{nilResult}" is nil

  @comparison
  Scenario: Array length validation
    Then "{users}" is an slice of objects with length "2"

  @comparison
  Scenario: String array validation
    Given "colorArray" is a string array with colors
    Then "{colorArray}" is an slice of strings with the following values
      | value |
      | red   |
      | blue  |
      | green |

  @comparison
  Scenario: Empty checks
    Given "emptyArray" is an empty array
    And "emptyString" is an empty string
    Then "{emptyArray}" is empty
    And "{emptyString}" is empty

  @comparison
  Scenario: Struct field navigation
    Given "person" is a struct with Name "Alice" and Age "30"
    Then "{person.Name}" is "Alice"
    And "{person.Age}" is "30"
    And "{person}" is not nil

  @comparison
  Scenario: Array contains at least some elements
    Then "{users}" is a slice of objects with at least the following contents
      | name     | active |
      | John Doe | true   |

  @comparison
  Scenario: Array doesn't contain unwanted elements
    Then "{users}" is a slice of objects which doesn't contain any of
      | name       | active |
      | Intruder 1 | true   |
      | Dodgy Guy  | false  |
