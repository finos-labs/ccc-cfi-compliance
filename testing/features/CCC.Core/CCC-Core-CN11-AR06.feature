@CCC.Core @CCC.Core.CN11 @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN11.AR06 - Encryption Key Rotation (90 days)
  As a security administrator
  I want to ensure encryption keys are rotated within 90 days
  So that stringent key rotation requirements are met

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @CCC.ObjStor
  Scenario: Object storage key rotation compliance (90 days)
    When I attempt policy check "object-storage-key-rotation-90" for control "CCC.Core.CN11" assessment requirement "AR06" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true
