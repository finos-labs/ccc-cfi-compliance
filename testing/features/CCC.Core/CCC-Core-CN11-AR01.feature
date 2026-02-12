@CCC.Core @CCC.Core.CN11 @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN11.AR01 - Encryption Key Algorithm Validation
  As a security administrator
  I want to ensure encryption keys use industry-standard cryptographic algorithms
  So that data protection meets current security standards

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @CCC.ObjStor
  Scenario: Object storage key algorithm compliance
    When I attempt policy check "object-storage-key-algorithm" for control "CCC.Core.CN11" assessment requirement "AR01" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true
