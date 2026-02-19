@PerService @CCC.Core @CCC.Core.CN11 @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN11.AR05 - Encryption Key Rotation (365 days)
  As a security administrator
  I want to ensure encryption keys are rotated within 365 days
  So that annual key rotation requirements are met

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @CCC.ObjStor
  Scenario: Object storage key rotation compliance (365 days)
    When I attempt policy check "object-storage-key-rotation-365" for control "CCC.Core.CN11" assessment requirement "AR05" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true
