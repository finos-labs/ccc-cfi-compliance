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

### Executable test inputs

For the executable CN03 feature/API test:

- `REQUESTER_VPC_ID` is derived from the in-scope VPC resource UID.
- Set `PEER_VPC_ID` (or `CN03_PEER_VPC_ID`) to a disallowed peer VPC.
- Optionally set `PEER_OWNER_ID` (or `CN03_PEER_OWNER_ID`) for cross-account tests.

## How to demonstrate this test works

Use the same account/region and run these cases:

1. **AR01 compliance case (disallowed peer, expected PASS)**
   - Set `PEER_VPC_ID` to a VPC that is not allowed by your IAM/SCP guardrails.
   - Run:
     - `./testing/run-compliance-tests.sh --provider aws --region <region> --service vpc --tag 'CCC.VPC.CN03.AR01 && CN03.DISALLOWED'`
   - Expected evidence:
     - `cn03-disallowed-peering-dry-run.json` shows `DryRunAllowed=false`
     - `cn03-disallowed-summary-compact.json` includes `Mode`, `Verdict`, `ResultClass`, `Reason`, and key IDs
     - Error code/message indicates deny (`AccessDenied`, `UnauthorizedOperation`, etc.)

2. **Optional sanity case (allowed peer, not AR01 compliance proof)**
   - Set `PEER_VPC_ID` to a VPC that is explicitly allowed by guardrails.
   - Set `CN03_ALLOWED_LIST_REFERENCE` to your allowlist basis (example: `scp-0abc1234`).
   - Run:
     - `./testing/run-compliance-tests.sh --provider aws --region <region> --service vpc --tag 'CCC.VPC && CN03.ALLOWED'`
   - Expected evidence:
     - `cn03-allowed-peering-dry-run.json` shows `DryRunAllowed=true`
     - `cn03-allowed-summary-compact.json` includes `Mode`, `Verdict`, `ResultClass`, `Reason`, and key IDs
     - Error code/message is `DryRunOperation` (action would be allowed)
   - If `CN03_ALLOWED_LIST_REFERENCE` is missing, this scenario returns `ResultClass=SETUP_ERROR`.

### What to share with collaborators

- Guardrail definition source (IAM/SCP policy statement controlling peering)
- Exact env inputs used (`AWS_REGION`, `PEER_VPC_ID`, optional `PEER_OWNER_ID`)
- Test command executed
- Result artifacts from `testing/output/`:
  - `resource-<vpc>.html`
  - `resource-<vpc>.ocsf.json`
  - attached `cn03-disallowed-peering-dry-run.json` / `cn03-allowed-peering-dry-run.json`
  - attached `cn03-disallowed-summary-compact.json` / `cn03-allowed-summary-compact.json`

This gives a reproducible, auditable proof that the test logic is functioning and that guardrails drive pass/fail outcomes.
