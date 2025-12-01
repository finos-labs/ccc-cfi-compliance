@CCC.ObjStor @tlp-amber @tlp-red @CCC.ObjStor.CN01.AR02
Feature: CCC.ObjStor.CN01.AR02
  As a security administrator
  I want to prevent any requests to read protected objects using untrusted KMS keys
  So that data encryption integrity and availability are protected against unauthorized encryption

  Background:
    Given a cloud api for "{Provider}" in "api"
    And I call "{api}" with "GetServiceAPI" with parameter "object-storage"
    And I refer to "{result}" as "storage"
    And I call "{api}" with "GetServiceAPI" with parameter "iam"
    And I refer to "{result}" as "iamService"
    And I call "{storage}" with "CreateObject" with parameters "{ResourceName}", "test-object.txt" and "test content"
    And "{result}" is not an error

  Scenario: Service prevents reading object with no access
    Given I call "{iamService}" with "ProvisionUser" with parameter "test-user-no-access"
    And I refer to "{result}" as "testUserNoAccess"
    And I call "{iamService}" with "SetAccess" with parameters "{testUserNoAccess}", "{UID}" and "none"
    And I attach "{result}" to the test output as "no-access-user-policy.json"
    Given I attach "{testUserNoAccess}" to the test output as "no-access-user-identity.json"
    And I call "{api}" with "GetServiceAPIWithIdentity" with parameters "object-storage" and "{testUserNoAccess}"
    And "{result}" is not an error
    And I refer to "{result}" as "userStorage"
    When I call "{userStorage}" with "ReadObject" with parameters "{ResourceName}" and "test-object.txt"
    Then "{result}" is an error
    And I attach "{result}" to the test output as "no-access-read-object-error.txt"

  Scenario: Service allows reading object with read access
    Given I call "{iamService}" with "ProvisionUser" with parameter "test-user-read"
    And I refer to "{result}" as "testUserRead"
    And I call "{iamService}" with "SetAccess" with parameters "{testUserRead}", "{UID}" and "read"
    And I attach "{result}" to the test output as "read-user-policy.json"
    Given I attach "{testUserRead}" to the test output as "read-user-identity.json"
    And I call "{api}" with "GetServiceAPIWithIdentity" with parameters "object-storage" and "{testUserRead}"
    And "{result}" is not an error
    And I attach "{result}" to the test output as "read-storage-service.json"
    And I refer to "{result}" as "userStorage"
    When I call "{userStorage}" with "ReadObject" with parameters "{ResourceName}" and "test-object.txt"
    Then "{result}" is not an error
    And I attach "{result}" to the test output as "read-read-object-result.json"
    And I call "{storage}" with "DeleteObject" with parameters "{ResourceName}" and "test-object.txt"
