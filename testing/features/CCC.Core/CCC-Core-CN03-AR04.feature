@CCC.Core @CCC.Core.CN03 @tlp-amber @tlp-red
Feature: CCC.Core.CN03.AR04 - API Authentication for Viewing
  As a security administrator
  I want to ensure API viewing requires credentials from within trust perimeter
  So that unauthorized data viewing is prevented

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @CCC.ObjStor
  Scenario: API viewing requires credential and trust perimeter origin
    # This control requires behavioral testing - verifying API auth mechanisms
    # Cannot be verified with a simple policy check
    Then no-op required
