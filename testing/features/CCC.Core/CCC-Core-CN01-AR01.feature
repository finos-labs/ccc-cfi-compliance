@CCC.Core @tlp-green @tlp-amber @tlp-red @CCC.Core.CN01 @tls
Feature: CCC.Core.CN01.AR01
  As a security administrator
  I want to ensure all non-SSH network traffic uses TLS 1.3 or higher
  So that data integrity and confidentiality are protected during transmission

  Background:
    Given a cloud api for "{Provider}" in "api"

  @Behavioural @PerPort
  Scenario: Service accepts TLS 1.3 encrypted traffic
    Given an openssl s_client request using "tls1_3" to "{portNumber}" on "{hostName}" protocol "{protocol}"
    And I refer to "{result}" as "connection"
    And "{connection}" state is open
    And "{connection.State}" is "open"
    And I close connection "{connection}"
    Then "{connection}" state is closed

  @Behavioural @PerPort
  Scenario: Service rejects TLS 1.2 traffic
    Given an openssl s_client request using "tls1_2" to "{portNumber}" on "{hostName}" protocol "{protocol}"
    And I refer to "{result}" as "connection"
    And we wait for a period of "40" ms
    Then "{connection.State}" is "closed"

  @Behavioural @PerPort
  Scenario: Service rejects TLS 1.1 traffic
    Given an openssl s_client request using "tls1_1" to "{portNumber}" on "{hostName}" protocol "{protocol}"
    And I refer to "{result}" as "connection"
    And we wait for a period of "40" ms
    Then "{connection.State}" is "closed"

  @Behavioural @PerPort
  Scenario: Service rejects TLS 1.0 traffic
    Given an openssl s_client request using "tls1" to "{portNumber}" on "{hostName}" protocol "{protocol}"
    And I refer to "{result}" as "connection"
    And we wait for a period of "40" ms
    Then "{connection.State}" is "closed"

  @Behavioural @PerPort
  Scenario: Verify SSL/TLS protocol support
    Given "report" contains details of SSL Support type "protocols" for "{hostName}" on port "{portNumber}"
    Then "{report}" is a slice of objects which doesn't contain any of
      | id     | finding |
      | SSLv2  | offered |
      | SSLv3  | offered |
      | TLS1   | offered |
      | TLS1_1 | offered |
      | TLS1_2 | offered |
    And "{report}" is a slice of objects with at least the following contents
      | id     | finding            |
      | TLS1_3 | offered with final |

  @Behavioural @PerPort
  Scenario: Verify no known SSL/TLS vulnerabilities
    Given "report" contains details of SSL Support type "vulnerable" for "{hostName}" on port "{portNumber}"
    Then "{report}" is a slice of objects with at least the following contents
      | id            | severity |
      | heartbleed    | OK       |
      | CCS           | OK       |
      | ticketbleed   | OK       |
      | ROBOT         | OK       |
      | secure_renego | OK       |

  @Behavioural @PerPort
  Scenario: Verify TLS 1.3 only certificate validity
    Given "report" contains details of SSL Support type "server-defaults" for "{hostName}" on port "{portNumber}"
    Then "{report}" is a slice of objects with at least the following contents
      | id                    | severity |
      | cert_expirationStatus | OK       |
      | cert_chain_of_trust   | OK       |

  @Policy @PerService @CCC.ObjStor
  Scenario: Storage account enforces minimum TLS version
    When I attempt policy check "object-storage-tls-policy" for control "CCC.Core.CN01" assessment requirement "AR01" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true

  @Policy @PerService @CCC.LoadBalancer
  Scenario: Load balancer enforces minimum TLS version
    When I attempt policy check "load-balancer-tls-policy" for control "CCC.Core.CN01" assessment requirement "AR01" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
    Then "{result}" is true
