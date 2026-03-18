@PerService @CCC.Core @CCC.Core.CN07 @tlp-clear @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN07.AR02 - Log Enumeration Activities
  As a security administrator
  I want to ensure enumeration activities are logged
  So that reconnaissance attempts can be investigated

  Background:
    Given a cloud api for "{Instance}" in "api"

  @Policy @object-storage
  Scenario: Enumeration activities are logged
    When I attempt policy check "enumeration-logging-policy" for control "CCC.Core.CN07" assessment requirement "AR02" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true

  @Behavioural @object-storage
  Scenario: Enumeration logging cannot be verified automatically
    # Verifying enumeration activities are logged requires performing operations
    # and querying cloud audit logs - cross-service integration (object-storage +
    # logging) and log retrieval timing make full automation complex.
    #
    # Manual verification steps:
    # 1. Perform enumeration activity (e.g., ListBuckets, ListObjects)
    # 2. Query cloud audit logs for corresponding entries
    # 3. Confirm log entries contain required fields for investigation
    Then no-op required
