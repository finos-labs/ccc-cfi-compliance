# CCC.VPC.CN01 - Restrict Default Network Creation

| Field                 | Value                                                                                                                                                                                              |
| --------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Family**            | Network Security                                                                                                                                                                                   |
| **Control ID**        | CCC.VPC.CN01                                                                                                                                                                                       |
| **Control Title**     | Restrict Default Network Creation                                                                                                                                                                  |
| **Control Objective** | Restrict the automatic creation of default virtual networks and related resources during subscription initialization to avoid insecure default configurations and enforce custom network policies. |

## Assessment Requirements

| Assessment ID     | Requirement Text                                                                             |
| ----------------- | -------------------------------------------------------------------------------------------- |
| CCC.VPC.CN01.AR01 | When a subscription is created, the subscription MUST NOT contain default network resources. |
