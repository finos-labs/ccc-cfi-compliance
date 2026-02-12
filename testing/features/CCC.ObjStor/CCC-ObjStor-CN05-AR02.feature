@CCC.ObjStor @CCC.ObjStor.CN05 @tlp-clear @tlp-green @tlp-amber @tlp-red
Feature: CCC.ObjStor.CN05.AR02 - New Version ID on Modification
  As a security administrator
  I want to ensure modified objects receive new version identifiers
  So that changes are tracked

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy
  Scenario: Modified objects receive new version identifiers
    # This is inherent to versioning being enabled - covered by CN05.AR01
    Then no-op required
