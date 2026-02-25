@PerService @CCC.Core @CCC.Core.CN05 @tlp-clear @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN05.AR03 - Block Cross-Tenant Access
  As a security administrator
  I want to ensure cross-tenant access is blocked unless explicitly allowed
  So that multi-tenant isolation is maintained

  Background:
    Given a cloud api for "{Instance}" in "api"

  @Policy @object-storage
  Scenario: Cross-tenant access is blocked without explicit allowlist
    When I attempt policy check "object-storage-cross-tenant-block" for control "CCC.Core.CN05" assessment requirement "AR03" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true
