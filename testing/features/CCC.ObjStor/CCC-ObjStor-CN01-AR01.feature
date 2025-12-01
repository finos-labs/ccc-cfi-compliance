@CCC.ObjStor @tlp-amber @tlp-red
Feature: CCC.ObjStor.CN01.AR01
  As a security administrator
  I want to prevent any requests to read protected buckets using untrusted KMS keys
  So that data encryption integrity and availability are protected against unauthorized encryption

  Background:
    Given a cloud api for "{Provider}" in "api"
    And I call "{api}" with "GetServiceAPI" with parameter "object-storage"
    And I refer to "{result}" as "storage"
    And I call "{api}" with "GetServiceAPI" with parameter "iam"
    And I refer to "{result}" as "iamService"

  Scenario: Service prevents reading bucket with no access
    Given I call "{iamService}" with "ProvisionUser" with parameter "test-user-no-access"
    And I refer to "{result}" as "testUserNoAccess"
    And I call "{iamService}" with "SetAccess" with parameters "{testUserNoAccess}", "{UID}" and "none"
    And I attach "{result}" to the test output as "no-access-user-policy.json"
    Given I attach "{testUserNoAccess}" to the test output as "no-access-user-identity.json"
    And I call "{api}" with "GetServiceAPIWithIdentity" with parameters "object-storage" and "{testUserNoAccess}"
    And "{result}" is not an error
    And I refer to "{result}" as "userStorage"
    When I call "{userStorage}" with "ListObjects" with parameter "{ResourceName}"
    Then "{result}" is an error
    And I attach "{result}" to the test output as "no-access-list-error.txt"

  Scenario: Service allows reading bucket with read access
    Given I call "{iamService}" with "ProvisionUser" with parameter "test-user-read"
    And I refer to "{result}" as "testUserRead"
    And I call "{iamService}" with "SetAccess" with parameters "{testUserRead}", "{UID}" and "read"
    And I attach "{result}" to the test output as "read-user-policy.json"
    Given I attach "{testUserRead}" to the test output as "read-user-identity.json"
    And I call "{api}" with "GetServiceAPIWithIdentity" with parameters "object-storage" and "{testUserRead}"
    And "{result}" is not an error
    And I attach "{result}" to the test output as "read-storage-service.json"
    And I refer to "{result}" as "userStorage"
    When I call "{userStorage}" with "ListObjects" with parameter "{ResourceName}"
    Then "{result}" is not an error
    And I attach "{result}" to the test output as "read-list-objects-result.json"
