@PerPort @tlp-clear @tlp-green @tlp-amber @tlp-red
Feature: CCC.Core.C01.TR07
  As a security administrator
  I want to ensure that only the IANA-assigned protocol runs on each port
  So that services follow standard port assignments and avoid misconfigurations

  @http @plaintext
  Scenario: Verify HTTP uses IANA-assigned port 80
    HTTP must use port 80 as assigned by IANA.
    Running HTTP on non-standard ports violates IANA assignments.

    Then "{portNumber}" is "80"

  @http @tls
  Scenario: Verify HTTPS uses IANA-assigned port 443
    HTTPS must use port 443 as assigned by IANA.
    This is the standard port for encrypted web traffic.

    Then "{portNumber}" is "443"

  @ssh
  Scenario: Verify SSH uses IANA-assigned port 22
    SSH must use port 22 as assigned by IANA.
    Running SSH on non-standard ports or other services on port 22 violates IANA assignments.

    Then "{portNumber}" is "22"

  @smtp @plaintext
  Scenario: Verify SMTP uses IANA-assigned port 25
    SMTP must use port 25 as assigned by IANA.
    This is the standard port for mail transfer between servers.

    Then "{portNumber}" is "25"

  @smtp @tls
  Scenario: Verify SMTPS uses IANA-assigned port 465 or 587
    SMTPS can use port 465 (implicit TLS) or 587 (STARTTLS) as assigned by IANA.

    Then "{portNumber}" is "465"

  @dns
  Scenario: Verify DNS uses IANA-assigned port 53
    DNS must use port 53 as assigned by IANA.
    Both TCP and UDP port 53 are reserved for domain name resolution.

    Then "{portNumber}" is "53"

  @ftp @plaintext
  Scenario: Verify FTP uses IANA-assigned port 21
    FTP must use port 21 as assigned by IANA.
    If FTP is disabled for security, this port should not be exposed.

    Then "{portNumber}" is "21"

  @ldap @plaintext
  Scenario: Verify LDAP uses IANA-assigned port 389
    LDAP must use port 389 as assigned by IANA.
    This is the standard port for directory services.

    Then "{portNumber}" is "389"

  @ldap @tls
  Scenario: Verify LDAPS uses IANA-assigned port 636
    LDAPS must use port 636 as assigned by IANA.
    This is the secure LDAP port with implicit TLS.

    Then "{portNumber}" is "636"

  @mysql
  Scenario: Verify MySQL uses IANA-assigned port 3306
    MySQL must use port 3306 as assigned by IANA.
    Running other services on database ports can cause application failures.

    Then "{portNumber}" is "3306"

  @postgres
  Scenario: Verify PostgreSQL uses IANA-assigned port 5432
    PostgreSQL must use port 5432 as assigned by IANA.
    This is the standard PostgreSQL database port.

    Then "{portNumber}" is "5432"
