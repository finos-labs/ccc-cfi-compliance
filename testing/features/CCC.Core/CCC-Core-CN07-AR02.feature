@PerService @CCC.Core @CCC.Core.CN07 @tlp-clear @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN07.AR02 - Log Enumeration Activities
  As a security administrator
  I want to ensure enumeration activities are logged
  So that reconnaissance attempts can be investigated

  Background:
    Given a cloud api for "{Instance}" in "api"

  @Policy @object-storage
  Scenario: Enumeration activities are logged
    When I attempt policy check "enumeration-logging-policy" for control "CCC.Core.CN07" assessment requirement "AR02" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true
