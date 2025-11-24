Feature: Error handling in method and function calls
  As a developer
  I want to test error handling when calling methods that return (value, error)
  So that I can verify errors are properly caught and reported

  Background:
    Given I have an API client with error handling in "errorClient"
    And I have a service with error methods in "errorService"

  @errors
  Scenario: Object method with no parameters returns error
    When I call "{errorService}" with "FailWithNoParams"
    Then "{result}" is an error
    And "{result}" has error message "method failed: no params"

  @errors
  Scenario: Object method with one parameter returns error
    When I call "{errorService}" with "FailWithOneParam" with parameter "test"
    Then "{result}" is an error
    And "{result}" has error message "method failed with param: test"

  @errors
  Scenario: Object method with two parameters returns error
    When I call "{errorService}" with "FailWithTwoParams" with parameters "foo" and "bar"
    Then "{result}" is an error
    And "{result}" has error message "method failed with params: foo, bar"

  @errors
  Scenario: Object method with three parameters returns error
    When I call "{errorService}" with "FailWithThreeParams" with parameters "a", "b" and "c"
    Then "{result}" is an error
    And "{result}" has error message "method failed with params: a, b, c"

  @errors
  Scenario: Object method with no parameters returns success
    When I call "{errorService}" with "SucceedWithNoParams"
    Then "{result}" is not nil
    And "{result}" is "success: no params"

  @errors
  Scenario: Object method with one parameter returns success
    When I call "{errorService}" with "SucceedWithOneParam" with parameter "test"
    Then "{result}" is "success: test"

  @errors
  Scenario: Object method with two parameters returns success
    When I call "{errorService}" with "SucceedWithTwoParams" with parameters "x" and "y"
    Then "{result}" is "success: x, y"

  @errors
  Scenario: Object method with three parameters returns success
    When I call "{errorService}" with "SucceedWithThreeParams" with parameters "1", "2" and "3"
    Then "{result}" is "success: 1, 2, 3"

  @errors @async
  Scenario: Async task with method returning error (one param)
    When I start task "failTask" by calling "{errorService}" with "FailWithOneParam" with parameter "async-fail"
    And I wait for task "failTask" to complete
    Then "{failTask}" is an error
    And "{failTask}" has error message "method failed with param: async-fail"

  @errors @async
  Scenario: Async task with method returning success (one param)
    When I start task "successTask" by calling "{errorService}" with "SucceedWithOneParam" with parameter "async-ok"
    And I wait for task "successTask" to complete
    Then "{successTask}" is "success: async-ok"

  @errors @async
  Scenario: Async task with method returning error (two params)
    When I start task "failTask2" by calling "{errorService}" with "FailWithTwoParams" with parameters "p1" and "p2"
    And I wait for task "failTask2" to complete
    Then "{failTask2}" is an error
    And "{failTask2}" has error message "method failed with params: p1, p2"

  @errors @async
  Scenario: Async task with method returning success (two params)
    When I start task "successTask2" by calling "{errorService}" with "SucceedWithTwoParams" with parameters "alpha" and "beta"
    And I wait for task "successTask2" to complete
    Then "{successTask2}" is "success: alpha, beta"

  @errors @async
  Scenario: Async task with method returning error (three params)
    When I start task "failTask3" by calling "{errorService}" with "FailWithThreeParams" with parameters "x", "y" and "z"
    And I wait for task "failTask3" to complete
    Then "{failTask3}" is an error
    And "{failTask3}" has error message "method failed with params: x, y, z"

  @errors @async
  Scenario: Async task with method returning success (three params)
    When I start task "successTask3" by calling "{errorService}" with "SucceedWithThreeParams" with parameters "red", "green" and "blue"
    And I wait for task "successTask3" to complete
    Then "{successTask3}" is "success: red, green, blue"

  @errors @async
  Scenario: All-in-one wait for method with error (one param)
    When I wait for "{errorService}" with "FailWithOneParam" with parameter "direct-fail"
    Then "{result}" is an error
    And "{result}" has error message "method failed with param: direct-fail"

  @errors @async
  Scenario: All-in-one wait for method with success (one param)
    When I wait for "{errorService}" with "SucceedWithOneParam" with parameter "direct-ok"
    Then "{result}" is "success: direct-ok"

  @errors
  Scenario: API client GetWithError returns error on invalid endpoint
    When I call "{errorClient}" with "GetWithError" with parameter "/invalid"
    Then "{result}" is an error
    And "{result}" has error message "endpoint not found: /invalid"

  @errors
  Scenario: API client GetWithError returns success on valid endpoint
    When I call "{errorClient}" with "GetWithError" with parameter "/users"
    Then "{result}" is not nil
    And "{result}" is an object with the following contents
      | status | message |
      |    200 | success |

  @errors
  Scenario: Mixed error and success in sequence
    # First call fails
    When I call "{errorService}" with "FailWithOneParam" with parameter "fail-first"
    Then "{result}" is an error
    And "{result}" has error message "method failed with param: fail-first"
    # Second call succeeds
    When I call "{errorService}" with "SucceedWithOneParam" with parameter "then-succeed"
    Then "{result}" is "success: then-succeed"
    And "{result}" is not an error

  @errors @panic
  Scenario: Panic recovery with no parameters
    When I call "{errorService}" with "PanicWithNoParams"
    Then "{result}" is an error
    And "{result}" contains "panic: this method panics"

  @errors @panic
  Scenario: Panic recovery with one parameter
    When I call "{errorService}" with "PanicWithOneParam" with parameter "test-panic"
    Then "{result}" is an error
    And "{result}" contains "panic: test-panic"

  @errors @panic
  Scenario: Panic recovery with two parameters
    When I call "{errorService}" with "PanicWithTwoParams" with parameters "foo" and "bar"
    Then "{result}" is an error
    And "{result}" contains "panic: foo, bar"

  @errors @panic
  Scenario: Panic recovery with three parameters
    When I call "{errorService}" with "PanicWithThreeParams" with parameters "a", "b" and "c"
    Then "{result}" is an error
    And "{result}" contains "panic: a, b, c"

  @errors @panic
  Scenario: Panic recovery preserves error in result
    When I call "{errorService}" with "PanicWithOneParam" with parameter "preserved"
    Then "{result}" is an error
    And "{result}" is not nil
    # Can still make further calls after panic
    When I call "{errorService}" with "SucceedWithOneParam" with parameter "after-panic"
    Then "{result}" is "success: after-panic"
