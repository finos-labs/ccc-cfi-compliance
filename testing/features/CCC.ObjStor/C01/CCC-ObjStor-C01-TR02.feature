@PerService @CCC.ObjStor @tlp-amber @tlp-red
Feature: CCC.ObjStor.C01.TR02
  As a security administrator
  I want to prevent any requests to read protected objects using untrusted KMS keys
  So that data encryption integrity and availability are protected against unauthorized encryption

  Background:
    Given a cloud api for "{Provider}" in "api"
    And I call "{api}" with "GetServiceAPI" with parameter "object-storage"
    And I refer to "{result}" as "storage"
    And I call "{api}" with "GetServiceAPI" with parameter "iam"
    And I refer to "{result}" as "iamService"

  Scenario: Service prevents reading object with untrusted KMS key
    Given I call "{iamService}" with "ProvisionUser" with parameter "test-user-untrusted"
    And I refer to "{result}" as "testUserUntrusted"
    And I call "{iamService}" with "SetAccess" with parameters "{testUserUntrusted}", "{UID}" and "none"
    And I call "{api}" with "GetServiceAPIWithIdentity" with parameters "object-storage" and "{testUserUntrusted}"
    And I refer to "{result}" as "userStorage"
    And I call "{storage}" with "CreateBucket" with parameter "test-bucket-obj-untrusted"
    And I refer to "{result}" as "bucketUntrusted"
    And I call "{storage}" with "CreateObject" with parameters "{bucketUntrusted.ID}", "test-object.txt" and "test content"
    When I call "{userStorage}" with "ReadObject" with parameters "{bucketUntrusted.ID}" and "test-object.txt"
    Then "{result}" is an error
    And I call "{storage}" with "DeleteObject" with parameters "{bucketUntrusted.ID}" and "test-object.txt"
    And I call "{storage}" with "DeleteBucket" with parameter "{bucketUntrusted.ID}"

  Scenario: Service allows reading object with trusted KMS key
    Given I call "{iamService}" with "ProvisionUser" with parameter "test-user-trusted"
    And I refer to "{result}" as "testUserTrusted"
    And I call "{iamService}" with "SetAccess" with parameters "{testUserTrusted}", "{UID}" and "read"
    And I call "{api}" with "GetServiceAPIWithIdentity" with parameters "object-storage" and "{testUserTrusted}"
    And I refer to "{result}" as "userStorage"
    And I call "{storage}" with "CreateBucket" with parameter "test-bucket-obj-trusted"
    And I refer to "{result}" as "bucketTrusted"
    And I call "{storage}" with "CreateObject" with parameters "{bucketTrusted.ID}", "test-object.txt" and "test content"
    When I call "{userStorage}" with "ReadObject" with parameters "{bucketTrusted.ID}" and "test-object.txt"
    Then "{result}" is not nil
    And I call "{storage}" with "DeleteObject" with parameters "{bucketTrusted.ID}" and "test-object.txt"
    And I call "{storage}" with "DeleteBucket" with parameter "{bucketTrusted.ID}"

  Scenario: Cleanup
    Given I call "{iamService}" with "DestroyUser" with parameter "{testUserUntrusted}"
    And I call "{iamService}" with "DestroyUser" with parameter "{testUserTrusted}"
