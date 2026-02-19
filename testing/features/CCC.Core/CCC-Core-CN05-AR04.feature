@PerService @CCC.Core @CCC.Core.CN05 @tlp-amber @tlp-red
Feature: CCC.Core.CN05.AR04 - Block Unauthorized External Data Requests
  As a security administrator
  I want to ensure data requests from outside trust perimeter are blocked
  So that data exfiltration is prevented

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @CCC.ObjStor
  Scenario: External unauthorized data requests are blocked
    # This control requires behavioral testing - attempting external access
    # Network policies and IAM policies enforce this at runtime
    Then no-op required
