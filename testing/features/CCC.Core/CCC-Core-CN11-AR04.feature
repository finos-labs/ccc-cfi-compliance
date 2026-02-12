@PerService @CCC.Core @CCC.Core.CN11 @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN11.AR04 - Encryption Key Access Control
  As a security administrator
  I want to ensure encryption key access follows least privilege principles
  So that key access is restricted to authorized entities only

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @CCC.ObjStor
  Scenario: Object storage key policy compliance
    When I attempt policy check "object-storage-key-policy" for control "CCC.Core.CN11" assessment requirement "AR04" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true
