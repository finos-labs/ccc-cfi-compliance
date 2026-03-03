@PerService @object-storage @CCC.ObjStor @tlp-amber @tlp-red @CCC.ObjStor.CN02.AR01
Feature: CCC.ObjStor.CN02.AR01 - Uniform Bucket-Level Access (Consistent Allow)
  When a permission set is allowed for an object in a bucket,
  the service MUST allow the same permission set to access all objects in the same bucket.
  
  This ensures uniform bucket-level access is enforced, preventing ad-hoc object-level permissions.

  Background:
    Given a cloud api for "{Instance}" in "api"
    And I call "{api}" with "GetServiceAPI" using argument "object-storage"
    And I refer to "{result}" as "storage"
    And I call "{api}" with "GetServiceAPI" using argument "iam"
    And I refer to "{result}" as "iamService"

  @Behavioural
  Scenario: Service enforces uniform bucket-level access by rejecting object-level permissions
    When I call "{storage}" with "CreateObject" using arguments "{ResourceName}", "test-object.txt", and "test data"
    Then "{result}" is not an error
    Given I call "{iamService}" with "ProvisionUserWithAccess" using arguments "test-user-read", "{UID}", and "read"
    And I refer to "{result}" as "testUserRead"
    And I attach "{result}" to the test output as "read-user-identity.json"
    And I call "{api}" with "GetServiceAPIWithIdentity" using arguments "object-storage", "{testUserRead}", and "{true}"
    And "{result}" is not an error
    And I refer to "{result}" as "userStorage"
    When I call "{userStorage}" with "ReadObject" using arguments "{ResourceName}" and "test-object.txt"
    Then "{result}" is not an error
    When I call "{storage}" with "SetObjectPermission" using arguments "{ResourceName}", "test-object.txt", and "none"
    Then "{result}" is an error
    And I attach "{result}" to the test output as "set-object-permission-error.txt"
    When I call "{userStorage}" with "ReadObject" using arguments "{ResourceName}" and "test-object.txt"
    Then "{result}" is not an error

  @Policy
  Scenario: Test policy for uniform access
    When I attempt policy check "uniform-bucket-level-access" for control "CCC.ObjStor.CN02" assessment requirement "AR01" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true
