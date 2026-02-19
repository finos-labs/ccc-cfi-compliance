@PerService @CCC.Core @CCC.Core.CN07 @tlp-clear @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN07.AR02 - Log Enumeration Activities
  As a security administrator
  I want to ensure enumeration activities are logged
  So that reconnaissance attempts can be investigated

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @CCC.ObjStor
  Scenario: Enumeration activities are logged
    # This control requires runtime logging verification
    # Covered by CN09.AR01 access logging configuration
    Then no-op required
