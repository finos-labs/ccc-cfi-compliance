@CCC.VPC @tlp-amber @tlp-red @CCC.VPC.CN02.AR01
Feature: CCC.VPC.CN02.AR01
  As a security administrator
  I want to ensure resources in public subnets are not assigned external IPs by default
  So that public exposure is minimized

  Background:
    Given a cloud api for "{Provider}" in "api"
    And I call "{api}" with "GetServiceAPI" with parameter "vpc"
    And I refer to "{result}" as "vpc"

  Scenario: Public subnets must not map public IP on launch
    When I call "{vpc}" with "SummarizePublicSubnets" with parameter "{UID}"
    And I attach "{result}" to the test output as "cn02-public-subnet-summary.txt"
    And I call "{vpc}" with "ListPublicSubnets" with parameter "{UID}"
    And I refer to "{result}" as "publicSubnets"
    And I attach "{publicSubnets}" to the test output as "cn02-public-subnets.json"
    Then "{publicSubnets}" is a slice of objects which doesn't contain any of
      | MapPublicIpOnLaunch |
      | true               |
