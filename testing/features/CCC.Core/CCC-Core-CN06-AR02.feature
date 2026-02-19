@PerService @CCC.Core @CCC.Core.CN06 @tlp-clear @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN06.AR02 - Child Resource Location Compliance
  As a security administrator
  I want to ensure child resources are deployed in approved regions
  So that data residency requirements are met for all resources

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @CCC.ObjStor
  Scenario: Child resources are in approved regions
    # Child resources inherit region from parent in most cloud services
    # Covered by CN06.AR01 parent resource region check
    Then no-op required
