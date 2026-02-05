@CCC.VPC @tlp-amber @tlp-red @CCC.VPC.CN01.AR01
Feature: CCC.VPC.CN01.AR01
  As a security administrator
  I want to ensure default network resources are not present
  So that insecure default network configurations are avoided

  Background:
    Given a cloud api for "{Provider}" in "api"
    And I call "{api}" with "GetServiceAPI" with parameter "vpc"
    And I refer to "{result}" as "vpc"

  Scenario: Subscription must not contain a default VPC
    When I call "{vpc}" with "CountDefaultVpcs"
    Then "{result}" is "0"

