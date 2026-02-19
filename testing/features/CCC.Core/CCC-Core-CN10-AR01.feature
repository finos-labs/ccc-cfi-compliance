@PerService @CCC.Core @CCC.Core.CN10 @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN10.AR01 - Replication Destination Trust
  As a security administrator
  I want to ensure data replication only occurs to trusted destinations
  So that data sovereignty and trust perimeter requirements are met

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @CCC.ObjStor
  Scenario: Object storage replication destination compliance
    When I attempt policy check "object-storage-replication-destination" for control "CCC.Core.CN10" assessment requirement "AR01" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true
