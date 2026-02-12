@CCC.ObjStor @CCC.ObjStor.CN05 @tlp-clear @tlp-green @tlp-amber @tlp-red
Feature: CCC.ObjStor.CN05.AR01 - Versioning with Unique Identifiers
  As a security administrator
  I want to ensure objects are stored with unique identifiers
  So that version tracking is enabled

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy
  Scenario: Objects are stored with unique version identifiers
    When I attempt policy check "object-storage-versioning" for control "CCC.ObjStor.CN05" assessment requirement "AR01" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true
