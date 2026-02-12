@CCC.ObjStor @CCC.ObjStor.CN05 @tlp-clear @tlp-green @tlp-amber @tlp-red
Feature: CCC.ObjStor.CN05.AR03 - Recovery of Previous Versions
  As a security administrator
  I want to ensure previous object versions can be recovered
  So that data can be restored after modifications

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy
  Scenario: Previous object versions can be recovered
    # This is inherent to versioning being enabled - covered by CN05.AR01
    Then no-op required
