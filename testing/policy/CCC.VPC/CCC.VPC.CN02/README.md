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

## Test approach (AWS)

This repository evaluates "external IP assigned by default" on AWS as the subnet-level default public IP assignment behavior.

- **Evidence (control-plane)**: identify public subnets (IGW route) and verify `MapPublicIpOnLaunch=false`
- **Pass condition**: all public subnets in the VPC have `MapPublicIpOnLaunch` disabled
- **Reference query definition (documentation only)**:
  - `testing/policy/CCC.VPC/CCC.VPC.CN02/aws/AR01/public-subnet-no-default-external-ip.yaml`

### Notes and limitations

- This check focuses on *default* assignment. Workloads can still request/attach public IPs explicitly depending on IAM/policies.
- "Public subnet" classification is environment-specific; this approach uses route table analysis (0.0.0.0/0 to an Internet Gateway) as a practical definition.
- If no public subnets are found for a VPC, the check is treated as **N/A** for that VPC (and will show this explicitly in test output).
