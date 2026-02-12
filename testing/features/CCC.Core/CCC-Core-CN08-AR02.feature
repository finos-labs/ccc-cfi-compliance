@CCC.Core @CCC.Core.CN08 @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN08.AR02 - Replication Status Visibility
  As a security administrator
  I want to ensure replication status is accurately represented
  So that data synchronization can be monitored

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @CCC.ObjStor
  Scenario: Object storage replication status is visible
    When I attempt policy check "object-storage-replication-status" for control "CCC.Core.CN08" assessment requirement "AR02" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true
