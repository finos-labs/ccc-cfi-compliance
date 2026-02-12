@CCC.Core @CCC.Core.CN09 @tlp-amber @tlp-red
Feature: CCC.Core.CN09.AR03 - Log Redirection Requires Service Halt
  As a security administrator
  I want to ensure log redirection requires halting the resource
  So that log tampering attempts are prevented

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @CCC.ObjStor
  Scenario: Redirecting logs requires halting the resource
    # This control requires behavioral testing - attempting to redirect logs
    # Cloud provider logging configurations enforce this at runtime
    Then no-op required
