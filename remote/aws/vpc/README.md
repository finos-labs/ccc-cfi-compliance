# AWS VPC Test Environment

This Terraform configuration provisions a minimal AWS VPC environment for `CCC.VPC` testing.

## Minimum orchestration policy (per CN)

- `CN01` (default network resources): observational only on AWS.
  - This module does not create/remove AWS account default VPC resources.
  - Declaration flag: `cn01_observation_only=true`.
- `CN02` (default public IP in public subnet): declarative subnet setting.
  - Controlled by `map_public_ip_on_launch`.
- `CN03` (restrict peering destinations): dry-run peering fixture support.
  - Creates requester VPCs in allow/disallow groups.
  - Keeps current module VPC as the peering receiver.
  - Exports an editable trial matrix JSON for API batch dry-run checks.
- `CN04` (flow logs): optional VPC flow-log declaration.
  - `cn04_enable_flow_logs`
  - `cn04_flow_log_traffic_type` (default `ALL`)

## What it creates

- Primary VPC with public subnets and IGW (always)
- CN03 requester trial VPCs for allow/disallow scenarios (optional)
- CN03 optional IAM guardrail policy declaration (optional)
- CN04 CloudWatch log group + IAM role + VPC flow log (optional)

## Cleanup tagging

Resources are tagged with:

- `Owner=<value from TF_VAR_owner>`
- `CFITargetID=<target id>`

Use these tags for targeted cleanup when runs fail mid-way.

You can extend the shared tag set:

```bash
export TF_VAR_owner="platform-security"
export TF_VAR_common_tags='{"CostCenter":"1234","DataClass":"internal"}'
```

## Local apply

```bash
cd remote/aws/vpc
terraform init
terraform apply -auto-approve -input=false
```

## Common control profiles

### CN02 pass/fail simulation

```bash
# Fail case: public subnets auto-assign public IP
export TF_VAR_map_public_ip_on_launch=true

# Pass case: public subnets do not auto-assign public IP
export TF_VAR_map_public_ip_on_launch=false
```

### CN03 requester fixture and trial matrix

```bash
# Keep the in-scope VPC as receiver; create requester trial VPCs.
export TF_VAR_cn03_profile=create
export TF_VAR_cn03_create_peers=true
export TF_VAR_cn03_create_non_allowlisted_requester=true

# Optional: attach guardrail policy to specific role and/or user names.
export TF_VAR_cn03_apply_guardrail=true
export TF_VAR_cn03_guardrail_attach_role_names='["my-test-role"]'
export TF_VAR_cn03_guardrail_attach_user_names='["my-iam-user"]'

# Alternative: update an existing attached managed policy in-place
# (no new policy creation and no new attachment needed).
export TF_VAR_cn03_existing_guardrail_policy_arn='arn:aws:iam::<account-id>:policy/CN03PeeringGuardrail'
```

When `cn03_existing_guardrail_policy_arn` is set, apply updates the existing policy by creating a new default policy version via AWS CLI (and removes one oldest non-default version only if IAM's 5-version limit is reached).

Guardrail allow-list is dynamic and tag-driven (no hardcoded VPC IDs):

- `PeerClass=allowed` VPCs are treated as allowed requester sources.
- `PeerClass=disallowed` VPCs are tracked for trial evidence.
- Guardrail condition key is `ec2:RequesterVpc` with `ArnNotEquals`.

After apply, export the matrix + env inputs:

```bash
./export-cn03-artifacts.sh
```

This writes:

- `cn03-peer-trials.json` (editable trial matrix file)
- `cn03-feature.env` (feature placeholders as shell exports)

Use with CN03 API/feature runs:

```bash
source ./cn03-feature.env
export CN03_PEER_TRIAL_MATRIX_FILE="$(pwd)/cn03-peer-trials.json"
```

### CN04 flow log configuration

```bash
export TF_VAR_cn04_enable_flow_logs=true
export TF_VAR_cn04_flow_log_traffic_type=ALL
```

Inspect declarations:

```bash
terraform output cn_control_declarations
terraform output cn03_peer_trial_matrix
terraform output cn03_feature_env
terraform output cn04_flow_log_id
terraform output cn04_flow_log_log_group_name
```

## Destroy

```bash
terraform destroy -auto-approve -input=false
```
