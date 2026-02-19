@PerService @CCC.Core @CCC.Core.CN06 @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN06.AR01 - Resource Location Compliance
  As a security administrator
  I want to ensure cloud resources are deployed in approved regions
  So that data residency and sovereignty requirements are met

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @CCC.ObjStor
  Scenario: Object storage region compliance
    When I attempt policy check "object-storage-region" for control "CCC.Core.CN06" assessment requirement "AR01" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true

  @Policy @CCC.VPC
  Scenario: VPC region compliance
    When I set the VPC peer to...
    When I attempt policy check "vpc-region" for control "CCC.Core.CN06" assessment requirement "AR01" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true
