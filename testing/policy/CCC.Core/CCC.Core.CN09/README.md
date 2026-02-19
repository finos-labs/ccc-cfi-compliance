# CCC.Core.CN09 - Ensure Integrity of Access Logs

| Field                 | Value                                                                                                                                                   |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Family**            | Data                                                                                                                                                    |
| **Control ID**        | CCC.Core.CN09                                                                                                                                           |
| **Control Title**     | Ensure Integrity of Access Logs                                                                                                                         |
| **Control Objective** | Ensure that access logs are always recorded to an external location that cannot be manipulated from the context of the service(s) it contains logs for. |

## Assessment Requirements

| Assessment ID      | Requirement Text                                                                                                                                                   |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| CCC.Core.CN09.AR01 | When the service is operational, its logs and any child resource logs MUST NOT be accessible from the resource they record access to.                              |
| CCC.Core.CN09.AR02 | When the service is operational, disabling the logs for the service or its child resources MUST NOT be possible without also disabling the corresponding resource. |
