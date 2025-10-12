@PerPort @tls @tlp-amber @tlp-red
Feature: CCC.Core.C01.TR08
  As a security administrator
  I want to ensure mutual TLS is implemented for all TLS connections
  So that both client and server are authenticated to prevent unauthorized access

  @tls
  Scenario: Verify mTLS requires client certificate authentication
    Mutual TLS (mTLS) requires both server and client certificates for authentication.
    This test verifies that the server is configured to require client certificates,
    ensuring that only authenticated clients can establish connections.

    Given "report" contains details of SSL Support type "server-defaults" for "{hostName}" on port "{portNumber}"
    Then "{report.scanResult[0].serverDefaults}" is an slice of objects with at least the following contents
      | id         | finding  |
      | clientAuth | required |
