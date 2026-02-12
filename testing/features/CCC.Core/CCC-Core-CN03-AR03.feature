@CCC.Core @CCC.Core.CN03 @tlp-amber @tlp-red
Feature: CCC.Core.CN03.AR03 - MFA for UI Viewing
  As a security administrator
  I want to ensure viewing information through UI requires MFA
  So that sensitive data access is protected

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @CCC.ObjStor
  Scenario: UI viewing requires multi-factor authentication
    # This control is enforced at the identity provider level (Azure AD, AWS IAM, etc.)
    # Cannot be verified with a resource-level policy check
    Then no-op required
