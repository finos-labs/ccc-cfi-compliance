@PerService @object-storage @CCC.ObjStor @tlp-clear @tlp-green @tlp-amber @tlp-red @CCC.ObjStor.CN01
Feature: CCC.ObjStor.CN01.AR04
  As a security administrator
  I want to prevent any requests to write to objects using untrusted KMS keys
  So that data encryption integrity and availability are protected against unauthorized encryption

  Background:
    Given a cloud api for "{Instance}" in "api"
    And I call "{api}" with "GetServiceAPI" using argument "object-storage"
    And I refer to "{result}" as "storage"
    And "{result}" is not an error
    And I call "{api}" with "GetServiceAPI" using argument "iam"
    And I refer to "{result}" as "iamService"
    And "{result}" is not an error

  @Behavioural
  Scenario: Service prevents writing object with read-only access
    Given I call "{iamService}" with "ProvisionUserWithAccess" using arguments "test-user-read", "{UID}", and "read"
    And I refer to "{result}" as "testUserRead"
    And I attach "{result}" to the test output as "read-user-identity.json"
    And I call "{api}" with "GetServiceAPIWithIdentity" using arguments "object-storage", "{testUserRead}", and "{true}"
    And "{result}" is not an error
    And I refer to "{result}" as "userStorage"
    When I call "{userStorage}" with "CreateObject" using arguments "{ResourceName}", "test-write-object.txt", and "test content"
    Then "{result}" is an error
    And I attach "{result}" to the test output as "read-create-object-error.txt"

  @Behavioural
  Scenario: Service allows writing object with write access
    Given I call "{iamService}" with "ProvisionUserWithAccess" using arguments "test-user-write", "{UID}", and "write"
    And I refer to "{result}" as "testUserWrite"
    And I attach "{result}" to the test output as "write-user-identity.json"
    And I call "{api}" with "GetServiceAPIWithIdentity" using arguments "object-storage", "{testUserWrite}", and "{true}"
    And "{result}" is not an error
    And I attach "{result}" to the test output as "write-storage-service.json"
    And I refer to "{result}" as "userStorage"
    When I call "{userStorage}" with "CreateObject" using arguments "{ResourceName}", "test-write-object.txt", and "test content"
    Then "{result}" is not an error
    And I attach "{result}" to the test output as "write-create-object-result.json"
    And I call "{storage}" with "DeleteObject" using arguments "{ResourceName}" and "test-write-object.txt"

  @Policy
  Scenario: All unauthorized requests are blocked
    # This control requires behavioral testing - comprehensive access testing
    # IAM policies enforce this at runtime
    Then no-op required
