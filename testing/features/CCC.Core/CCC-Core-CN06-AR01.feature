@CCC.Core @CCC.Core.CN06 @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN06.AR01 - Resource Location Compliance
  As a security administrator
  I want to ensure cloud resources are deployed in approved regions
  So that data residency and sovereignty requirements are met

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy
  Scenario: Policy checks
    When I run policy checks for control "CCC.Core.CN06" assessment requirement "AR01" for service "{ServiceType}" on resource "{ResourceName}"
    Then "{result}" is true
