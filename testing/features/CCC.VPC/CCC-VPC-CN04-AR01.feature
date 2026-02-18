@tlp-amber @tlp-red @CCC.VPC.CN04.AR01
Feature: CCC.VPC.CN04.AR01 - Flow logs must capture all VPC traffic
  As a security administrator
  I want VPC traffic to be captured and logged
  So that audit and investigation requirements are met

  Background:
    Given a cloud api for "{Provider}" in "api"
    And I call "{api}" with "GetServiceAPI" with parameter "vpc"
    And I refer to "{result}" as "vpcService"

  # Policy check: flow logs are configured as ACTIVE with TrafficType=ALL.
  @Policy @MAIN @DEFAULT
  @CCC.VPC
  Scenario: Main check (config): flow logs are active and capture all traffic
    Given I refer to "{UID}" as "TargetVpcId"
    When I call "{vpcService}" with "EvaluateVpcFlowLogsControl" with parameter "{TargetVpcId}"
    Then "{result.FlowLogCount}" should be greater than "0"
    And "{result.NonCompliantCount}" is "0"

  # Behavior check: generate traffic and observe new flow log records.
  @Behavior @OPT_IN @PENDING_API
  # NOTE: no @CCC.VPC tag => opt-in only (may generate traffic and incur cost)
  Scenario: Behavioral check (active): traffic produces flow log records
    Given I refer to "{UID}" as "TargetVpcId"
    When I call "{vpcService}" with "PrepareFlowLogDeliveryObservation" with parameter "{TargetVpcId}"
    And I call "{vpcService}" with "GenerateTestTraffic" with parameter "{TargetVpcId}"
    And I call "{vpcService}" with "ObserveRecentFlowLogDelivery" with parameter "{TargetVpcId}"
    Then "{result.RecordsObserved}" is true
