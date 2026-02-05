# CCC.VPC.CN03 - Restrict VPC Peering to Authorized Accounts

| Field                 | Value                                                                                                                                                               |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Family**            | Network Security                                                                                                                                                    |
| **Control ID**        | CCC.VPC.CN03                                                                                                                                                        |
| **Control Title**     | Restrict VPC Peering to Authorized Accounts                                                                                                                         |
| **Control Objective** | Ensure VPC peering connections are only established with explicitly authorized destinations to limit network exposure and enforce boundary controls.                |

## Assessment Requirements

| Assessment ID     | Requirement Text                                                                                                                                           |
| ----------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| CCC.VPC.CN03.AR01 | When a VPC peering connection is requested, the service MUST prevent connections from VPCs that are not explicitly allowed.                               |

## Test approach (AWS)

This control is **behavioral/negative** on AWS: it requires attempting a peering request that should be disallowed and verifying it is prevented.

- **Evidence (behavioral)**: attempt `ec2:CreateVpcPeeringConnection` to a non-allowlisted peer VPC/account and confirm it fails
- **Pass condition**: request is denied (e.g., IAM/SCP denies the action, or guardrails prevent unauthorized peering)
- **Reference query definition (documentation only)**:
  - `testing/policy/CCC.VPC/CCC.VPC.CN03/aws/AR01/disallowed-vpc-peering-request.yaml`

### Notes and limitations

- AWS does not provide a single native “allowed peering destinations” list at the VPC level; enforcement is typically via IAM/SCP/guardrails.
- This requires a test setup with:
  - a local VPC (requester) and
  - a peer VPC ID/account that is intentionally *not* authorized.

