# CCC.Core.CN11 - Protect Encryption Keys

| Field                 | Value                                                                                                                                                             |
| --------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Family**            | Data                                                                                                                                                              |
| **Control ID**        | CCC.Core.CN11                                                                                                                                                     |
| **Control Title**     | Protect Encryption Keys                                                                                                                                           |
| **Control Objective** | Ensure that encryption keys are managed securely by enforcing the use of approved algorithms, regular key rotation, and customer-managed encryption keys (CMEKs). |

## Assessment Requirements

| Assessment ID      | Requirement Text                                                                                                                                                                          |
| ------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| CCC.Core.CN11.AR01 | When encryption keys are used, the service MUST verify that all encryption keys use the latest industry-standard cryptographic algorithms.                                                |
| CCC.Core.CN11.AR02 | When encryption keys are used, the service MUST rotate active keys within 180 days of issuance.                                                                                           |
| CCC.Core.CN11.AR03 | When encrypting data, the service MUST verify that customer-managed encryption keys (CMEKs) are used.                                                                                     |
| CCC.Core.CN11.AR04 | When encryption keys are accessed, the service MUST verify that access to encryption keys is restricted to authorized personnel and services, following the principle of least privilege. |
| CCC.Core.CN11.AR05 | When encryption keys are used, the service MUST rotate active keys within 365 days of issuance.                                                                                           |
| CCC.Core.CN11.AR06 | When encryption keys are used, the service MUST rotate active keys within 90 days of issuance.                                                                                            |
