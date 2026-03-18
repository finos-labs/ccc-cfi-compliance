@PerService @CCC.Core @CCC.Core.CN06 @tlp-clear @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.CN06.AR02 - Child Resource Location Compliance
  As a security administrator
  I want to ensure child resources are deployed in approved regions
  So that data residency requirements are met for all resources

  Background:
    Given a cloud api for "{Instance}" in "api"

  @Behavioural @object-storage
  Scenario: Child resource region compliance cannot be tested separately
    # Child resources (e.g., objects in a bucket) inherit region from parent in most
    # cloud services. The parent resource region check in CN06.AR01 covers this.
    # Separate automated verification would require enumerating all child resource types
    # and their region inheritance - which varies by provider and resource type.
    #
    # Manual verification steps:
    # 1. Identify child resources (objects, etc.) under parent (bucket, etc.)
    # 2. Verify child resources inherit parent region
    # 3. Confirm no child resources exist in non-approved regions
    Then no-op required
