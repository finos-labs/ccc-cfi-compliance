@PerService @CCC.ObjStor @tlp-amber @tlp-red
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
    Given I call "{iamService}" with "ProvisionUser" with parameter "test-user-untrusted"
    And I refer to "{result}" as "testUserUntrusted"
    And I call "{iamService}" with "SetAccess" with parameters "{testUserUntrusted}", "{UID}" and "none"
    And I attach "{result}" to the test output as "untrusted-user-policy.json"
    Given I call "{iamService}" with "ProvisionUser" with parameter "test-user-trusted"
    And I refer to "{result}" as "testUserTrusted"
    And I call "{iamService}" with "SetAccess" with parameters "{testUserTrusted}", "{UID}" and "read"
    And I attach "{result}" to the test output as "trusted-user-policy.json"

  Scenario: Service prevents reading bucket with untrusted KMS key
    Given I attach "{testUserUntrusted}" to the test output as "untrusted-user-identity.json"
    And I call "{api}" with "GetServiceAPIWithIdentity" with parameters "object-storage" and "{testUserUntrusted}"
    And "{result}" is not an error
    And I refer to "{result}" as "userStorage"
    And we wait for a period of "10000" ms
    When I call "{userStorage}" with "ListObjects" with parameters "{ResourceName}" and "{Region}"
    Then "{result}" is an error
    And I attach "{result}" to the test output as "untrusted-list-error.txt"

  Scenario: Service allows reading bucket with trusted KMS key
    Given I attach "{testUserTrusted}" to the test output as "trusted-user-identity.json"
    And I call "{api}" with "GetServiceAPIWithIdentity" with parameters "object-storage" and "{testUserTrusted}"
    And "{result}" is not an error
    And I attach "{result}" to the test output as "trusted-storage-service.json"
    And I refer to "{result}" as "userStorage"
    And we wait for a period of "10000" ms
    When I call "{userStorage}" with "ListObjects" with parameters "{ResourceName}" and "{Region}"
    Then "{result}" is not an error
    And I attach "{result}" to the test output as "trusted-list-objects-result.json"
