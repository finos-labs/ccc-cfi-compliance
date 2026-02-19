# CCC.Core.CN14 - Maintain Recent Backups

| Field                 | Value                                                                                                                 |
| --------------------- | --------------------------------------------------------------------------------------------------------------------- |
| **Family**            | Data                                                                                                                  |
| **Control ID**        | CCC.Core.CN14                                                                                                         |
| **Control Title**     | Maintain Recent Backups                                                                                               |
| **Control Objective** | Ensure that all backups used for disaster recovery are recent and subject to a retention policy that limits deletion. |

## Assessment Requirements

| Assessment ID      | Requirement Text                                                                                                                                   |
| ------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| CCC.Core.CN14.AR01 | When backups are created for disaster recovery purposes, the storage mechanism MUST NOT allow modification or deletion within 30 days of creation. |
