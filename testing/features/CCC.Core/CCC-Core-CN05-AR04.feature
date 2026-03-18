@PerService @CCC.Core @CCC.Core.CN05 @tlp-amber @tlp-red
Feature: CCC.Core.CN05.AR04 - Block Unauthorized External Data Requests
  As a security administrator
  I want to ensure data requests from outside trust perimeter are blocked
  So that data exfiltration is prevented

  Background:
    Given a cloud api for "{Instance}" in "api"

  @Policy @object-storage
  Scenario: External unauthorized data requests are blocked
    When I attempt policy check "object-storage-block-public-read" for control "CCC.Core.CN05" assessment requirement "AR04" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true
