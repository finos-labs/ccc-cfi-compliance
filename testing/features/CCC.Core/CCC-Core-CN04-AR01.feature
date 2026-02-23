@PerService @CCC.Core @CCC.Core.CN04 @tlp-clear @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN04.AR01 - Log Administrative Access Attempts
  As a security administrator
  I want to ensure all administrative access attempts are logged
  So that audit trails are maintained for compliance

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @object-storage @vpc
  Scenario: Admin logging compliance
    When I attempt policy check "admin-logging" for control "CCC.Core.CN04" assessment requirement "AR01" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true

  @Behavioural @object-storage
  Scenario: Verify admin actions are logged with identity and timestamp
    Given I call "{api}" with "GetServiceAPI" using argument "object-storage"
    And I refer to "{result}" as "storage"
    When I call "{storage}" with "UpdateBucketPolicy" using arguments "{ResourceName}" and "test-policy-update"
    Then "{result}" is not an error
    And I refer to "{result}" as "policyUpdateResult"
    And I attach "{policyUpdateResult}" to the test output as "Policy Update Result"
    When I call "{storage}" with "QueryAdminLogs" using arguments "{ResourceName}" and "60"
    Then "{result}" is not an error
    And I refer to "{result}" as "adminLogs"
    And I attach "{adminLogs}" to the test output as "Admin Activity Logs"
    And "{adminLogs}" is an array of objects with at least the following contents
      | Identity  | Action            |
      | .+        | .*Policy.*        |
