@tlp-amber @tlp-red @CCC.VPC.CN03.AR01
Feature: CCC.VPC.CN03.AR01 - Restrict VPC peering to explicitly allowed destinations
  As a security administrator
  I want peering requests to non-approved VPCs to be denied
  So that network connectivity is restricted to authorized boundaries

  Background:
    Given a cloud api for "{Provider}" in "api"
    And I call "{api}" with "GetServiceAPI" with parameter "vpc"
    And I refer to "{result}" as "vpcService"

  # Dry-run is used so no real peering connection is created.

  @Destructive @MAIN @OPT_IN @PENDING_API
  # NOTE: explicit requester+peer method is planned API work.
  Scenario: Main check (dry-run): disallowed peering request uses requester and peer VPC IDs
    Given I refer to "{UID}" as "RequesterVpcId"
    And I refer to "{CN03_PEER_VPC_ID}" as "PeerVpcId"
    When I call "{vpcService}" with "AttemptVpcPeeringDryRun" with parameters "{RequesterVpcId}" and "{PeerVpcId}"
    Then "{result.ExitCode}" should be greater than "0"
    And "{result.DryRunAllowed}" is false

  @Destructive @SANITY @OPT_IN @PENDING_API
  # NOTE: no @CCC.VPC tag => opt-in only
  @CN03.ALLOWED
  Scenario: Sanity check (dry-run): allowed peering request uses requester and peer VPC IDs
    Given I refer to "{UID}" as "RequesterVpcId"
    And I refer to "{CN03_ALLOWED_PEER_VPC_ID}" as "PeerVpcId"
    When I call "{vpcService}" with "AttemptVpcPeeringDryRun" with parameters "{RequesterVpcId}" and "{PeerVpcId}"
    Then "{result.DryRunAllowed}" is true
