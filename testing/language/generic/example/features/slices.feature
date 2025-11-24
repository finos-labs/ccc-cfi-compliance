Feature: Slice validation and assertions
  As a developer
  I want to test slice operations and validations
  So that I can verify array/slice handling works correctly

  Background:
    Given I have test slices configured in "testSlices"

  @slices
  Scenario: Validate slice of objects with exact contents
    When I refer to "{testSlices.users}" as "usersList"
    Then "{usersList}" is a slice of objects with the following contents
      | id | name      | email              | active |
      | 1  | Alice     | alice@example.com  | true   |
      | 2  | Bob       | bob@example.com    | false  |
      | 3  | Charlie   | charlie@example.com| true   |

  @slices
  Scenario: Validate slice of objects with length
    When I refer to "{testSlices.users}" as "usersList"
    Then "{usersList}" is a slice of objects with length "3"

  @slices
  Scenario: Validate slice of strings with values
    When I refer to "{testSlices.colors}" as "colorsList"
    Then "{colorsList}" is a slice of strings with the following values
      | value   |
      | red     |
      | green   |
      | blue    |

  @slices
  Scenario: Validate empty slice has length zero
    When I refer to "{testSlices.empty}" as "emptyList"
    Then "{emptyList}" is a slice of objects with length "0"

  @slices
  Scenario: Validate single item slice
    When I refer to "{testSlices.single}" as "singleItem"
    Then "{singleItem}" is a slice of objects with length "1"
    And "{singleItem}" is a slice of objects with the following contents
      | id | value      |
      | 1  | only-item  |

  @slices
  Scenario: Validate slice of numbers as strings
    When I refer to "{testSlices.numbers}" as "numbersList"
    Then "{numbersList}" is a slice of strings with the following values
      | value |
      | 1     |
      | 2     |
      | 3     |
      | 4     |
      | 5     |

  @slices
  Scenario: Validate slice length with variable
    Given "expectedLength" is a function which returns a value of "3"
    When I call "{expectedLength}"
    And I refer to "{result}" as "len"
    And I refer to "{testSlices.users}" as "usersList"
    Then "{usersList}" is a slice of objects with length "{len}"

  @slices
  Scenario: Validate slice of objects with nested properties
    When I refer to "{testSlices.products}" as "productsList"
    Then "{productsList}" is a slice of objects with the following contents
      | id | name      | price  |
      | 1  | Widget    | 9.99   |
      | 2  | Gadget    | 19.99  |
      | 3  | Doohickey | 29.99  |

  @slices
  Scenario: Validate boolean values in slice
    When I refer to "{testSlices.flags}" as "flagsList"
    Then "{flagsList}" is a slice of objects with the following contents
      | enabled | visible |
      | true    | false   |
      | false   | true    |
      | true    | true    |

  @slices
  Scenario: Validate slice with special characters in strings
    When I refer to "{testSlices.special}" as "specialList"
    Then "{specialList}" is a slice of strings with the following values
      | value        |
      | hello-world  |
      | test_value   |
      | some.thing   |

  @slices
  Scenario: Validate large slice length
    When I refer to "{testSlices.large}" as "largeList"
    Then "{largeList}" is a slice of objects with length "10"

  @slices
  Scenario: Chain operations with slices
    When I refer to "{testSlices.colors}" as "colorsList"
    Then "{colorsList}" is a slice of objects with length "3"
    And "{colorsList}" is a slice of strings with the following values
      | value   |
      | red     |
      | green   |
      | blue    |

  @slices
  Scenario: Validate API response as slice
    Given I have an API client configured in "apiClient"
    When I call "{apiClient}" with "Get" with parameter "/users"
    And I refer to "{result.data}" as "responseData"
    Then "{responseData}" is a slice of objects with length "2"
    And "{responseData}" is a slice of objects with the following contents
      | id | name     | active |
      | 1  | John Doe | true   |
      | 2  | Jane Doe | false  |

  @slices
  Scenario: Validate slice of countries with codes
    When I refer to "{testSlices.countries}" as "countriesList"
    Then "{countriesList}" is a slice of objects with the following contents
      | code | name           |
      | US   | United States  |
      | UK   | United Kingdom |
      | CA   | Canada         |

  @slices
  Scenario: Validate days of week slice
    When I refer to "{testSlices.daysOfWeek}" as "daysList"
    Then "{daysList}" is a slice of strings with the following values
      | value     |
      | Monday    |
      | Tuesday   |
      | Wednesday |
      | Thursday  |
      | Friday    |
      | Saturday  |
      | Sunday    |
    And "{daysList}" is a slice of objects with length "7"

  @slices
  Scenario: Validate slice after filtering operation
    When I refer to "{testSlices.users}" as "usersList"
    And I refer to "{usersList}" as "filteredUsers"
    Then "{filteredUsers}" is a slice of objects with length "3"

  @slices
  Scenario: Validate mixed data types in slice of objects
    When I refer to "{testSlices.mixed}" as "mixedList"
    Then "{mixedList}" is a slice of objects with the following contents
      | id | name   | count | active |
      | 1  | First  | 100   | true   |
      | 2  | Second | 200   | false  |

  @slices
  Scenario: Validate status codes slice
    When I refer to "{testSlices.statusCodes}" as "codesList"
    Then "{codesList}" is a slice of strings with the following values
      | value |
      | 200   |
      | 201   |
      | 400   |
      | 404   |
      | 500   |
    And "{codesList}" is a slice of objects with length "5"

