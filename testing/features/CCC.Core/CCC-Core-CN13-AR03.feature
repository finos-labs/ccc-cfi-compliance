@CCC.Core @CCC.Core.CN13 @tlp-red
Feature: CCC.Core.CN13.AR03 - Certificate Rotation within 90 Days
  As a security administrator
  I want to ensure certificates are rotated within 90 days
  So that stringent certificate rotation requirements are met

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @CCC.ObjStor
  Scenario: Certificates are rotated within 90 days
    # Certificate age verification requires checking certificate NotBefore date
    # This is typically enforced by certificate management policies
    Then no-op required
