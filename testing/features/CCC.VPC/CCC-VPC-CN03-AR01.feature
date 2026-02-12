@CCC.VPC @tlp-amber @tlp-red
Feature: CCC.VPC.CN03.AR01
  As a security administrator
  I want unauthorized VPC peering requests to be denied
  So that peering is restricted to explicitly approved destinations

  Background:
    Given a cloud api for "{Provider}" in "api"
    And I call "{api}" with "GetServiceAPI" with parameter "vpc"
    And I refer to "{result}" as "vpc"

  @CCC.VPC.CN03.AR01 @CN03.DISALLOWED
  Scenario: Disallowed peer target must be denied
    When I call "{vpc}" with "AttemptDisallowedPeeringDryRun" with parameter "{UID}"
    And I refer to "{result}" as "disallowedPeeringDryRun"
    And I attach "{disallowedPeeringDryRun}" to the test output as "cn03-disallowed-peering-dry-run.json"
    And I call "{vpc}" with "SummarizePeeringOutcomeCompact" with parameters "{disallowedPeeringDryRun}" and "disallowed"
    And I refer to "{result}" as "disallowedCompactSummary"
    And I attach "{disallowedCompactSummary}" to the test output as "cn03-disallowed-summary-compact.json"
    Then "{disallowedCompactSummary.Mode}" is "disallowed"
    And "{disallowedCompactSummary.Verdict}" is "PASS"
    And "{disallowedCompactSummary.ResultClass}" is "PASS"
    And "{disallowedCompactSummary.DryRunAllowed}" is "false"

  @CN03.ALLOWED @CN03.SANITY
  Scenario: Allowed peer target sanity check should be dry-run allowed
    When I call "{vpc}" with "AttemptDisallowedPeeringDryRun" with parameter "{UID}"
    And I refer to "{result}" as "allowedPeeringDryRun"
    And I attach "{allowedPeeringDryRun}" to the test output as "cn03-allowed-peering-dry-run.json"
    And I call "{vpc}" with "SummarizePeeringOutcomeCompact" with parameters "{allowedPeeringDryRun}" and "allowed"
    And I refer to "{result}" as "allowedCompactSummary"
    And I attach "{allowedCompactSummary}" to the test output as "cn03-allowed-summary-compact.json"
    Then "{allowedCompactSummary.Mode}" is "allowed"
    And "{allowedCompactSummary.Verdict}" is "PASS"
    And "{allowedCompactSummary.ResultClass}" is "PASS"
    And "{allowedCompactSummary.DryRunAllowed}" is "true"
