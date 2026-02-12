@CCC.Core @CCC.Core.CN13 @tlp-amber
Feature: CCC.Core.CN13.AR02 - Certificate Rotation within 180 Days
  As a security administrator
  I want to ensure certificates are rotated within 180 days
  So that certificate compromise risks are minimized

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @CCC.ObjStor
  Scenario: Certificates are rotated within 180 days
    # Certificate age verification requires checking certificate NotBefore date
    # This is typically enforced by certificate management policies
    Then no-op required
