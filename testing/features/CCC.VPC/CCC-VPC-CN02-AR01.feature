@tlp-red @CCC.VPC.CN02.AR01
Feature: CCC.VPC.CN02.AR01 - No external IP by default in public subnets
  As a security administrator
  I want to ensure resources created in public subnets are not assigned an external IP address by default
  So that public exposure is minimized

  Background:
    Given a cloud api for "{Provider}" in "api"
    And I call "{api}" with "GetServiceAPI" with parameter "vpc"
    And I refer to "{result}" as "vpcService"

  # Public subnet: has a route to an Internet Gateway (IGW)
  # Default external IP assignment: subnet setting MapPublicIpOnLaunch = true

  @Policy @MAIN @DEFAULT
  @CCC.VPC
  Scenario: Main check (config): public subnets do not auto-assign external IPs
    Given I refer to "{UID}" as "TargetVpcId"
    When I call "{vpcService}" with "EvaluatePublicSubnetDefaultIPControl" with parameter "{TargetVpcId}"
    Then "{result.ViolatingSubnetCount}" is "0"
    And "{result.Reason}" contains "public subnet"

  @Behavior @OPT_IN @PENDING_API
  # NOTE: no @CCC.VPC tag => opt-in only (creates and deletes a test resource)
  Scenario: Behavioral check (active): creating a resource in a public subnet does not assign an external IP by default
    Given I refer to "{UID}" as "TargetVpcId"
    When I call "{vpcService}" with "SelectPublicSubnetForTest" with parameter "{TargetVpcId}"
    And I refer to "{result.SubnetId}" as "TestSubnetId"
    And I call "{vpcService}" with "CreateTestResourceInSubnet" with parameter "{TestSubnetId}"
    And I refer to "{result.ResourceId}" as "TestResourceId"
    And I call "{vpcService}" with "GetResourceExternalIpAssignment" with parameter "{TestResourceId}"
    And I refer to "{result.HasExternalIp}" as "HasExternalIp"
    # And we wait for a period of "20000" ms # uncomment to allow visisble confirmation for checking instance live for 20 seconds
    And I call "{vpcService}" with "DeleteTestResource" with parameter "{TestResourceId}"
    Then "{result.Deleted}" is true
    And "{HasExternalIp}" is false
