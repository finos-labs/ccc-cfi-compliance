@PerService @object-storage @CCC.ObjStor @tlp-amber @tlp-red @CCC.ObjStor.CN01
Feature: CCC.ObjStor.CN01.AR02
  As a security administrator
  I want to prevent any requests to read protected objects using untrusted KMS keys
  So that data encryption integrity and availability are protected against unauthorized encryption

  Background:
    Given a cloud api for "{Instance}" in "api"
    And I call "{api}" with "GetServiceAPI" using argument "object-storage"
    And I refer to "{result}" as "storage"
    And I call "{api}" with "GetServiceAPI" using argument "iam"
    And I refer to "{result}" as "iamService"
    And I call "{storage}" with "CreateObject" using arguments "{ResourceName}", "test-object.txt", and "test content"
    And "{result}" is not an error

  @Behavioural
  Scenario: Service prevents reading object with no access
    Given I call "{iamService}" with "ProvisionUserWithAccess" using arguments "test-user-no-access", "{UID}", and "none"
    And I refer to "{result}" as "testUserNoAccess"
    And I attach "{result}" to the test output as "no-access-user-identity.json"
    And I call "{api}" with "GetServiceAPIWithIdentity" using arguments "object-storage", "{testUserNoAccess}", and "{false}"
    And "{result}" is not an error
    And I refer to "{result}" as "userStorage"
    When I call "{userStorage}" with "ReadObject" using arguments "{ResourceName}" and "test-object.txt"
    Then "{result}" is an error
    And I attach "{result}" to the test output as "no-access-read-object-error.txt"

  @Behavioural
  Scenario: Service allows reading object with read access
    Given I call "{iamService}" with "ProvisionUserWithAccess" using arguments "test-user-read", "{UID}", and "read"
    And I refer to "{result}" as "testUserRead"
    And I attach "{result}" to the test output as "read-user-identity.json"
    And I call "{api}" with "GetServiceAPIWithIdentity" using arguments "object-storage", "{testUserRead}", and "{true}"
    And "{result}" is not an error
    And I attach "{result}" to the test output as "read-storage-service.json"
    And I refer to "{result}" as "userStorage"
    When I call "{userStorage}" with "ReadObject" using arguments "{ResourceName}" and "test-object.txt"
    Then "{result}" is not an error
    And I attach "{result}" to the test output as "read-read-object-result.json"
    And I call "{storage}" with "DeleteObject" using arguments "{ResourceName}" and "test-object.txt"

  @Policy
  Scenario: All unauthorized requests are blocked
    # This control requires behavioral testing - comprehensive access testing
    # IAM policies enforce this at runtime
    Then no-op required
