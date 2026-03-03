@PerService @CCC.Core @CCC.Core.CN07 @tlp-amber @tlp-red
Feature: CCC.Core.CN07.AR01 - Publish Enumeration Activity Events
  As a security administrator
  I want to ensure enumeration activities trigger events to monitored channels
  So that reconnaissance attempts are detected

  Background:
    Given a cloud api for "{Instance}" in "api"

  @Policy @object-storage
  Scenario: Enumeration activities publish events to monitored channels
    When I attempt policy check "enumeration-monitoring-policy" for control "CCC.Core.CN07" assessment requirement "AR01" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true
