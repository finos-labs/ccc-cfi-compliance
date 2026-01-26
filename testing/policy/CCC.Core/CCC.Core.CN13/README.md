# CCC.Core.CN13 - Minimize Lifetime of Encryption and Authentication Certificates

| Field                 | Value                                                                                                                                                                |
| --------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Family**            | Data                                                                                                                                                                 |
| **Control ID**        | CCC.Core.CN13                                                                                                                                                        |
| **Control Title**     | Minimize Lifetime of Encryption and Authentication Certificates                                                                                                      |
| **Control Objective** | Ensure that encryption and authentication certificates have a limited lifetime to reduce the risk of compromise and ensure the use of up-to-date security practices. |

## Assessment Requirements

| Assessment ID      | Requirement Text                                                                                                                                                  |
| ------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| CCC.Core.CN13.AR01 | When a port is exposed that uses certificate-based encryption, the service MUST only use valid, unexpired certificates issued by a trusted certificate authority. |
| CCC.Core.CN13.AR02 | When a port is exposed that uses certificate-based encryption, the service MUST rotate active certificates within 180 days of issuance.                           |
| CCC.Core.CN13.AR03 | When a port is exposed that uses certificate-based encryption, the service MUST rotate active certificates within 90 days of issuance.                            |
