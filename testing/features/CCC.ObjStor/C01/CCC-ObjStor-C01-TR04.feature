@PerService @CCC.ObjStor @tlp-clear @tlp-green @tlp-amber @tlp-red
Feature: CCC.ObjStor.C01.TR04
  As a security administrator
  I want to prevent any requests to write to objects using untrusted KMS keys
  So that data encryption integrity and availability are protected against unauthorized encryption

  Background:
    Given a cloud api for "{Provider}" in "api"
    And I call "{api}" with "GetServiceAPI" with parameter "object-storage"
    And I refer to "{result}" as "storage"
    And I call "{api}" with "GetServiceAPI" with parameter "iam"
    And I refer to "{result}" as "iamService"
    And I call "{storage}" with "CreateBucket" with parameter "test-bucket-obj-write"
    And I refer to "{result}" as "bucket"

  Scenario: Service prevents writing object with untrusted KMS key
    Given I call "{iamService}" with "ProvisionUser" with parameter "test-user-untrusted"
    And I refer to "{result}" as "testUserUntrusted"
    And I call "{iamService}" with "SetAccess" with parameters "{testUserUntrusted}", "{UID}" and "read"
    And I call "{api}" with "GetServiceAPIWithIdentity" with parameters "object-storage" and "{testUserUntrusted}"
    And I refer to "{result}" as "userStorage"
    When I call "{userStorage}" with "CreateObject" with parameters "{bucket.ID}", "test-object.txt" and "test content"
    Then "{result}" is an error
    And I call "{iamService}" with "DestroyUser" with parameter "{testUserUntrusted}"

  Scenario: Service allows writing object with trusted KMS key
    Given I call "{iamService}" with "ProvisionUser" with parameter "test-user-trusted"
    And I refer to "{result}" as "testUserTrusted"
    And I call "{iamService}" with "SetAccess" with parameters "{testUserTrusted}", "{uid}" and "write"
    And I call "{api}" with "GetServiceAPIWithIdentity" with parameters "object-storage" and "{testUserTrusted}"
    And I refer to "{result}" as "userStorage"
    When I call "{userStorage}" with "CreateObject" with parameters "{bucket.ID}", "test-object.txt" and "test content"
    Then "{result.ID}" is "test-object.txt"
    And I call "{iamService}" with "DestroyUser" with parameter "{testUserTrusted}"
    And I call "{storage}" with "DeleteObject" with parameters "{bucket.ID}" and "test-object.txt"

  Scenario: Cleanup
    Given I call "{storage}" with "DeleteBucket" with parameter "{bucket.ID}"
    And I call "{iamService}" with "DestroyUser" with parameter "{testUserTrusted}"
    And I call "{iamService}" with "DestroyUser" with parameter "{testUserUntrusted}"
