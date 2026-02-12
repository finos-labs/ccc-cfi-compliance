@CCC.Core @CCC.Core.CN11 @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN11.AR03 - Customer-Managed Encryption Keys
  As a security administrator
  I want to ensure customer-managed encryption keys (CMEKs) are used
  So that key management remains under organizational control

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @CCC.ObjStor
  Scenario: Object storage CMEK compliance
    When I attempt policy check "object-storage-cmek" for control "CCC.Core.CN11" assessment requirement "AR03" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true
