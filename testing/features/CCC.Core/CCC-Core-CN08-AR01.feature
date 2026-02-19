@PerService @CCC.Core @CCC.Core.CN08 @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN08.AR01 - Data Replication and Redundancy
  As a security administrator
  I want to ensure data is replicated to a physically separate data center
  So that disaster recovery requirements are met

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @CCC.ObjStor
  Scenario: Object storage replication compliance
    When I attempt policy check "object-storage-replication" for control "CCC.Core.CN08" assessment requirement "AR01" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true
