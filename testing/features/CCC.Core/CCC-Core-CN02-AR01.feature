@PerService @CCC.Core @CCC.Core.CN02 @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN02.AR01 - Data Encryption at Rest
  As a security administrator
  I want to ensure all stored data is encrypted using industry-standard methods
  So that data confidentiality is protected

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @CCC.ObjStor
  Scenario: Object storage encryption compliance
    When I attempt policy check "object-storage-encryption" for control "CCC.Core.CN02" assessment requirement "AR01" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true

  @Behavioural @CCC.ObjStor
  Scenario: Verify objects are encrypted at rest
    Given I call "{api}" with "GetServiceAPI" with parameter "object-storage"
    And I refer to "{result}" as "storage"
    When I call "{storage}" with "CreateObject" with parameters "{ResourceName}", "test-encryption-check.txt" and "encryption test data"
    Then "{result}" is not an error
    And I refer to "{result}" as "uploadResult"
    And "{uploadResult.Encryption}" is not nil
    And "{uploadResult.EncryptionAlgorithm}" is "AES256"
    And I attach "{uploadResult}" to the test output as "Upload Result with Encryption Details"
    And I call "{storage}" with "DeleteObject" with parameters "{ResourceName}" and "test-encryption-check.txt"
