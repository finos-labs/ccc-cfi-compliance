@PerService @CCC.ObjStor @tlp-clear @tlp-green @tlp-amber @tlp-red @CCC.ObjStor.CN04.AR02
Feature: CCC.ObjStor.CN04.AR02
  As a security administrator
  I want to prevent deletion or modification of objects under active retention
  So that data integrity and compliance requirements are maintained

  Background:
    Given a cloud api for "{Provider}" in "api"
    And I call "{api}" with "GetServiceAPI" using argument "object-storage"
    And I refer to "{result}" as "storage"
    And I call "{api}" with "GetServiceAPI" using argument "iam"
    And I refer to "{result}" as "iamService"

  Scenario: Service prevents object deletion by write user during retention period
    Given I call "{iamService}" with "ProvisionUserWithAccess" using arguments "test-user-write", "{UID}", and "write"
    And I refer to "{result}" as "testUserWrite"
    And I attach "{result}" to the test output as "write-user-identity.json"
    And I call "{api}" with "GetServiceAPIWithIdentity" using arguments "object-storage", "{testUserWrite}", and "{true}"
    And "{result}" is not an error
    And I refer to "{result}" as "userStorage"
    When I call "{userStorage}" with "CreateObject" using arguments "{ResourceName}", "protected-object.txt", and "immutable data"
    Then "{result}" is not an error
    And I attach "{result}" to the test output as "protected-object.json"
    When I call "{userStorage}" with "DeleteObject" using arguments "{ResourceName}" and "protected-object.txt"
    Then "{result}" is an error
    And I attach "{result}" to the test output as "delete-protected-error.txt"
    And "{result}" should contain one of "retention, locked, immutable, protected"

  Scenario: Service prevents object deletion by admin user during retention period
    When I call "{storage}" with "CreateObject" using arguments "{ResourceName}", "admin-protected-object.txt", and "compliance data"
    Then "{result}" is not an error
    When I call "{storage}" with "DeleteObject" using arguments "{ResourceName}" and "admin-protected-object.txt"
    Then "{result}" is an error
    And I attach "{result}" to the test output as "admin-delete-protected-error.txt"
    And "{result}" should contain "retention"

  Scenario: Service prevents object modification during retention period
    Given I call "{iamService}" with "ProvisionUserWithAccess" using arguments "test-user-write", "{UID}", and "write"
    And I refer to "{result}" as "testUserWrite"
    And I call "{api}" with "GetServiceAPIWithIdentity" using arguments "object-storage", "{testUserWrite}", and "{true}"
    And "{result}" is not an error
    And I refer to "{result}" as "userStorage"
    When I call "{userStorage}" with "CreateObject" using arguments "{ResourceName}", "modify-test-object.txt", and "original content"
    Then "{result}" is not an error
    And I attach "{result}" to the test output as "original-object.json"
    When I call "{userStorage}" with "CreateObject" using arguments "{ResourceName}", "modify-test-object.txt", and "modified content"
    Then "{result}" is an error
    And I attach "{result}" to the test output as "modify-protected-error.txt"
    And "{result}" should contain one of "retention, locked, immutable, protected, exists"

  Scenario: Service allows object read access during retention period
    When I call "{storage}" with "CreateObject" using arguments "{ResourceName}", "readable-protected-object.txt", and "readable data"
    Then "{result}" is not an error
    Given I call "{iamService}" with "ProvisionUserWithAccess" using arguments "test-user-read", "{UID}", and "read"
    And I refer to "{result}" as "testUserRead"
    And I attach "{result}" to the test output as "read-user-identity.json"
    And I call "{api}" with "GetServiceAPIWithIdentity" using arguments "object-storage", "{testUserRead}", and "{true}"
    And "{result}" is not an error
    And I refer to "{result}" as "userStorage"
    When I call "{userStorage}" with "ReadObject" using arguments "{ResourceName}" and "readable-protected-object.txt"
    Then "{result}" is not an error
    And I refer to "{result}" as "readResult"
    And I attach "{result}" to the test output as "read-protected-object.json"
    And "{readResult.Name}" is "readable-protected-object.txt"
    And I call "{storage}" with "DeleteObject" using arguments "{ResourceName}" and "readable-protected-object.txt"
