@PerService @CCC.Core @CCC.Core.CN14 @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN14.AR01 - Data Immutability for Disaster Recovery
  As a security administrator
  I want to ensure backup data cannot be modified or deleted within retention period
  So that disaster recovery data integrity is guaranteed

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @CCC.ObjStor
  Scenario: Object storage immutability compliance
    When I attempt policy check "object-storage-immutability" for control "CCC.Core.CN14" assessment requirement "AR01" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true
