@PerService @object-storage @CCC.ObjStor @tlp-clear @tlp-green @tlp-amber @tlp-red @CCC.ObjStor.CN01.AR03
Feature: CCC.ObjStor.CN01.AR03
  As a security administrator
  I want to prevent any requests to create buckets using untrusted KMS keys
  So that data encryption integrity and availability are protected against unauthorized encryption

  Background:
    Given a cloud api for "{Instance}" in "api"
    And I call "{api}" with "GetServiceAPI" using argument "object-storage"
    And I refer to "{result}" as "storage"
    And I call "{api}" with "GetServiceAPI" using argument "iam"
    And I refer to "{result}" as "iamService"

  @Behavioural
  Scenario: Service prevents creating bucket with no access
    Given I call "{iamService}" with "ProvisionUserWithAccess" using arguments "test-user-no-access", "{UID}", and "none"
    And I refer to "{result}" as "testUserNoAccess"
    And I attach "{result}" to the test output as "no-access-user-identity.json"
    And I call "{api}" with "GetServiceAPIWithIdentity" using arguments "object-storage", "{testUserNoAccess}", and "{false}"
    And "{result}" is not an error
    And I refer to "{result}" as "userStorage"
    When I call "{userStorage}" with "CreateBucket" using argument "test-bucket-no-access"
    Then "{result}" is an error
    And I attach "{result}" to the test output as "no-access-create-bucket-error.txt"

  @Behavioural
  Scenario: Service allows creating bucket with write access
    Given I call "{iamService}" with "ProvisionUserWithAccess" using arguments "test-user-write", "{UID}", and "write"
    And I refer to "{result}" as "testUserWrite"
    And I attach "{result}" to the test output as "write-user-identity.json"
    And I call "{api}" with "GetServiceAPIWithIdentity" using arguments "object-storage", "{testUserWrite}", and "{true}"
    And "{result}" is not an error
    And I attach "{result}" to the test output as "write-storage-service.json"
    And I refer to "{result}" as "userStorage"
    When I call "{userStorage}" with "CreateBucket" using argument "test-bucket-write"
    Then "{result}" is not an error
    And I attach "{result}" to the test output as "write-create-bucket-result.json"
    And I call "{storage}" with "DeleteBucket" using argument "{result.ID}"
