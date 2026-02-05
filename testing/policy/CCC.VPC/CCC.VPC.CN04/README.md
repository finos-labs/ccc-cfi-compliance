# CCC.VPC.CN04 - Enforce VPC Flow Logs on VPCs

| Field                 | Value                                                                                                                                               |
| --------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Family**            | Network Security                                                                                                                                    |
| **Control ID**        | CCC.VPC.CN04                                                                                                                                        |
| **Control Title**     | Enforce VPC Flow Logs on VPCs                                                                                                                       |
| **Control Objective** | Ensure VPCs are configured with flow logs enabled to capture traffic information.                                                                   |

## Assessment Requirements

| Assessment ID     | Requirement Text                                                                                                                              |
| ----------------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| CCC.VPC.CN04.AR01 | When any network traffic goes to or from an interface in the VPC, the service MUST capture and log all relevant information.                  |

## Test approach (AWS)

This control has two parts on AWS:

1. **Configuration evidence**: flow logs are enabled for the VPC and configured to capture traffic (`TrafficType=ALL`).
2. **Behavioral evidence**: generate traffic and verify corresponding log records are delivered to the configured destination.

- **Reference query definitions (documentation only)**:
  - `testing/policy/CCC.VPC/CCC.VPC.CN04/aws/AR01/vpc-flow-logs-enabled.yaml`
  - `testing/policy/CCC.VPC/CCC.VPC.CN04/aws/AR01/vpc-flow-logs-delivery-observation.yaml`

