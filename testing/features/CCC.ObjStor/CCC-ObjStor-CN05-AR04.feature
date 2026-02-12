@CCC.ObjStor @CCC.ObjStor.CN05 @tlp-clear @tlp-green @tlp-amber @tlp-red
Feature: CCC.ObjStor.CN05.AR04 - Retain Versions on Delete
  As a security administrator
  I want to ensure object versions are retained when objects are deleted
  So that deleted data can be recovered

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy
  Scenario: Object versions are retained after deletion
    # This is inherent to versioning being enabled - covered by CN05.AR01
    Then no-op required
