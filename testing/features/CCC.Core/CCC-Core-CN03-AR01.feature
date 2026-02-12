@CCC.Core @CCC.Core.CN03 @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN03.AR01 - Multi-Factor Authentication for Destructive Operations
  As a security administrator
  I want to ensure destructive operations require multi-factor authentication
  So that accidental or malicious deletions are prevented

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @CCC.ObjStor
  Scenario: Object storage delete protection compliance
    When I attempt policy check "object-storage-delete-protection" for control "CCC.Core.CN03" assessment requirement "AR01" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true
