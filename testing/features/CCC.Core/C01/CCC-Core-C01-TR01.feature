@PerPort @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.C01.TR01
  As a security administrator
  I want to ensure all non-SSH network traffic uses TLS 1.3 or higher
  So that data integrity and confidentiality are protected during transmission
  # Scenario: Service accepts TLS 1.3 encrypted traffic
  #   Given an openssl s_client request using "tls1_3" to "{portNumber}" on "{hostName}" protocol "{protocol}"
  #   And I refer to "{result}" as "connection"
  #   And "{connection}" state is open
  #   And "{connection.State}" is "open"
  #   And I close connection "{connection}"
  #   Then "{connection}" state is closed
  # Scenario: Service rejects TLS 1.2 traffic
  #   Given an openssl s_client request using "tls1_2" to "{portNumber}" on "{hostName}" protocol "{protocol}"
  #   And I refer to "{result}" as "connection"
  #   And we wait for a period of "40" ms
  #   Then "{connection.State}" is "closed"
  # Scenario: Service rejects TLS 1.1 traffic
  #   Given an openssl s_client request using "tls1_1" to "{portNumber}" on "{hostName}" protocol "{protocol}"
  #   And I refer to "{result}" as "connection"
  #   And we wait for a period of "40" ms
  #   Then "{connection.State}" is "closed"
  # Scenario: Service rejects TLS 1.0 traffic
  #   Given an openssl s_client request using "tls1" to "{portNumber}" on "{hostName}" protocol "{protocol}"
  #   And I refer to "{result}" as "connection"
  #   And we wait for a period of "40" ms
  #   Then "{connection.State}" is "closed"

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
  # Scenario: Verify no known SSL/TLS vulnerabilities
  #   Given "report" contains details of SSL Support type "vulnerable" for "{hostName}" on port "{portNumber}"
  #   Then "{report}" is a slice of objects with at least the following contents
  #     | id            | finding                                |
  #     | heartbleed    | not vulnerable, no heartbeat extension |
  #     | CCS           | not vulnerable                         |
  #     | ticketbleed   | not vulnerable                         |
  #     | ROBOT         | not vulnerable                         |
  #     | secure_renego | supported                              |
  # Scenario: Verify TLS 1.3 only certificate validity
  #   Given "report" contains details of SSL Support type "server-defaults" for "{hostName}" on port "{portNumber}"
  #   Then "{report}" is a slice of objects with at least the following contents
  #     | id                    | finding |
  #     | cert_expirationStatus | ok      |
  #     | cert_chain_of_trust   | passed. |
