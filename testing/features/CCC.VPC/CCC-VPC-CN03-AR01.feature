@tlp-amber @tlp-red @CCC.VPC.CN03.AR01
Feature: CCC.VPC.CN03.AR01 - Restrict VPC peering requests from non-allowlisted requesters
  As a security administrator
  I want peering requests from non-approved requester VPCs to be denied
  So that network connectivity is restricted to authorized boundaries

  Background:
    Given a cloud api for "{Provider}" in "api"
    And I call "{api}" with "GetServiceAPI" with parameter "vpc"
    And I refer to "{result}" as "vpcService"
    And I refer to "{CN03_RECEIVER_VPC_ID}" as "ReceiverVpcId"
    And "{ReceiverVpcId}" is not nil

  # Inputs:
  # - CN03_RECEIVER_VPC_ID: fixed receiver/target VPC ID for dry-run attempts
  # - CN03_ALLOWED_REQUESTER_VPC_ID_1..N: allowed requester VPC IDs for scenario coverage
  # - CN03_DISALLOWED_REQUESTER_VPC_ID_1..N: disallowed requester VPC IDs for scenario coverage
  # - CN03_NON_ALLOWLISTED_REQUESTER_VPC_ID: requester VPC ID outside explicit allow/disallow lists
  # - CN03_ALLOWED_REQUESTER_VPC_IDS: optional CSV allow-list used by requester classification
  # - CN03_PEER_OWNER_ID: optional (cross-account)
  # - CN03_PEER_TRIAL_MATRIX_FILE: optional JSON file for batch dry-run coverage
  #
  # Dry-run is used so no real peering connection is created.
  # Requester VPC is always "<RequesterVpcId>" and receiver VPC is always "{ReceiverVpcId}".

  @Policy @SANITY @ALLOWLIST @OPT_IN
  Scenario Outline: Allow-list classification: allowed requesters are allowed and disallowed requesters are not
    Given "<RequesterVpcId>" is not nil
    When I call "{vpcService}" with "EvaluatePeerAgainstAllowList" with parameter "<RequesterVpcId>"
    Then "{result.AllowedListDefined}" is true
    And "{result.Allowed}" is <ExpectedAllowed>

    Examples:
      | RequesterVpcId                       | ExpectedAllowed |
      | {CN03_ALLOWED_REQUESTER_VPC_ID_1}    | true            |
      | {CN03_ALLOWED_REQUESTER_VPC_ID_2}    | true            |
      | {CN03_DISALLOWED_REQUESTER_VPC_ID_1} | false           |
      | {CN03_DISALLOWED_REQUESTER_VPC_ID_2} | false           |

  @Destructive @MAIN @DEFAULT
  @CCC.VPC
  Scenario Outline: Enforcement proof (dry-run): disallowed requesters are denied against in-scope receiver VPC
    Given "<RequesterVpcId>" is not nil
    When I call "{vpcService}" with "AttemptVpcPeeringDryRun" with parameters "<RequesterVpcId>" and "{ReceiverVpcId}"
    Then "{result.DryRunAllowed}" is false
    And "{result.ExitCode}" should be greater than "0"

    Examples:
      | RequesterVpcId                       |
      | {CN03_DISALLOWED_REQUESTER_VPC_ID_1} |
      | {CN03_DISALLOWED_REQUESTER_VPC_ID_2} |

  @Destructive @MAIN @DEFAULT
  @CCC.VPC
  Scenario: Enforcement proof (dry-run): non-allowlisted requester is denied even when not explicitly listed as disallowed
    Given "{CN03_NON_ALLOWLISTED_REQUESTER_VPC_ID}" is not nil
    When I call "{vpcService}" with "EvaluatePeerAgainstAllowList" with parameter "{CN03_NON_ALLOWLISTED_REQUESTER_VPC_ID}"
    Then "{result.AllowedListDefined}" is true
    And "{result.Allowed}" is false
    When I call "{vpcService}" with "AttemptVpcPeeringDryRun" with parameters "{CN03_NON_ALLOWLISTED_REQUESTER_VPC_ID}" and "{ReceiverVpcId}"
    Then "{result.DryRunAllowed}" is false
    And "{result.ExitCode}" should be greater than "0"

  @Destructive @SANITY @OPT_IN
  # NOTE: no @CCC.VPC tag => opt-in only
  Scenario Outline: Enforcement sanity (dry-run): allowed requesters would be permitted against in-scope receiver VPC
    Given "<RequesterVpcId>" is not nil
    When I call "{vpcService}" with "AttemptVpcPeeringDryRun" with parameters "<RequesterVpcId>" and "{ReceiverVpcId}"
    Then "{result.DryRunAllowed}" is true

    Examples:
      | RequesterVpcId                    |
      | {CN03_ALLOWED_REQUESTER_VPC_ID_1} |
      | {CN03_ALLOWED_REQUESTER_VPC_ID_2} |

  @Destructive @SANITY @OPT_IN
  # NOTE: no @CCC.VPC tag => opt-in only
  Scenario: Batch trial matrix (dry-run): all file-listed requesters match expected outcomes
    Given "{CN03_PEER_TRIAL_MATRIX_FILE}" is not nil
    When I call "{vpcService}" with "RunVpcPeeringDryRunTrialsFromFile" with parameter "{CN03_PEER_TRIAL_MATRIX_FILE}"
    Then "{result.TotalTrials}" should be greater than "0"
    And "{result.UnexpectedCount}" is "0"
