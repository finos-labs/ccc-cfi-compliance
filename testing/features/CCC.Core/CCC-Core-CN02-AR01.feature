@CCC.Core @CCC.Core.CN02 @tlp-green @tlp-amber @tlp-red
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
