@tls @tlp-amber @tlp-red @CCC.Core @CCC.Core.CN01
Feature: CCC.Core.CN01.AR08
  As a security administrator
  I want to ensure mutual TLS is implemented for all TLS connections
  So that both client and server are authenticated to prevent unauthorized access

  @Behavioural @PerPort @tls
  Scenario: Verify mTLS requires client certificate authentication
    Mutual TLS (mTLS) requires both server and client certificates for authentication.
    This test verifies that the server is configured to require client certificates,
    ensuring that only authenticated clients can establish connections.

    Given "report" contains details of SSL Support type "server-defaults" for "{hostName}" on port "{portNumber}"
    Then "{report}" is a slice of objects with at least the following contents
      | id         | finding  |
      | clientAuth | required |

  @Policy @PerService
  Scenario: Load balancer enforces mutual TLS
    When I attempt policy check "load-balancer-mtls" for control "CCC.Core.CN01" assessment requirement "AR08" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true

  @Policy @PerService
  Scenario: Load balancer has valid trust store
    When I attempt policy check "load-balancer-trust-store" for control "CCC.Core.CN01" assessment requirement "AR08" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true
