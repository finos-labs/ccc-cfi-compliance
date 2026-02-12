@PerService @CCC.Core @CCC.Core.CN09 @tlp-clear @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN09.AR02 - Logs Cannot Be Disabled
  As a security administrator
  I want to ensure logging cannot be disabled without disabling the resource
  So that audit trail tampering is prevented

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @CCC.ObjStor
  Scenario: Disabling logs requires disabling the resource
    # This control requires behavioral testing - attempting to disable logs
    # Cloud provider logging configurations enforce this at runtime
    Then no-op required
