# CCC.VPC.CN02 - Limit Resource Creation in Public Subnet

| Field                 | Value                                                                                                                   |
| --------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| **Family**            | Network Security                                                                                                        |
| **Control ID**        | CCC.VPC.CN02                                                                                                            |
| **Control Title**     | Limit Resource Creation in Public Subnet                                                                                |
| **Control Objective** | Restrict the creation of resources in the public subnet with direct access to the internet to minimize attack surfaces. |

## Assessment Requirements

| Assessment ID     | Requirement Text                                                                                                     |
| ----------------- | -------------------------------------------------------------------------------------------------------------------- |
| CCC.VPC.CN02.AR01 | When a resource is created in a public subnet, that resource MUST NOT be assigned an external IP address by default. |
