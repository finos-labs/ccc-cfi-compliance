# CCC.Core.CN06 - Restrict Deployments to Trust Perimeter

| Field                 | Value                                                                                                                                                           |
| --------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Family**            | Data                                                                                                                                                            |
| **Control ID**        | CCC.Core.CN06                                                                                                                                                   |
| **Control Title**     | Restrict Deployments to Trust Perimeter                                                                                                                         |
| **Control Objective** | Ensure that the service and its child resources are only deployed on infrastructure in locations that are explicitly included within a defined trust perimeter. |

## Assessment Requirements

| Assessment ID      | Requirement Text                                                                                                                                                       |
| ------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| CCC.Core.CN06.AR01 | When the service is running, its region and availability zone MUST be included in a list of explicitly trusted or approved locations within the trust perimeter.       |
| CCC.Core.CN06.AR02 | When a child resource is deployed, its region and availability zone MUST be included in a list of explicitly trusted or approved locations within the trust perimeter. |
