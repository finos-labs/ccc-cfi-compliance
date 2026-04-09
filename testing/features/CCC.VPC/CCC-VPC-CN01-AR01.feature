@vpc @tlp-amber @tlp-red @CCC.VPC.CN01 @CCC.VPC.CN01.AR01
Feature: CCC.VPC.CN01.AR01 - Subscription must not contain default network resources
  As a security administrator
  I want to ensure default network resources are not present in the subscription
  So that insecure default configurations are avoided and custom network policies are enforced

  Background:
    Given a cloud api for "{Instance}" in "api"
    And I call "{api}" with "GetServiceAPI" using argument "vpc"
    And I refer to "{result}" as "vpcService"

  @Policy @MAIN @CCC.VPC @DEFAULT
  Scenario: Main check: no default VPC exists
    When I call "{vpcService}" with "CountDefaultVpcs"
    Then "{result}" is "0"

  # @Policy @NEGATIVE @OPT_IN
  # TODO: negative check pending — purpose is to validate check logic correctness for false negatives, not VPC state
  # Scenario: Negative check: default VPC exists
  #   When I call "{vpcService}" with "CountDefaultVpcs"
  #   Then "{result}" should be greater than "0"
