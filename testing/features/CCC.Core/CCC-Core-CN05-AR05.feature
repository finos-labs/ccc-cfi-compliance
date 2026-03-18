@PerService @CCC.Core @CCC.Core.CN05 @tlp-red
Feature: CCC.Core.CN05.AR05 - Hide Service Existence from External Requests
  As a security administrator
  I want to ensure external requests receive no indication that service exists
  So that service discovery attacks are prevented

  Background:
    Given a cloud api for "{Instance}" in "api"

  @Policy @NotTested @object-storage
  Scenario: External requests do not reveal service existence
    # This is unsupported by the policy engine - all cloud accounts will
    # return a 403 Forbidden error if an attacker correctly guesses the 
    # account / bucket / blob names.
    Then no-op required
