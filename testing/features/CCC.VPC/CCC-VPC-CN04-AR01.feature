@CCC.VPC @tlp-amber @tlp-red @CCC.VPC.CN04.AR01
Feature: CCC.VPC.CN04.AR01
  As a security administrator
  I want VPC flow logs enabled for all traffic
  So that network activity is captured for monitoring and investigation

  Background:
    Given a cloud api for "{Provider}" in "api"
    And I call "{api}" with "GetServiceAPI" with parameter "vpc"
    And I refer to "{result}" as "vpc"

  Scenario: VPC flow logs must be active and capture all traffic
    When I call "{vpc}" with "SummarizeVpcFlowLogs" with parameter "{UID}"
    And I attach "{result}" to the test output as "cn04-flow-logs-summary.txt"
    And I call "{vpc}" with "ListVpcFlowLogs" with parameter "{UID}"
    And I refer to "{result}" as "flowLogs"
    And I attach "{flowLogs}" to the test output as "cn04-flow-logs.json"
    And I call "{vpc}" with "HasActiveAllTrafficFlowLogs" with parameter "{UID}"
    Then "{result}" is "true"
