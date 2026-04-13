@PerService @CCC.Core @CCC.Core.CN09 @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN09.AR01 - Access Logging Separation
  As a security administrator
  I want to ensure access logs are stored separately from the resources they monitor
  So that log integrity is protected

  Background:
    Given a cloud api for "{Instance}" in "api"

  @Policy @object-storage
  Scenario: Object storage access logging compliance
    When I attempt policy check "object-storage-access-logging" for control "CCC.Core.CN09" assessment requirement "AR01" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true

  @Policy @vpc
  Scenario: VPC flow log separation compliance
    When I attempt policy check "vpc-flow-log-separation" for control "CCC.Core.CN09" assessment requirement "AR01" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true
