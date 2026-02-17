@PerService @CCC.Core @CCC.Core.CN03 @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN03.AR01 - Multi-Factor Authentication for Destructive Operations
  As a security administrator
  I want to ensure destructive operations require multi-factor authentication
  So that accidental or malicious deletions are prevented

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Policy @CCC.ObjStor
  Scenario: Object storage delete protection compliance
    When I attempt policy check "object-storage-delete-protection" for control "CCC.Core.CN03" assessment requirement "AR01" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true

  @Behavioural @CCC.ObjStor
  Scenario: MFA requirement for destructive operations cannot be tested automatically
    # Multi-factor authentication (MFA) for destructive operations requires human interaction
    # to complete the second factor challenge (e.g., TOTP code, push notification, hardware key).
    # Automated testing cannot simulate this interactive flow without compromising security.
    # 
    # Manual verification steps:
    # 1. Attempt to delete a protected resource (bucket, object with retention, etc.)
    # 2. Verify that MFA prompt is triggered before deletion proceeds
    # 3. Confirm deletion only succeeds after valid MFA response
    #
    # Policy check "object-storage-delete-protection" above validates that MFA Delete
    # is configured at the infrastructure level.
    Then no-op required
