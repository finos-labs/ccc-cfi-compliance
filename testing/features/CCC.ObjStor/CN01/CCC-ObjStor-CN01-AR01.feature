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
    Given I call "{iamService}" with "ProvisionUser" with parameter "test-user-trusted"
    And I refer to "{result}" as "testUserTrusted"
    And I call "{iamService}" with "SetAccess" with parameters "{testUserTrusted}", "{UID}" and "read"

  Scenario: Service prevents reading bucket with untrusted KMS key
    Given I attach "{testUserUntrusted}" to the test output
    And I call "{api}" with "GetServiceAPIWithIdentity" with parameters "object-storage" and "{testUserUntrusted}"
    And "{result}" is not an error
    And I refer to "{result}" as "userStorage"
    When I call "{userStorage}" with "ListObjects" with parameter "{ResourceName}"
    Then "{result}" is an error
    And I attach "{result}" to the test output

  Scenario: Service allows reading bucket with trusted KMS key
    Given I attach "{testUserTrusted}" to the test output
    And I call "{api}" with "GetServiceAPIWithIdentity" with parameters "object-storage" and "{testUserTrusted}"
    And "{result}" is not an error
    And I attach "{result}" to the test output
    And I refer to "{result}" as "userStorage"
    When I call "{userStorage}" with "ListObjects" with parameter "{ResourceName}"
    Then "{result}" is not an error
    And I attach "{result}" to the test output

  Scenario: Cleanup
    Given I call "{iamService}" with "DestroyUser" with parameter "{testUserUntrusted}"
    Then "{result}" is not an error
    And I call "{iamService}" with "DestroyUser" with parameter "{testUserTrusted}"
    Then "{result}" is not an error
