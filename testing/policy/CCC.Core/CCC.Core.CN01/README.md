# CCC.Core.CN01 - Encrypt Data for Transmission

| Field                 | Value                                                                                                  |
| --------------------- | ------------------------------------------------------------------------------------------------------ |
| **Family**            | Data                                                                                                   |
| **Control Title**     | Encrypt Data for Transmission                                                                          |
| **Control Objective** | Ensure that all communications are encrypted in transit to protect data integrity and confidentiality. |

## Assessment Requirements

| Assessment ID      | Requirement Text                                                                                                                                                                                                               |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| CCC.Core.CN01.AR01 | When a port is exposed for non-SSH network traffic, all traffic MUST include a TLS handshake AND be encrypted using TLS 1.3 or higher.                                                                                         |
| CCC.Core.CN01.AR02 | When a port is exposed for SSH network traffic, all traffic MUST include a SSH handshake AND be encrypted using SSHv2 or higher.                                                                                               |
| CCC.Core.CN01.AR07 | When a port is exposed, the service MUST ensure that the protocol and service officially assigned to that port number by the IANA Service Name and Transport Protocol Port Number Registry, and no other, is run on that port. |
| CCC.Core.CN01.AR08 | When a service transmits data using TLS, mutual TLS (mTLS) MUST be implemented to require both client and server certificate authentication for all connections.                                                               |
