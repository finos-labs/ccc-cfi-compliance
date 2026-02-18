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
  - a local in-scope VPC (receiver) and
  - requester VPC IDs that are intentionally split across allowed and disallowed sets.

### Executable test inputs

For the executable CN03 feature/API test:

- `ReceiverVpcId` is derived from the in-scope VPC resource UID.
- Set requester VPC list inputs:
  - `CN03_ALLOWED_REQUESTER_VPC_ID_1..N`
  - `CN03_DISALLOWED_REQUESTER_VPC_ID_1..N`
  - optional CSV: `CN03_ALLOWED_REQUESTER_VPC_IDS`
- Optional file-driven batch input:
  - `CN03_PEER_TRIAL_MATRIX_FILE` (JSON with receiver + requester allow/disallow lists)
- Optional cross-account input:
  - `PEER_OWNER_ID` (or `CN03_PEER_OWNER_ID`)

## How to demonstrate this test works

Use the same account/region and run these cases:

1. **AR01 compliance case (disallowed requester, expected PASS)**
   - Set `CN03_DISALLOWED_REQUESTER_VPC_ID_1` to a requester VPC that is not allowed by your IAM/SCP guardrails.
   - Run:
     - `./testing/run-compliance-tests.sh --provider aws --region <region> --service vpc --tag 'CCC.VPC.CN03.AR01 && MAIN'`
   - Expected evidence:
     - Dry-run evidence returns `DryRunAllowed=false`
     - `ExitCode` is non-zero
     - Error code/message indicates deny (`AccessDenied`, `UnauthorizedOperation`, etc.)

2. **Optional sanity case (allowed requester, not AR01 compliance proof)**
   - Set `CN03_ALLOWED_REQUESTER_VPC_ID_1` to a requester VPC that is explicitly allowed by guardrails.
   - Run:
     - `./testing/run-compliance-tests.sh --provider aws --region <region> --service vpc --tag 'CCC.VPC.CN03.AR01 && SANITY && OPT_IN'`
   - Expected evidence:
     - Dry-run evidence returns `DryRunAllowed=true`
     - Error code/message is `DryRunOperation` (action would be allowed)

3. **Optional batch case (file input, all items checked)**
   - Export trial matrix from IaC:
     - `terraform output -json cn03_peer_trial_matrix > cn03-peer-trials.json`
   - Set:
     - `export CN03_PEER_TRIAL_MATRIX_FILE=$(pwd)/cn03-peer-trials.json`
   - Run:
     - `./testing/run-compliance-tests.sh --provider aws --region <region> --service vpc --tag 'CCC.VPC.CN03.AR01 && OPT_IN'`
   - Expected evidence:
     - `TotalTrials` is greater than `0`
     - `UnexpectedCount` is `0`

### What to share with collaborators

- Guardrail definition source (IAM/SCP policy statement controlling peering)
- Exact env inputs used (`AWS_REGION`, requester IDs/lists, optional `PEER_OWNER_ID`, optional `CN03_PEER_TRIAL_MATRIX_FILE`)
- Test command executed
- Result artifacts from `testing/output/`:
  - `resource-<vpc>.html`
  - `resource-<vpc>.ocsf.json`
  - attached dry-run evidence from CN03 scenarios

This gives a reproducible, auditable proof that the test logic is functioning and that guardrails drive pass/fail outcomes.
