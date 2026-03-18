@PerService @CCC.Core @CCC.Core.CN08 @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN08.AR02 - Replication Status Visibility
  As a security administrator
  I want to ensure replication status is accurately represented
  So that data synchronization can be monitored

  Background:
    Given a cloud api for "{Instance}" in "api"
    And I call "{api}" with "GetServiceAPI" using argument "object-storage"
    And I refer to "{result}" as "storage"

  @Policy @object-storage
  Scenario: Object storage replication status is visible
    When I attempt policy check "object-storage-replication-status" for control "CCC.Core.CN08" assessment requirement "AR02" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true

  @Behavioural @object-storage
  Scenario: Replication status can be retrieved for monitoring
    When I call "{storage}" with "GetReplicationStatus" using argument "{ResourceName}"
    And I refer to "{result}" as "replicationStatus"
    And I attach "{replicationStatus}" to the test output as "Replication Status"
    And I refer to "{replicationStatus.Locations}" as "locations"
    Then "{repllocations}" is an array of objects with at least the following contents
      | value           |
      | {locations[0]}  |
