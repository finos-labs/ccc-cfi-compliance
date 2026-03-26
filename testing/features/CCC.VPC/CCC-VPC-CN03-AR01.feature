@vpc @tlp-amber @tlp-red @CCC.VPC.CN03 @CCC.VPC.CN03.AR01
Feature: CCC.VPC.CN03.AR01 - Restrict VPC peering requests from non-allowlisted requesters
  As a security administrator
  I want peering requests from non-approved requester VPCs to be denied
  So that network connectivity is restricted to authorized boundaries

  Background:
    Given a cloud api for "{Instance}" in "api"
    And I call "{api}" with "GetServiceAPI" using argument "vpc"
    And I refer to "{result}" as "vpcService"
    And I load environment variable "CN03_RECEIVER_VPC_ID" as "ReceiverVpcId"
    And I load environment variable "CN03_NON_ALLOWLISTED_REQUESTER_VPC_ID" as "NonAllowlistedRequesterVpcId"
    And I load environment variable "CN03_PEER_TRIAL_MATRIX_FILE" as "PeerTrialMatrixFile"
    And "{ReceiverVpcId}" is not nil

  # Inputs:
  # - CN03_RECEIVER_VPC_ID: fixed receiver/target VPC ID for dry-run attempts
  # - CN03_ALLOWED_REQUESTER_VPC_ID_1..N: allowed requester VPC IDs for scenario coverage
  # - CN03_DISALLOWED_REQUESTER_VPC_ID_1..N: disallowed requester VPC IDs for scenario coverage
  # - CN03_NON_ALLOWLISTED_REQUESTER_VPC_ID: requester VPC ID outside explicit allow/disallow lists
  # - CN03_ALLOWED_REQUESTER_VPC_IDS: optional CSV allow-list used by requester classification
  # - CN03_DISALLOWED_REQUESTER_VPC_IDS: optional CSV disallow-list (mirrors allowed pattern)
  # - CN03_PEER_OWNER_ID: optional (cross-account)
  # - CN03_PEER_TRIAL_MATRIX_FILE: optional JSON file for batch dry-run coverage
  #
  # Dry-run is used so no real peering connection is created.

  @Policy @SANITY @ALLOWLIST @OPT_IN
  # Evaluates every VPC in the allow-list from all sources (terraform fixtures,
  # env vars, environment.yaml guardrail entries) and asserts all are correctly
  # classified.
  Scenario: Allow-list classification: all configured allowed VPCs are correctly classified
    When I call "{vpcService}" with "ValidateAllowListClassification"
    And I attach "{result.Results}" to the test output as "Allow-list Classification"
    Then "{result.AllowListDefined}" is true
    And "{result.AllowedCount}" should be greater than "0"
    And "{result.AllClassifiedCorrectly}" is true
    And "{result.MisclassifiedCount}" is "0"

  @Destructive @MAIN @DEFAULT @CCC.VPC
  # Dry-runs every VPC in the disallow-list (terraform fixtures, env vars,
  # environment.yaml) against the in-scope receiver VPC. A guardrail mismatch
  # means a disallowed VPC was permitted — that is a compliance failure.
  Scenario: Enforcement proof (dry-run): all disallowed requesters are denied against in-scope receiver VPC
    When I call "{vpcService}" with "ValidateDisallowListEnforcement" using argument "{ReceiverVpcId}"
    And I attach "{result.Summary}" to the test output as "Disallow-list Enforcement Summary"
    And I attach "{result.Results}" to the test output as "Disallow-list Enforcement"
    Then "{result.ListDefined}" is true
    And "{result.TestedCount}" should be greater than "0"
    And "{result.AllCorrect}" is true
    And "{result.ViolationCount}" is "0"

  @Destructive @MAIN @DEFAULT @CCC.VPC
  Scenario: Enforcement proof (dry-run): non-allowlisted requester is denied even when not explicitly listed as disallowed
    Given "{NonAllowlistedRequesterVpcId}" is not nil
    When I call "{vpcService}" with "EvaluatePeerAgainstAllowList" using argument "{NonAllowlistedRequesterVpcId}"
    Then "{result.AllowedListDefined}" is true
    And "{result.Allowed}" is false
    When I call "{vpcService}" with "AttemptVpcPeeringDryRun" using arguments "{NonAllowlistedRequesterVpcId}" and "{ReceiverVpcId}"
    Then "{result.DryRunAllowed}" is false
    And "{result.AllowListDefined}" is true
    And "{result.RequesterInAllowList}" is false
    And "{result.GuardrailExpectation}" is "deny"
    And "{result.GuardrailMismatch}" is false
    And "{result.ExitCode}" should be greater than "0"
    And "{result.Reason}" contains "guardrail aligned"
    And "{result.ConflictType}" is ""

  @Destructive @SANITY @OPT_IN
  # Dry-runs every VPC in the allow-list against the in-scope receiver VPC.
  # A guardrail mismatch means a legitimately allowed VPC was denied — that
  # indicates misconfigured guardrail policy.
  Scenario: Enforcement sanity (dry-run): all allowed requesters are permitted against in-scope receiver VPC
    When I call "{vpcService}" with "ValidateAllowListEnforcement" using argument "{ReceiverVpcId}"
    And I attach "{result.Results}" to the test output as "Allow-list Enforcement"
    Then "{result.ListDefined}" is true
    And "{result.TestedCount}" should be greater than "0"
    And "{result.AllCorrect}" is true
    And "{result.ViolationCount}" is "0"

  @Destructive @SANITY @OPT_IN
  Scenario: Batch trial matrix (dry-run): all file-listed requesters match expected outcomes
    Given "{PeerTrialMatrixFile}" is not nil
    When I call "{vpcService}" with "RunVpcPeeringDryRunTrialsFromFile" using argument "{PeerTrialMatrixFile}"
    Then "{result.TotalTrials}" should be greater than "0"
    And "{result.UnexpectedCount}" is "0"
    And "{result.Compliant}" is true
