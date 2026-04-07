@vpc @tlp-red @CCC.VPC.CN02 @CCC.VPC.CN02.AR01
Feature: CCC.VPC.CN02.AR01 - No external IP by default in public subnets
  As a security administrator
  I want to ensure resources created in public subnets are not assigned an external IP address by default
  So that public exposure is minimized

  Background:
    Given a cloud api for "{Instance}" in "api"
    And I call "{api}" with "GetServiceAPI" using argument "vpc"
    And I refer to "{result}" as "vpcService"

  # Public subnet: has a route to an Internet Gateway (IGW)
  # Default external IP assignment: subnet setting MapPublicIpOnLaunch = true

  @Policy @MAIN @CCC.VPC @DEFAULT
  Scenario: Main check (config): public subnets do not auto-assign external IPs
    Given I refer to "{UID}" as "TargetVpcId"
    When I call "{vpcService}" with "EvaluatePublicSubnetDefaultIPControl" using argument "{TargetVpcId}"
    Then "{result.ViolatingSubnetCount}" is "0"
    And "{result.Reason}" contains "disable default public IP"

  @Policy @NEGATIVE @OPT_IN
  # Redeploy with map_public_ip_on_launch=true in terraform.tfvars before running this scenario.
  # Run with: --tags '@NEGATIVE'
  Scenario: Negative check: public subnets auto-assign external IPs (failure simulation)
    Given I refer to "{UID}" as "TargetVpcId"
    When I call "{vpcService}" with "EvaluatePublicSubnetDefaultIPControl" using argument "{TargetVpcId}"
    Then "{result.ViolatingSubnetCount}" should be greater than "0"
    And "{result.Reason}" contains "MapPublicIpOnLaunch=true"

  @Behavioural @MAIN @CCC.VPC 
  # Requires CN_TEST_AMI_ID set in compliance-testing.env (region-specific AMI ID).
  # Leave CN_TEST_AMI_ID blank to skip. Launches and deletes a short-lived EC2 instance.
  # Run with: --tags '@Behavioural'
  Scenario: Behavioural check (active): resource launched in public subnet is not assigned an external IP
    Given I refer to "{UID}" as "TargetVpcId"
    When I call "{vpcService}" with "SelectPublicSubnetForTest" using argument "{TargetVpcId}"
    And I refer to "{result.SubnetId}" as "TestSubnetId"
    And I call "{vpcService}" with "CreateTestResourceInSubnet" using argument "{TestSubnetId}"
    And I refer to "{result.ResourceId}" as "TestResourceId"
    And I call "{vpcService}" with "GetResourceExternalIpAssignment" using argument "{TestResourceId}"
    And I refer to "{result.HasExternalIp}" as "HasExternalIp"
    # And we wait for a period of "20000" ms # uncomment to allow visible confirmation for checking instance live for 20 seconds
    Then "{HasExternalIp}" is false
    When I call "{vpcService}" with "DeleteTestResource" using argument "{TestResourceId}"
    Then "{result.Deleted}" is true
