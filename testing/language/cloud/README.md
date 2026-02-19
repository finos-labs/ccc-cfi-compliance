# Cloud Cucumber Steps for Go

This package provides cloud-specific Cucumber/Godog step definitions for CCC (Common Cloud Controls) compliance testing. It extends the generic steps with cloud provider integrations, SSL/TLS testing capabilities, and protocol-specific connection handling.

## Prerequisites

### testssl.sh

SSL/TLS analysis steps require [testssl.sh](https://github.com/drwetter/testssl.sh) to be installed:

**macOS:**
```bash
brew install testssl
```

**Linux (Ubuntu/Debian):**
```bash
apt-get install testssl.sh
# or install from source:
git clone --depth 1 https://github.com/drwetter/testssl.sh.git /opt/testssl
sudo ln -s /opt/testssl/testssl.sh /usr/local/bin/testssl.sh
```

The code will automatically use the system-installed `testssl.sh` if available, or fall back to a local copy in this directory.

## Features

- Cloud provider API initialization (AWS, Azure, GCP)
- SSL/TLS analysis via testssl.sh integration
- OpenSSL s_client connections with STARTTLS support
- Plaintext protocol connections (HTTP, FTP, Telnet)
- Protocol-specific test filtering via annotations
- Automatic JSON report attachments for test results

---

## Step Definition Reference

### 1. Annotations

Control which tests run based on context:

```gherkin
@PerPort                        # Test is written for a single port
@PerService                     # Test applies across the whole service

@http, @ssh, @ftp, @smtp        # Test only applies to a specific protocol
@plaintext, @tls                # Applies to only plaintext/tls ports (http is plaintext, https is tls)

@CCC.ObjStor, @CCC.RDMS, etc.   # Test only applies to a specific CCC catalog type
@tlp-green @tlp-amber @tlp-red  # Traffic-light protocol level of the control
```

---

### 2. Pre-Configured Variables

Variables automatically available based on test context:

#### For `@PerPort` Tests

| Variable              | Example                                          | Description                          |
| --------------------- | ------------------------------------------------ | ------------------------------------ |
| `portNumber`          | `22`                                             | Port number being tested             |
| `hostName`            | `example.com`                                    | Hostname or endpoint                 |
| `protocol`            | `imap`, `pop3`, `ldap`, `postgres`               | Protocol type                        |
| `providerServiceType` | `s3`, `rds`, `Microsoft.Storage/storageAccounts` | Cloud provider-specific service type |
| `catalogType`         | `CCC.ObjStor`, `CCC.RDMS`, `CCC.VM`              | CCC catalog type                     |

#### For `@PerService` Tests

| Variable              | Example                                          | Description                          |
| --------------------- | ------------------------------------------------ | ------------------------------------ |
| `hostName`            | `my-bucket.s3.amazonaws.com`                     | Service hostname or endpoint         |
| `providerServiceType` | `s3`, `rds`, `Microsoft.Storage/storageAccounts` | Cloud provider-specific service type |
| `catalogType`         | `CCC.ObjStor`, `CCC.RDMS`, `CCC.VM`              | CCC catalog type                     |

> **Note:** `serviceType` is deprecated - use `providerServiceType` or `catalogType` instead.

---

### 3. Cloud API Initialization

```gherkin
Given a cloud api for "{Provider}" in "{api}"
```

Initializes a cloud API factory for the specified provider and stores it with the given name.

**Parameters:**

- `{Provider}`: Cloud provider name (must be: `aws`, `azure`, or `gcp`)
- `{api}`: A name to reference this API instance in subsequent steps

---

### 4. Connection Handling

Many steps create a connection object stored in `result`. Access it via `{result}` in subsequent steps.

#### Connection Properties

| Property | Description                            |
| -------- | -------------------------------------- |
| `state`  | Either `open` or `closed`              |
| `input`  | Channel to send data to the remote end |
| `output` | String containing all received data    |

#### Connection State Management

```gherkin
Then I close connection "{connection}"
And "{connection}" state is closed
And "{connection}" state is open
```

---

### 5. OpenSSL Protocol Connections

#### Basic TLS Connection

```gherkin
Given an openssl s_client request to "{portNumber}" on "{hostName}" protocol "smtp"
```

#### TLS Version-Specific Connection

```gherkin
Given an openssl s_client request using "tls1_2" to "{portNumber}" on "{hostName}" protocol "smtp"
```

**TLS Version Arguments:** `tls1_1`, `tls1_2`, `tls1_3`

#### STARTTLS Protocol Support

| Protocol | Port | Start-TLS flag       |
| -------- | ---- | -------------------- |
| SMTP     | 587  | `-starttls smtp`     |
| IMAP     | 143  | `-starttls imap`     |
| POP3     | 110  | `-starttls pop3`     |
| LDAP     | 389  | `-starttls ldap`     |
| Postgres | 5432 | `-starttls postgres` |
| XMPP     | 5222 | `-starttls xmpp`     |

#### Example: HTTPS Request

```gherkin
Given an openssl s_client request to "{portNumber}" on "{hostName}" protocol "https"
And I refer to "{result}" as "connection"
Then I transmit "{httpRequest}" to "{connection.input}"
```

Where `httpRequest` might be:

```
GET / HTTP/1.1
Host: example.com
Connection: close
```

Response available in `connection.output`.

#### Example: SMTP Request

```gherkin
Given an openssl s_client request to "{portNumber}" on "{hostName}" protocol "smtp"
And I refer to "{result}" as "connection"
Then I transmit "{smtpRequest}" to "{connection.input}"
```

---

### 6. SSL/TLS Analysis (testssl.sh)

```gherkin
Given "report" contains details of SSL Support type "X" for "{hostName}" on port "{portNumber}"
Given "report" contains details of SSL Support type "X" for "{hostName}" on port "{portNumber}" with STARTTLS
```

Uses the `testssl.sh` project to return a JSON report about SSL details on a specific port. Add `with STARTTLS` to connect to a plaintext port and upgrade to TLS.

> **Note:** The complete JSON report from testssl.sh is automatically attached to the test results and can be viewed in the HTML report.

#### Test Types

| Type                | Flag                  | Description                                                  |
| ------------------- | --------------------- | ------------------------------------------------------------ |
| `each-cipher`       | `--each-cipher`       | Checks each local cipher remotely                            |
| `cipher-per-proto`  | `--cipher-per-proto`  | Checks ciphers per protocol                                  |
| `std`               | `--std`               | Tests standard cipher categories by strength                 |
| `forward-secrecy`   | `-f`                  | Checks forward secrecy settings                              |
| `protocols`         | `-p`                  | Checks TLS/SSL protocols (including QUIC/HTTP/3, ALPN/HTTP2) |
| `grease`            | `--grease`            | Tests server bugs like GREASE and size limitations           |
| `server-defaults`   | `-S`                  | Displays server's default picks and certificate info         |
| `server-preference` | `--server-preference` | Displays server's picks: protocol+cipher                     |
| `vulnerable`        | `-U`                  | Tests for vulnerabilities (heartbleed, etc.)                 |

#### Example: Protocol Validation

```gherkin
Then "{report}" is a slice of objects which doesn't contain any of
  | id     | finding |
  | SSLv2  | offered |
  | SSLv3  | offered |
  | TLS1   | offered |
  | TLS1_1 | offered |
And "{report}" is a slice of objects with at least the following contents
  | id     | finding            |
  | TLS1_3 | offered with final |
```

#### Example: Vulnerability Check

```gherkin
Then "{report}" is a slice of objects with at least the following contents
  | id            | finding                                |
  | heartbleed    | not vulnerable, no heartbeat extension |
  | CCS           | not vulnerable                         |
  | ticketbleed   | not vulnerable                         |
  | ROBOT         | not vulnerable                         |
  | secure_renego | supported                              |
```

#### Example: Certificate Validation

```gherkin
Then "{report}" is a slice of objects with at least the following contents
  | id                    | finding |
  | cert_expirationStatus | ok      |
  | cert_chain_of_trust   | passed. |
```

#### JSON Examples

See the `examples_of_testssl/` directory for sample JSON output from each test type.

---

### 7. Plaintext Protocol Connections

```gherkin
Given a client connects to "{hostName}" with protocol "{protocol}" on port "{portNumber}"
```

Establishes a plaintext connection to verify the server is listening and responding.

#### HTTP Example

```gherkin
Given a client connects to "{hostName}" with protocol "http" on port "{portNumber}"
Then "{result}" is not nil
And "{result}" is not an error
And "{result.output}" contains "HTTP/1.1"
```

Response in `result.output`:

```
HTTP/1.1 200 OK
Server: nginx/1.18.0
Content-Type: text/html
```

> **Note:** HTTP should generally be redirected to HTTPS in production environments.

#### Telnet Example

```gherkin
Given a client connects to "{hostName}" with protocol "telnet" on port "{portNumber}"
```

Response in `result.output`:

```
Ubuntu 22.04.1 LTS
login:
```

> **Warning:** Telnet transmits credentials in plaintext and should NOT be used in production.

#### FTP Example

```gherkin
Given a client connects to "{hostName}" with protocol "ftp" on port "{portNumber}"
```

Response in `result.output`:

```
220 (vsFTPd 3.0.3)
```

---

### 8. Policy Checks

```gherkin
When I attempt policy check "{check-name}" for control "{control}" assessment requirement "{AR}" for service "{service}" on resource "{resource}" and provider "{provider}"
```

Runs a specific policy check against a cloud resource.

**Parameters:**

- `check-name`: The policy check name (e.g., `s3-bucket-region`, `s3-object-lock`)
- `control`: The CCC control ID (e.g., `CCC.Core.CN14`, `CCC.ObjStor.CN01`)
- `AR`: The assessment requirement identifier (e.g., `AR01`, `AR02`)
- `service`: The service type (e.g., `object-storage`, `iam`, `vpc`)
- `resource`: The resource name or identifier (supports variable references like `{ResourceName}`)
- `provider`: The cloud provider (e.g., `aws`, `azure`, `gcp`)

**How it works:**

1. Constructs the policy path directly: `policy/{CatalogType}/{Control}/{AR}/{check-name}/{provider}.yaml`
2. If the file is missing, returns **fail**
3. If the policy's `service_type` doesn't match the resource, returns **pass** (not applicable)
4. Executes the policy's query against the resource
5. Validates results against defined rules
6. Attaches policy result as JSON to the test report

**Result:**

- Sets `result` to `true` if policy passes or is not applicable
- Sets `result` to `false` and returns error if policy fails or file is missing

**Example:**

```gherkin
Given a cloud api for "{Provider}" in "api"
When I attempt policy check "s3-bucket-region" for control "CCC.Core.CN06" assessment requirement "AR01" for service "{ServiceType}" on resource "{ResourceName}" and provider "{Provider}"
Then "{result}" is true
```

---

## Generating Test Examples

Use the `examples_of_testssl/generate-examples.sh` script to generate sample testssl.sh output:

```bash
cd examples_of_testssl
./generate-examples.sh <hostname>:<port>
# e.g., ./generate-examples.sh robmoff.at:443
```

This generates JSON files for all test types: `<hostname>_<port>_<test-type>.json`
