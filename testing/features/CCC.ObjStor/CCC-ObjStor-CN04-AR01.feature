@PerService @CCC.ObjStor @tlp-clear @tlp-green @tlp-amber @tlp-red @CCC.ObjStor.CN04.AR01
Feature: CCC.ObjStor.CN04.AR01
  As a security administrator
  I want objects to automatically receive a default retention policy upon upload
  So that critical data is protected from premature deletion or modification

  Background:
    Given a cloud api for "{Provider}" in "api"
    And I call "{api}" with "GetServiceAPI" with parameter "object-storage"
    And I refer to "{result}" as "storage"
    And I call "{api}" with "GetServiceAPI" with parameter "iam"
    And I refer to "{result}" as "iamService"

  Scenario: Service applies default retention policy to newly uploaded object
    Given I call "{iamService}" with "ProvisionUserWithAccess" with parameters "test-user-write", "{UID}" and "write"
    And I refer to "{result}" as "testUserWrite"
    And I attach "{result}" to the test output as "write-user-identity.json"
    And I call "{api}" with "GetServiceAPIWithIdentity" with parameters "object-storage", "{testUserWrite}" and "{true}"
    And "{result}" is not an error
    And I refer to "{result}" as "userStorage"
    When I call "{userStorage}" with "CreateObject" with parameters "{ResourceName}", "test-retention-object.txt" and "protected data"
    And I attach "{result}" to the test output as "uploaded-object.json"
    And I call "{userStorage}" with "GetObjectRetentionDurationDays" with parameters "{ResourceName}" and "test-retention-object.txt"
    Then "{result}" should be greater than "1"
    And I call "{storage}" with "DeleteObject" with parameters "{ResourceName}" and "test-retention-object.txt"

  Scenario: Service enforces retention policy on newly created objects
    When I call "{storage}" with "CreateObject" with parameters "{ResourceName}", "immediate-delete-test.txt" and "test content"
    Then "{result}" is not an error
    When I call "{storage}" with "DeleteObject" with parameters "{ResourceName}" and "immediate-delete-test.txt"
    Then "{result}" is an error
    And I attach "{result}" to the test output as "immediate-delete-error.txt"
    And "{result}" should contain "retention"

  Scenario: Service validates retention period meets minimum requirements
    When I call "{storage}" with "CreateObject" with parameters "{ResourceName}", "retention-period-test.txt" and "compliance data"
    And I call "{storage}" with "GetObjectRetentionDurationDays" with parameters "{ResourceName}" and "retention-period-test.txt"
    Then "{result}" should be greater than "1"
    And I attach "{result}" to the test output as "retention-period-days.json"
    And I call "{storage}" with "DeleteObject" with parameters "{ResourceName}" and "retention-period-test.txt"
