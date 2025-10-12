Feature: Asynchronous operation patterns
  As a developer
  I want to test async operations
  So that I can verify goroutine-based task execution works correctly

  Background:
    Given I have an API client configured in "apiClient"

  @async
  Scenario: Start task by calling function
    Given "slowFunction" is a function which returns a value of "completed"
    When I start task "task1" by calling "{slowFunction}"
    Then I wait for task "task1" to complete
    And "{result}" is "completed"

  @async
  Scenario: Start task by calling function with parameter
    Given "fn1" is a test function with one parameter
    When I start task "task2" by calling "{fn1}" with parameter "async"
    Then I wait for task "task2" to complete
    And "{result}" is "one-param:async"

  @async
  Scenario: Start task by calling function with two parameters
    Given "fn2" is a test function with two parameters
    When I start task "task3" by calling "{fn2}" with parameters "p1" and "p2"
    Then I wait for task "task3" to complete
    And "{result}" is "two-params:p1,p2"

  @async
  Scenario: Start task by calling function with three parameters
    Given "fn3" is a test function with three parameters
    When I start task "task4" by calling "{fn3}" with parameters "a", "b" and "c"
    Then I wait for task "task4" to complete
    And "{result}" is "three-params:a,b,c"

  @async
  Scenario: Start task by calling object method
    When I start task "task5" by calling "{apiClient}" with "Get" with parameter "/data"
    Then I wait for task "task5" to complete
    And "{result}" is not nil

  @async
  Scenario: Start task by calling object method with parameter
    When I start task "task6" by calling "{apiClient}" with "Get" with parameter "/users"
    Then I wait for task "task6" to complete
    And "{result}" is an object with the following contents
      | status | message |
      |    200 | success |

  @async
  Scenario: Wait for task with timeout
    Given "slowFunction" is a function which returns a value of "done"
    When I start task "task7" by calling "{slowFunction}"
    Then I wait for task "task7" to complete within "5000" ms
    And "{result}" is "done"

  @async
  Scenario: All-in-one wait for function
    Given "quickFunction" is a function which returns a value of "quick result"
    When I wait for "{quickFunction}"
    Then "{result}" is "quick result"

  @async
  Scenario: All-in-one wait for function with parameter
    Given "fn1" is a test function with one parameter
    When I wait for "{fn1}" with parameter "wait"
    Then "{result}" is "one-param:wait"

  @async
  Scenario: All-in-one wait for function with two parameters
    Given "fn2" is a test function with two parameters
    When I wait for "{fn2}" with parameters "x" and "y"
    Then "{result}" is "two-params:x,y"

  @async
  Scenario: All-in-one wait for function with three parameters
    Given "fn3" is a test function with three parameters
    When I wait for "{fn3}" with parameters "1", "2" and "3"
    Then "{result}" is "three-params:1,2,3"

  @async
  Scenario: All-in-one wait for object method
    When I wait for "{apiClient}" with "Get" with parameter "/data"
    Then "{result}" is not nil

  @async
  Scenario: All-in-one wait for object method with parameter
    When I wait for "{apiClient}" with "Get" with parameter "/status"
    Then "{result}" is an object with the following contents
      | status | message |
      |    200 | success |

  @async
  Scenario: Start task with three parameter function
    Given "fn3" is a test function with three parameters
    When I start task "sumTask" by calling "{fn3}" with parameters "10", "20" and "30"
    Then I wait for task "sumTask" to complete
    And "{result}" is "three-params:10,20,30"

  @async
  Scenario: Start task with object method (no params)
    Given I have a test object in "testObj"
    When I start task "getValueTask" by calling "{testObj}" with "GetValue"
    Then I wait for task "getValueTask" to complete
    And "{result}" is "test-value"

  @async
  Scenario: Start task with object method (two params)
    Given I have a test object in "testObj"
    When I start task "combineTask" by calling "{testObj}" with "CombineStrings" with parameters "foo" and "bar"
    Then I wait for task "combineTask" to complete
    And "{result}" is "foo-bar"

  @async
  Scenario: Start task with object method (three params)
    Given I have a test object in "testObj"
    When I start task "joinTask" by calling "{testObj}" with "JoinThree" with parameters "x", "y" and "z"
    Then I wait for task "joinTask" to complete
    And "{result}" is "x-y-z"

  @async
  Scenario: All-in-one wait for function with three params
    Given "fn3" is a test function with three parameters
    When I wait for "{fn3}" with parameters "5", "15" and "25"
    Then "{result}" is "three-params:5,15,25"

  @async
  Scenario: All-in-one wait for object method (no params)
    Given I have a test object in "testObj"
    When I wait for "{testObj}" with "GetValue"
    Then "{result}" is "test-value"

  @async
  Scenario: All-in-one wait for object method (two params)
    Given I have a test object in "testObj"
    When I wait for "{testObj}" with "CombineStrings" with parameters "async" and "test"
    Then "{result}" is "async-test"

  @async
  Scenario: All-in-one wait for object method (three params)
    Given I have a test object in "testObj"
    When I wait for "{testObj}" with "JoinThree" with parameters "1", "2" and "3"
    Then "{result}" is "1-2-3"

  @async
  Scenario: Multiple concurrent tasks
    Given "task1Fn" is a function which returns a value of "result1"
    And "task2Fn" is a function which returns a value of "result2"
    And "task3Fn" is a function which returns a value of "result3"
    When I start task "concurrent1" by calling "{task1Fn}"
    And I start task "concurrent2" by calling "{task2Fn}"
    And I start task "concurrent3" by calling "{task3Fn}"
    When I wait for task "concurrent1" to complete
    Then "{result}" is "result1"
    When I wait for task "concurrent2" to complete
    Then "{result}" is "result2"
    When I wait for task "concurrent3" to complete
    Then "{result}" is "result3"
