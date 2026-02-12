@CCC.Core @CCC.Core.CN13 @tlp-clear @tlp-green @tlp-amber @tlp-red @PerPort
Feature: CCC.Core.CN13.AR01 - Valid Unexpired Certificates
  As a security administrator
  I want to ensure only valid, unexpired certificates are used
  So that secure communications are maintained

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Behavioral @CCC.ObjStor
  Scenario: Certificates are valid and unexpired
    Given "report" contains details of SSL Support type "server-defaults" for "{hostName}" on port "{portNumber}"
    Then "{report}" is a slice of objects with at least the following contents
      | id                    | finding |
      | cert_expirationStatus | ok      |
      | cert_chain_of_trust   | passed. |
