@tlp-amber @tlp-red @CCC.VPC.CN01.AR01
Feature: CCC.VPC.CN01.AR01 - Subscription must not contain default network resources
  As a security administrator
  I want to ensure default network resources are not present in the subscription
  So that insecure default configurations are avoided and custom network policies are enforced

  Background:
    Given a cloud api for "{Provider}" in "api"
    And I call "{api}" with "GetServiceAPI" with parameter "vpc"
    And I refer to "{result}" as "vpcService"

  @Policy @DEFAULT @CCC.VPC
  Scenario: Main check: no default VPC exists
    When I call "{vpcService}" with "CountDefaultVpcs"
    Then "{result}" is "0"

  @Policy @NEGATIVE @OPT_IN
  # No @CCC.VPC tag => excluded from default VPC runs
  Scenario: Negative check: default VPC exists
    When I call "{vpcService}" with "CountDefaultVpcs"
    Then "{result}" should be greater than "0"
