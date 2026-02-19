@PerService @CCC.ObjStor @CCC.ObjStor.CN06 @tlp-amber @tlp-red
Feature: CCC.ObjStor.CN06.AR01 - Access Logs in Separate Data Store
  As a security administrator
  I want to ensure access logs are stored separately from the bucket they monitor
  So that log integrity is protected

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy
  Scenario: Access logs are stored in a separate data store
    When I attempt policy check "object-storage-access-logging" for control "CCC.ObjStor.CN06" assessment requirement "AR01" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true
