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

## Test approach (AWS)

This repository evaluates the AWS interpretation of "default network resources" as the presence of an AWS **default VPC** in the configured region.

- **Evidence (control-plane)**: `ec2:DescribeVpcs` filtered by `is-default=true`
- **Pass condition**: zero default VPCs are returned for the region under test
- **Implementation**:
  - Evidence collection (AWS SDK): `testing/api/vpc/aws-vpc.go`
  - Executable test (Godog): `testing/features/CCC.VPC/CCC-VPC-CN01-AR01.feature`
  - Reference query definitions (documentation only):
    - `testing/policy/CCC.VPC/CCC.VPC.CN01/aws/AR01/default-vpc-absence.yaml`
    - `testing/policy/CCC.VPC/CCC.VPC.CN01/aws/AR01/default-vpc-default-subnets-absence.yaml`
    - `testing/policy/CCC.VPC/CCC.VPC.CN01/aws/AR01/default-vpc-internet-gateway-absence.yaml`
    - `testing/policy/CCC.VPC/CCC.VPC.CN01/aws/AR01/default-vpc-main-route-table-absence.yaml`
    - `testing/policy/CCC.VPC/CCC.VPC.CN01/aws/AR01/default-vpc-default-network-acl-absence.yaml`

### Notes and limitations

- AWS accounts commonly include a default VPC per region unless it has been removed. For this control to pass, the test account/region(s) must be prepared accordingly.
- This is a **current-state** check. "When a subscription is created" is not directly observable after the fact in AWS; the practical enforcement is ensuring the default VPC does not exist in each in-scope region.
- These checks are region-scoped; repeat validation for every region that is in scope for your trust perimeter.
