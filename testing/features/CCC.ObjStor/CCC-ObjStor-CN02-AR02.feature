@CCC.ObjStor @tlp-amber @tlp-red @CCC.ObjStor.CN02.AR02
Feature: CCC.ObjStor.CN02.AR02 - Uniform Bucket-Level Access (Consistent Deny)
  When a permission set is denied for an object in a bucket,
  the service MUST deny the same permission set to access all objects in the same bucket.
  
  This ensures uniform bucket-level access is enforced, preventing ad-hoc object-level permissions.

  Background:
    Given a cloud api for "{Provider}" in "api"
    And I call "{api}" with "GetServiceAPI" with parameter "object-storage"
    And I refer to "{result}" as "storage"
    And I call "{api}" with "GetServiceAPI" with parameter "iam"
    And I refer to "{result}" as "iamService"

  Scenario: Service enforces uniform bucket-level access denial
    When I call "{storage}" with "CreateObject" with parameters "{ResourceName}", "test-object.txt" and "test data"
    Then "{result}" is not an error
    Given I call "{iamService}" with "ProvisionUserWithAccess" with parameters "test-user-no-access", "{UID}" and "none"
    And I refer to "{result}" as "testUserNoAccess"
    And I attach "{result}" to the test output as "no-access-user-identity.json"
    And I call "{api}" with "GetServiceAPIWithIdentity" with parameters "object-storage", "{testUserNoAccess}" and "{false}"
    And "{result}" is not an error
    And I refer to "{result}" as "userStorage"
    When I call "{userStorage}" with "ReadObject" with parameters "{ResourceName}" and "test-object.txt"
    Then "{result}" is an error
    When I call "{storage}" with "SetObjectPermission" with parameters "{ResourceName}", "test-object.txt" and "read"
    Then "{result}" is an error
    And I attach "{result}" to the test output as "set-object-permission-error.txt"
    When I call "{userStorage}" with "ReadObject" with parameters "{ResourceName}" and "test-object.txt"
    Then "{result}" is an error
