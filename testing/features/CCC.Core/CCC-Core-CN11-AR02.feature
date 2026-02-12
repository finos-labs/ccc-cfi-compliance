@CCC.Core @CCC.Core.CN11 @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN11.AR02 - Encryption Key Rotation (180 days)
  As a security administrator
  I want to ensure encryption keys are rotated within 180 days
  So that key compromise risks are minimized

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @CCC.ObjStor
  Scenario: Object storage key rotation compliance (180 days)
    When I attempt policy check "object-storage-key-rotation" for control "CCC.Core.CN11" assessment requirement "AR02" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true
