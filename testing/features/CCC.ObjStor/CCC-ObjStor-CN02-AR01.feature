@PerService @CCC.ObjStor @tlp-amber @tlp-red @CCC.ObjStor.CN02.AR01
Feature: CCC.ObjStor.CN02.AR01 - Uniform Bucket-Level Access (Consistent Allow)
  When a permission set is allowed for an object in a bucket,
  the service MUST allow the same permission set to access all objects in the same bucket.
  
  This ensures uniform bucket-level access is enforced, preventing ad-hoc object-level permissions.

  Background:
    Given a cloud api for "{Provider}" in "api"
    And I call "{api}" with "GetServiceAPI" with parameter "object-storage"
    And I refer to "{result}" as "storage"
    And I call "{api}" with "GetServiceAPI" with parameter "iam"
    And I refer to "{result}" as "iamService"

  Scenario: Service enforces uniform bucket-level access by rejecting object-level permissions
    When I call "{storage}" with "CreateObject" with parameters "{ResourceName}", "test-object.txt" and "test data"
    Then "{result}" is not an error
    Given I call "{iamService}" with "ProvisionUserWithAccess" with parameters "test-user-read", "{UID}" and "read"
    And I refer to "{result}" as "testUserRead"
    And I attach "{result}" to the test output as "read-user-identity.json"
    And I call "{api}" with "GetServiceAPIWithIdentity" with parameters "object-storage", "{testUserRead}" and "{true}"
    And "{result}" is not an error
    And I refer to "{result}" as "userStorage"
    When I call "{userStorage}" with "ReadObject" with parameters "{ResourceName}" and "test-object.txt"
    Then "{result}" is not an error
    When I call "{storage}" with "SetObjectPermission" with parameters "{ResourceName}", "test-object.txt" and "none"
    Then "{result}" is an error
    And I attach "{result}" to the test output as "set-object-permission-error.txt"
    When I call "{userStorage}" with "ReadObject" with parameters "{ResourceName}" and "test-object.txt"
    Then "{result}" is not an error
